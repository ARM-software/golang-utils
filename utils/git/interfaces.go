package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/ARM-software/golang-utils/utils/config"
)

// ILimits defines general GitLimits for actions performed during a git clone
type ILimits interface {
	config.IServiceConfiguration
	// Apply states whether the limit should be applied
	Apply() bool
	// GetMaxFileSize returns the maximum size in byte a file can have on a file system
	GetMaxFileSize() int64
	// GetMaxTotalSize returns the maximum size in byte a location can have on a file system (whether it is a file or a folder)
	GetMaxTotalSize() int64
	// GetMaxFileCount returns the maximum number of files allowed in a reposittory
	GetMaxFileCount() int64
	// GetMaxTreeDepth returns the maximum tree depth for a repository
	GetMaxTreeDepth() int64
	// GetMaxEntries returns the maximum total entries allowed in the it repo
	GetMaxEntries() int64
	// GetMaxTrueSize returns the maximum true size of the repo based on blobs
	GetMaxTrueSize() int64
}

type IGitActionConfig interface {
	config.IServiceConfiguration
	GetUrl() string
	GetAuth() transport.AuthMethod
	GetDepth() int
	GetReference() string
	GetRecursiveSubModules() git.SubmoduleRescursivity
	GetTags() git.TagMode
	GetCreate() bool
	GetNoCheckout() bool
}
