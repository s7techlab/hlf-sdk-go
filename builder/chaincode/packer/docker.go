package packer

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	fsPath "path"
	"strings"

	"github.com/docker/docker/api/types/mount"
	docker "github.com/fsouza/go-dockerclient"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
)

const (
	imageGoPath = `/go`
	imageGoSrc  = `src`

	tempPkgPath = `/tmp/chaincode.pkg`

	fabricImage = "hyperledger/fabric-tools"

	fabricV1Image = fabricImage + `:1.4`
	fabricV2Image = fabricImage + `:2.3`
)

type Params struct {
	Image string
	// UseLifecycle capability HLF 2.+
	UseLifecycle bool
}

type Docker struct {
	cli           *docker.Client
	fabricVersion chaincode.FabricVersion
	image         string
	lifecycle     bool
	log           *zap.Logger
}

func New(fabricVersion chaincode.FabricVersion, log *zap.Logger) (*Docker, error) {

	var (
		image     string
		lifecycle bool
	)

	switch fabricVersion {
	case chaincode.FabricV1:
		image = fabricV1Image
	case chaincode.FabricV2:
		image = fabricV2Image
	case chaincode.FabricV2Lifecycle:
		image = fabricV2Image
		lifecycle = true
	default:
		return nil, fmt.Errorf("unknown fabric version: %s", fabricVersion)
	}

	// Create new Docker client based on environment variables
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Ping Docker daemon for checking connectivity
	if err = client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Docker: %w", err)
	}

	return &Docker{
		cli:           client,
		fabricVersion: fabricVersion,
		image:         image,
		lifecycle:     lifecycle,
		log:           log,
	}, nil
}

// checkImage checks image existences, otherwise pulls it
func (d Docker) checkImage(ctx context.Context) error {
	_, err := d.cli.InspectImage(d.image)
	switch err {
	case docker.ErrNoSuchImage:
		d.log.Info(`pulling image`, zap.String(`image`, d.image))
		return d.cli.PullImage(docker.PullImageOptions{
			Repository: d.image,
			Context:    ctx,
		}, docker.AuthConfiguration{})
	default:
		return err
	}
}

