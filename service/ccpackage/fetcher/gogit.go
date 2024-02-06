package fetcher

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.uber.org/zap"
)

const defaultMode = 0755

type (
	Git struct {
		s      storage.Storer
		auth   transport.AuthMethod
		Logger *zap.Logger
	}

	GitOpt func(*Git)
)

func WithLogger(l *zap.Logger) GitOpt {
	return func(g *Git) {
		g.Logger = l
	}
}

func GitBasicAuth(username, password string) GitOpt {
	return func(g *Git) {
		g.auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

func GitTokenAuth(token string) GitOpt {
	return func(g *Git) {
		g.auth = &http.TokenAuth{Token: token}
	}
}

func NewGit(opts ...GitOpt) *Git {
	gitFetcher := &Git{s: memory.NewStorage()}
	for _, o := range opts {
		o(gitFetcher)
	}

	if gitFetcher.Logger == nil {
		gitFetcher.Logger = zap.NewNop()
	}
	return gitFetcher
}

func (g *Git) prepareUrl(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return ``, fmt.Errorf("failed to parse url: %w", err)
	}

	if u.Scheme == `` {
		u.Scheme = `https`
	}

	if u.User != nil {
		pass, _ := u.User.Password()
		GitBasicAuth(u.User.Username(), pass)(g)
		u.User = nil
	} else if u.Fragment != `` {
		// consider as token
		GitTokenAuth(u.Fragment)(g)
		u.Fragment = ``
	}
	return u.String(), nil
}

func (g *Git) Fetch(ctx context.Context, repo, version string) ([]byte, error) {
	repoUrl, err := g.prepareUrl(repo)
	if err != nil {
		return nil, fmt.Errorf("prepare repo url: %w", err)
	}

	fs := memfs.New()

	var refName plumbing.ReferenceName

	if version != `` {
		refName = plumbing.NewHashReference(plumbing.Master, plumbing.NewHash(version)).Target()
	}

	fields := []zap.Field{
		zap.String(`url`, repoUrl),
		zap.String(`version`, version),
		zap.String(`auth`, fmt.Sprintf(`%T`, g.auth))}
	g.Logger.Info(`cloning git repo...`, fields...)

	if _, err = git.CloneContext(ctx, g.s, fs, &git.CloneOptions{
		URL:           repoUrl,
		Auth:          g.auth,
		SingleBranch:  true,
		Progress:      nil,
		ReferenceName: refName,
	}); err != nil {
		return nil, fmt.Errorf("clone repository=%s, version=%s: %w", repoUrl, version, err)
	}

	g.Logger.Debug(`git repo cloned`, fields...)

	bf := new(bytes.Buffer)
	tw := tar.NewWriter(bf)

	defer func() {
		twErr := tw.Close()
		if err == nil && twErr != nil {
			err = fmt.Errorf("close tar writer: %w", err)
		}
	}()

	if err = AddFileToTar(tw, fs.Root(), fs); err != nil {
		return nil, fmt.Errorf("add file to archive: %w", err)
	}

	return bf.Bytes(), nil
}
