package fetcher

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

const (
	FileProtocolPrefix       = `file://`
	GitProtocolPrefix        = `https://`
	LocalMountProtocolPrefix = `local://`
)

var (
	ErrUnknownProtocol = errors.New(`unknown protocol`)
)

type Fetcher interface {
	// Fetch code by presented path and returns tar representation
	Fetch(ctx context.Context, repo, version string) ([]byte, error)
}

// Fetch repo from file path , git repo
// file - file://path/to/repo
// git repo with basic auth and token - https://{user}:{pass}@github.com/s7techlab/{repo}.git
// git repo with token - https://{token}@github.com/s7techlab/{repo}.git
func Fetch(ctx context.Context, repo, version string, logger *zap.Logger) ([]byte, error) {
	f, err := Create(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("fetcher get: %w", err)
	}

	return f.Fetch(ctx, repo, version)
}

// Create returns Fetcher instance by repo scheme
// Scheme file:// returns FileFetcher
// Scheme https:// returns GitFetcher
// Scheme local:// returns Local mount fetcher
// git repo with basic auth and token - https://{user}:{pass}@github.com/s7techlab/{repo}.git
// git repo with token - https://{token}@github.com/s7techlab/{repo}.git
// Returns non-fetcher for unrecognized scheme
func Create(repo string, l *zap.Logger) (Fetcher, error) {
	switch {
	case strings.HasPrefix(repo, FileProtocolPrefix):
		return NewFile(l), nil
	case strings.HasPrefix(repo, GitProtocolPrefix):
		var opts []GitOpt
		// Set auth options for repository if provided
		u, err := url.Parse(repo)
		if err != nil {
			return nil, fmt.Errorf("parse repo url: %w", err)
		}
		if p, set := u.User.Password(); set {
			opts = append(opts, GitBasicAuth(u.User.Username(), p))
		} else if u.User.Username() != `` {
			opts = append(opts, GitTokenAuth(u.User.Username()))
		}
		opts = append(opts, WithLogger(l))
		return NewGit(opts...), nil
	case strings.HasPrefix(repo, LocalMountProtocolPrefix):
		return NewNope(), nil

	default:
		return nil, ErrUnknownProtocol
	}
}
