package cli

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewGitCLIBackend(logger log.Logger, rcf *wrexec.RecordingCommandFactory, dir common.GitDir, repoName api.RepoName) GitBackend {
	return &gitCLIBackend{
		logger:   logger,
		rcf:      rcf,
		dir:      dir,
		repoName: repoName,
	}
}

type gitCLIBackend struct {
	logger   log.Logger
	rcf      *wrexec.RecordingCommandFactory
	dir      common.GitDir
	repoName api.RepoName
}

func (g *gitCLIBackend) MergeBase(ctx context.Context, base, head api.CommitID) (api.CommitID, error) {
	cmd := g.gitCommand(ctx, "merge-base", "--", string(base), string(head))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandFailedError(err, cmd, out)
	}

	return api.CommitID(bytes.TrimSpace(out)), nil
}

func commandFailedError(err error, cmd wrexec.Cmder, out []byte) error {
	return errors.Wrapf(err, "git command %v failed (output: %q)", cmd.Unwrap().Args, out)
}

func (g *gitCLIBackend) gitCommand(ctx context.Context, args ...string) wrexec.Cmder {
	cmd := exec.Command("git", args...)
	g.dir.Set(cmd)

	return g.rcf.WrapWithRepoName(ctx, g.logger, g.repoName, cmd)
}
