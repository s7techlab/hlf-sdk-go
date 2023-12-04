package docker

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/mount"
	docker "github.com/fsouza/go-dockerclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
)

const (
	containerPkgOutputPath = `/tmp/chaincode.pkg`

	fabricImage          = "hyperledger/fabric-tools"
	DefaultFabricV1Image = fabricImage + `:1.4`
	DefaultFabricV2Image = fabricImage + `:2.5`
)

var ErrUnsupportedFabricVersion = errors.New(`unsupported fabric version`)

type (
	Packer struct {
		logger        *zap.Logger
		fabricV1Image string
		fabricV2Image string
	}

	PackerOpt func(*Packer)
)

func UseFabricV1Image(image string) PackerOpt {
	return func(p *Packer) {
		p.fabricV1Image = image
	}
}

func UseFabricV2Image(image string) PackerOpt {
	return func(p *Packer) {
		p.fabricV2Image = image
	}
}

func New(logger *zap.Logger, opts ...PackerOpt) *Packer {
	p := &Packer{
		logger: logger,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.fabricV1Image == `` {
		p.fabricV1Image = DefaultFabricV1Image
	}

	if p.fabricV2Image == `` {
		p.fabricV2Image = DefaultFabricV2Image
	}

	return p
}

func (p *Packer) imageLifecycle(version ccpackage.FabricVersion) (image string, lifecycle bool, err error) {
	switch version {
	case ccpackage.FabricVersion_FABRIC_V1:
		image = p.fabricV1Image
	case ccpackage.FabricVersion_FABRIC_V2:
		image = p.fabricV1Image
	case ccpackage.FabricVersion_FABRIC_V2_LIFECYCLE:
		image = p.fabricV2Image
		lifecycle = true
	case ccpackage.FabricVersion_FABRIC_VERSION_UNSPECIFIED:
		return ``, false, fmt.Errorf("version=%s: %w", version, ErrUnsupportedFabricVersion)
	}
	return image, lifecycle, nil
}

func (p *Packer) PackFromTar(ctx context.Context, spec *ccpackage.PackageSpec, tar []byte) (pkg *ccpackage.Package, err error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	image, lifecycle, err := p.imageLifecycle(spec.Id.FabricVersion)
	if err != nil {
		return nil, err
	}

	container, err := CreateContainer(ctx, image, p.logger)
	if err != nil {
		return nil, err
	}
	defer func() {
		removeErr := container.Remove(ctx)
		if removeErr != nil {
			err = removeErr
		}
	}()

	sourcePath := filepath.Join(containerSrcPath, spec.ChaincodePath)

	// Create chaincode path in container
	if err = container.Exec(ctx, []string{`mkdir`, `-p`, sourcePath}); err != nil {
		return nil, fmt.Errorf("create chaincode repo path=%s: %w", sourcePath, err)
	}
	// Set permissions on package path for upload
	if err = container.Exec(ctx, []string{`chmod`, `-R`, `777`, sourcePath}); err != nil {
		return nil, fmt.Errorf("set permissions on path=%s: %w", sourcePath, err)
	}

	if err = container.UploadTar(ctx, tar, sourcePath); err != nil {
		return nil, err
	}

	binaryPath := fmt.Sprintf("%s/%s", spec.ChaincodePath, spec.BinaryPath)

	// path should point to directory with go file  with main func
	if err = container.Exec(ctx, []string{"ls", binaryPath}); err != nil {
		return nil, fmt.Errorf(`check chaincode binary path=%s: %w`, binaryPath, err)
	}

	p.logger.Info(`peer chaincode package`,
		zap.String(`name`, spec.Id.Name),
		zap.String(`path`, binaryPath),
		zap.String(`version`, spec.Id.Version),
		zap.Stringer(`fabric_version`, spec.Id.FabricVersion))

	if err = container.Exec(ctx, packageCmd(lifecycle, binaryPath, spec.Id.Name, spec.Id.Version)); err != nil {
		return nil, fmt.Errorf("create chaincode package path=%s, fabric_version=%s: %w",
			binaryPath, spec.Id.FabricVersion, err)
	}

	data, downloadErr := container.DownloadPath(ctx, containerPkgOutputPath)
	if downloadErr != nil {
		return nil, fmt.Errorf(`download package=%s: %w`, containerPkgOutputPath, err)
	}

	return &ccpackage.Package{
		Id:        spec.Id,
		Size:      int64(len(data)),
		CreatedAt: timestamppb.Now(),
		Data:      data,
	}, nil
}

func (p *Packer) PackFromFiles(ctx context.Context, spec *ccpackage.PackageSpec, path string) (pkg *ccpackage.Package, err error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	image, lifecycle, imageErr := p.imageLifecycle(spec.Id.FabricVersion)
	if imageErr != nil {
		return nil, imageErr
	}
	sourcePath := filepath.Join(containerSrcPath, spec.ChaincodePath)

	p.logger.Info(`created container with mounted source code`,
		zap.String(`path`, path),
		zap.String(`target`, sourcePath))

	container, err := CreateContainer(ctx, image, p.logger, func(opts *docker.CreateContainerOptions) {
		opts.HostConfig = &docker.HostConfig{
			Mounts: []docker.HostMount{
				{
					Target:   sourcePath,
					Source:   path,
					Type:     string(mount.TypeBind),
					ReadOnly: true,
				},
			},
		}
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		err = container.Remove(ctx)
	}()

	fmt.Println(lifecycle)

	return nil, nil
}

func packageCmd(lifecycle bool, path, name, version string) []string {
	if lifecycle {
		return []string{
			`peer`, `lifecycle`, `chaincode`, `package`,
			`--path`, path,
			`--lang`, `golang`,
			`--label`, fmt.Sprintf("%s_%s", name, version),
			containerPkgOutputPath,
		}
	}

	return []string{
		`peer`, `chaincode`, `package`,
		`-n`, name,
		`-p`, path,
		`-v`, version,
		containerPkgOutputPath,
	}
}
