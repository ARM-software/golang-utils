package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/config"
)

// GitActionConfig describes how a clone or checkout should be performed.
type GitActionConfig struct {
	// The (possibly remote) repository URL to clone from.
	URL string `mapstructure:"url"`
	// Auth credentials, if required, to use with the remote repository.
	Auth http.BasicAuth `mapstructure:"auth"`
	// Limit fetching to the specified number of commits.
	Depth int `mapstructure:"depth"`
	// Regerence can be a hash, a branch, or a tag
	Reference string `mapstructure:"ref"`
	// RecurseSubmodules after the clone is created, initialise all submodules
	// within, using their default settings. This option is ignored if the
	// cloned repository does not have a worktree.
	RecurseSubmodules bool `mapstructure:"recursive_submodules"`
	// Tags describe how the tags will be fetched from the remote repository,
	// by default is AllTags.
	Tags git.TagMode `mapstructure:"tags"`
	// CreateBranch a new branch named Branch and start it at Hash.
	CreateBranch bool `mapstructure:"create_branch"`
	// No checkout of HEAD after clone if true.
	NoCheckout bool
}

func (c *GitActionConfig) GetURL() string {
	return c.URL
}

func (c *GitActionConfig) GetAuth() *http.BasicAuth {
	return &c.Auth
}

func (c *GitActionConfig) GetDepth() int {
	return c.Depth
}

func (c *GitActionConfig) GetReference() string {
	return c.Reference
}

func (c *GitActionConfig) GetRecursiveSubModules() bool {
	return c.RecurseSubmodules
}

func (c *GitActionConfig) GetTags() git.TagMode {
	return c.Tags
}

func (c *GitActionConfig) GetCreate() bool {
	return c.CreateBranch
}

func (c *GitActionConfig) GetNoCheckout() bool {
	return c.NoCheckout
}

func (c *GitActionConfig) Validate() error {
	validation.ErrorTag = "mapstructure"

	// Validate Embedded Structs
	err := config.ValidateEmbedded(c)
	if err != nil {
		return err
	}
	return validation.ValidateStruct(c,
		validation.Field(&c.URL, validation.Required),
	)
}

func NewGitActionConfig(url string) GitActionConfig {
	return GitActionConfig{
		URL: url,
	}
}
