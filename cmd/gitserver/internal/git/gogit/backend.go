package cli

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewGoGitBackend(logger log.Logger, dir common.GitDir, repoName api.RepoName) (GitBackend, error) {
	r, err := git.PlainOpen(dir.Path())
	if err != nil {
		return nil, errors.Wrap(err, "failed to open repository")
	}

	return &goGitBackend{
		repo:     r,
		logger:   logger,
		dir:      dir,
		repoName: repoName,
	}, nil
}

type goGitBackend struct {
	logger   log.Logger
	rcf      *wrexec.RecordingCommandFactory
	dir      common.GitDir
	repoName api.RepoName
	repo     *git.Repository
}

func (g *goGitBackend) MergeBase(ctx context.Context, base, head api.CommitID) (api.CommitID, error) {
	baseRef, err := g.repo.Reference(plumbing.ReferenceName(base), true)
	baseRef.Hash()

	baseC, err := g.repo.CommitObject(baseRef.Hash())
	baseC.Parents()
}
