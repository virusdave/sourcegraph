package git

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// GitBackend is the interface through which operations on a git repository can
// be performed. It encapsulates the underlying git implementation and allows
// us to test out alternative backends.
// A GitBackend is expected to be scoped to a specific repository directory at
// initialization time, ie. it should not be shared across various repositories.
type GitBackend interface {
	// Config returns a backend for interacting with git configuration.
	Config() GitConfigBackend
	GetObject(ctx context.Context, objectName string) (*gitdomain.GitObject, error)
	// MergeBase finds the merge base commit for the given base and head SHAs.
	// Both baseSHA and headSHA are expected to be valid SHAs and are not validated
	// for safety.
	MergeBase(baseSHA, headSHA api.CommitID) (api.CommitID, error)
	// SetRepositoryType sets the type of the repository.
	SetRepositoryType(typ string) error
	// GetRepositoryType returns the type of the repository.
	GetRepositoryType() (string, error)
	LatestCommitTimestamp() time.Time
	ComputeRefHash() ([]byte, error)
	RemoveBadRefs(ctx context.Context) error
	SetGitAttributes() error
	EnsureHEAD() error
	CleanTmpPackFiles()
}

// GitConfigBackend provides methods for interacting with git configuration.
type GitConfigBackend interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Unset(key string) error
}
