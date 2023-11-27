package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MergeBase finds the merge base commit for the given base and head SHAs.
// Both baseSHA and headSHA are expected to be valid SHAs and are not validated
// for safety.
func MergeBase(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, baseSHA, headSHA string) (api.CommitID, error) {
	cmd := exec.Command("git", "merge-base", "--", string(baseSHA), string(headSHA))
	dir.Set(cmd)

	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), repo, cmd)
	out, err := wrappedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", wrappedCmd.Args, out))
	}

	return api.CommitID(bytes.TrimSpace(out)), nil
}