func (d Docker) exec(ctx context.Context, contId string, cmd []string) error {
	// Create exec with peer command
	exec, err := d.cli.CreateExec(docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          cmd,
		Container:    contId,
		Context:      ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	// Start exec

	errorBuf := new(bytes.Buffer)
	outBuf := new(bytes.Buffer)

	if err = d.cli.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  errorBuf,
		OutputStream: outBuf,
		Context:      ctx,
	}); err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Inspect exec for results
			ex, err := d.cli.InspectExec(exec.ID)
			if err != nil {
				return fmt.Errorf("failed to inspect exec: %w", err)
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

func (d Docker) Pack(ctx context.Context, pkg *chaincode.Package) (err error) {
	// Check Docker image
	if err = d.checkImage(ctx); err != nil {
		err = fmt.Errorf("failed to check image: %w", err)
		return
	}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      d.image,
			WorkingDir: fsPath.Join(imageGoPath, imageGoSrc),
			Entrypoint: []string{`tail`, `-f`, `/dev/null`},
			Env: []string{
				fmt.Sprintf("GOPATH=%s", imageGoPath),
			},
		},
		Context: ctx,
	}

	pkgPath := fsPath.Join(imageGoPath, imageGoSrc, pkg.ChaincodePath)

	d.log.Info(`create docker container`, zap.String(`image`, d.image))
	if pkg.Source == nil {
		d.log.Info(`using volume without source code, mount source code`,
			zap.String(`source`, pkg.Repository),
			zap.String(`target`, pkgPath))
		opts.HostConfig = &docker.HostConfig{
			Mounts: []docker.HostMount{
				{
					Target:   pkgPath,
					Source:   pkg.Repository,
					Type:     string(mount.TypeBind),
					ReadOnly: true,
				},
			},
		}
	}
	// Create Docker container with infinite endpoint
	cont, err := d.cli.CreateContainer(opts)
	if err != nil {
		err = fmt.Errorf("failed to create container: %w", err)
		return
	}
	// Remove container anyway
	defer func() {
		// Remove container in all cases (cleanup)
		remErr := d.cli.RemoveContainer(docker.RemoveContainerOptions{
			ID:            cont.ID,
			RemoveVolumes: false,
			Force:         true,
			Context:       ctx,
		})
		if remErr != nil && err == nil {
			err = fmt.Errorf("failed to remove container: %w", remErr)
		}
	}()

	d.log.Debug(`start docker container`, zap.String(`id`, cont.ID))

	// Start Docker container
	if err = d.cli.StartContainer(cont.ID, nil); err != nil {
		err = fmt.Errorf("failed to start container: %w", err)
		return
	}

	d.log.Debug(`creating chaincode path`, zap.String(`path`, pkgPath))
	// Create chaincode path in container
	if err = d.exec(ctx, cont.ID, []string{`mkdir`, `-p`, pkgPath}); err != nil {
		err = fmt.Errorf("create chaincode path: %w", err)
		return
	}
	// Set permissions on package path for upload
	if pkg.Source != nil {
		if err = d.exec(ctx, cont.ID, []string{`chmod`, `-R`, `777`, pkgPath}); err != nil {
			err = fmt.Errorf("set permissions on path: %w", err)
			return
		}
	}

	if pkg.Source != nil {
		d.log.Debug(`upload tar to chaincode path`, zap.Int(`size`, len(pkg.Source)))
		// Upload tar archive with chaincode source code to container
		if err = d.cli.UploadToContainer(cont.ID, docker.UploadToContainerOptions{
			InputStream:          bytes.NewReader(pkg.Source),
			Path:                 pkgPath,
			NoOverwriteDirNonDir: true,
			Context:              ctx,
		}); err != nil {
			err = fmt.Errorf("failed to upload code to container: %w", err)
			return
		}
	} else {
		d.log.Debug(`ignoring upload source to container`)
	}

	path := fmt.Sprintf("%s/%s", pkg.ChaincodePath, pkg.BinaryPath)

	// path should point to directory with go file  with main func
	if err = d.exec(ctx, cont.ID, []string{"ls", path}); err != nil {
		return fmt.Errorf(`check binary path=%s: %w`, path, err)
	}

	d.log.Info(`peer chaincode package`,
		zap.String(`name`, pkg.Name),
		zap.String(`path`, path),
		zap.String(`version`, pkg.Version),
		zap.String(`fabric_version`, string(d.fabricVersion)))

	var args []string
	// Package chaincode inside container
	if d.lifecycle {
		args = []string{
			`peer`, `lifecycle`, `chaincode`, `package`,
			`--path`, path,
			`--lang`, `golang`,
			`--label`, fmt.Sprintf("%s_%s", pkg.Name, pkg.Version),
			tempPkgPath,
		}
	} else {
		args = []string{
			`peer`, `chaincode`, `package`,
			`-n`, pkg.Name,
			`-p`, path,
			`-v`, pkg.Version,
			tempPkgPath,
		}
	}

	d.log.Info(`peer chaincode packaging command: ` + strings.Join(args, ` `))

	if err = d.exec(ctx, cont.ID, args); err != nil {
		return fmt.Errorf("create chaincode package path=%s, fabric_version=%s: %w",
			path, d.fabricVersion, err)
	}

	// Download artifacts from container to buffer
	outBf := new(bytes.Buffer)
	if err = d.cli.DownloadFromContainer(cont.ID, docker.DownloadFromContainerOptions{
		OutputStream:      outBf,
		Path:              tempPkgPath,
		InactivityTimeout: 0,
		Context:           ctx,
	}); err != nil {
		err = fmt.Errorf("download chaincode package from container: %w", err)
	}

	pkgTar, err := unTarFirstFile(outBf)
	if err != nil {
		return
	}

	pkg.Data = pkgTar.Bytes()
	return nil
}

func unTarFirstFile(r io.Reader) (*bytes.Buffer, error) {
	tr := tar.NewReader(r)
	file := new(bytes.Buffer)

	header, err := tr.Next()

	switch {
	case header == nil:
		return nil, nil

	case err != nil:
		return nil, err
	}

	switch header.Typeflag {
	case tar.TypeReg:
		if _, err := io.CopyN(file, tr, header.Size); err != nil {
			return nil, err
		}
		return file, nil

	default:
		return nil, errors.New(`file not found in tar`)
	}

}
