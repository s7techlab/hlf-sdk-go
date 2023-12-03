package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
	"go.uber.org/zap"
)

const (
	ContainerGoPath  = `/go`
	containerSrcPath = ContainerGoPath + `/src`
)

type (
	Container struct {
		docker   *docker.Client
		instance *docker.Container // current container

		image  string
		logger *zap.Logger
	}

	CreateContainerOpt func(*docker.CreateContainerOptions)
)

var ErrContainerNotCreated = errors.New(`container not created`)

func CreateContainer(ctx context.Context, image string, logger *zap.Logger, opts ...CreateContainerOpt) (
	container *Container, err error) {
	// Create new Docker client based on environment variables
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	// Ping Docker daemon for checking connectivity
	if err = client.Ping(); err != nil {
		return nil, fmt.Errorf("ping docker: %w", err)
	}

	c := &Container{
		docker:   client,
		instance: nil,
		image:    image,
		logger:   logger,
	}

	logger.Info(`create container`, zap.String(`image`, image))
	if err := c.checkImage(ctx); err != nil {
		return nil, err
	}

	containerOpts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      image,
			WorkingDir: containerSrcPath,
			Entrypoint: []string{`tail`, `-f`, `/dev/null`},
			Env: []string{
				fmt.Sprintf("GOPATH=%s", ContainerGoPath),
			},
		},
		Context: ctx,
	}

	for _, opt := range opts {
		opt(&containerOpts)
	}
	// Create Docker container with infinite endpoint
	c.instance, err = c.docker.CreateContainer(containerOpts)
	if err != nil {
		return nil, fmt.Errorf("create docker container: %w", err)
	}

	c.logger.Debug(`container created`, zap.String(`id`, c.instance.ID))
	c.logger.Debug(`start docker container`, zap.String(`id`, c.instance.ID))

	// Start Docker container
	if err = c.docker.StartContainer(c.instance.ID, nil); err != nil {
		return nil, fmt.Errorf("start docker container: %w", err)
	}

	return c, nil
}

// checkImage checks image existences, otherwise pulls it
func (c *Container) checkImage(ctx context.Context) error {
	c.logger.Debug(`inspect docker image`, zap.String(`image`, c.image))

	_, err := c.docker.InspectImage(c.image)
	switch err {
	case docker.ErrNoSuchImage:
		c.logger.Info(`pulling image`, zap.String(`image`, c.image))

		return c.docker.PullImage(docker.PullImageOptions{
			Repository: c.image,
			Context:    ctx,
		}, docker.AuthConfiguration{})
	default:
		return err
	}
}

func (c *Container) Exec(ctx context.Context, cmd []string) error {
	c.logger.Info(`exec cmd`, zap.Strings(`cmd`, cmd))

	// Create exec with peer command
	exec, err := c.docker.CreateExec(docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          cmd,
		Container:    c.instance.ID,
		Context:      ctx,
	})
	if err != nil {
		return fmt.Errorf("create docker exec: %w", err)
	}

	// Start exec
	errorBuf := new(bytes.Buffer)
	outBuf := new(bytes.Buffer)

	if err = c.docker.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  errorBuf,
		OutputStream: outBuf,
		Context:      ctx,
	}); err != nil {
		return fmt.Errorf("start docker exec: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Inspect exec for results
			ex, err := c.docker.InspectExec(exec.ID)
			if err != nil {
				return fmt.Errorf("inspect docker exec: %w", err)
			}
			// Continue if program is still running
			if ex.Running {
				continue
			} else {
				// Expect error code is successful (zero)
				if ex.ExitCode != 0 {
					return fmt.Errorf("exec exit code is %d with output: %s", ex.ExitCode, errorBuf.String())
				}
				return nil
			}
		}
	}
}

func (c *Container) Remove(ctx context.Context) error {
	c.logger.Info(`remove container`, zap.String(`id`, c.instance.ID))
	if c.instance == nil {
		return ErrContainerNotCreated
	}

	// Remove container in all cases (cleanup)
	err := c.docker.RemoveContainer(docker.RemoveContainerOptions{
		ID:            c.instance.ID,
		RemoveVolumes: true,
		Force:         true,
		Context:       ctx,
	})
	if err != nil {
		return fmt.Errorf("remove docker container with id=%s: %w", c.instance.ID, err)
	}

	c.instance = nil
	return nil
}

func (c *Container) UploadTar(ctx context.Context, tar []byte, path string) error {
	c.logger.Debug(`upload tar to container`, zap.String(`path`, path), zap.Int(`size`, len(tar)))

	if c.instance == nil {
		return ErrContainerNotCreated
	}

	// Upload tar archive with chaincode source code to container
	if err := c.docker.UploadToContainer(c.instance.ID, docker.UploadToContainerOptions{
		InputStream:          bytes.NewReader(tar),
		Path:                 path,
		NoOverwriteDirNonDir: true,
		Context:              ctx,
	}); err != nil {
		return fmt.Errorf("upload tar to container: %w", err)
	}

	return nil
}

func (c *Container) DownloadPath(ctx context.Context, path string) ([]byte, error) {
	if c.instance == nil {
		return nil, ErrContainerNotCreated
	}

	// Download artifacts from container to buffer
	outBf := new(bytes.Buffer)
	if err := c.docker.DownloadFromContainer(c.instance.ID, docker.DownloadFromContainerOptions{
		OutputStream:      outBf,
		Path:              path,
		InactivityTimeout: 0,
		Context:           ctx,
	}); err != nil {
		err = fmt.Errorf("download path=%s from container: %w", path, err)
	}

	pkgTar, err := UnTarFirstFile(outBf)
	if err != nil {
		return nil, err
	}

	return pkgTar.Bytes(), nil
}
