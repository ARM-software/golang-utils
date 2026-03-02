package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/config"
)

// SSHAuthConfig holds SSH-specific authentication configuration.
// It is mutually exclusive with the HTTP basic auth (Auth field).
type SSHAuthConfig struct {
	// PrivateKeyPath is the path to the PEM-encoded SSH private key file.
	// Required when UseAgent is false.
	PrivateKeyPath string `mapstructure:"private_key_path"`
	// PrivateKeyPassword is an optional passphrase protecting the private key.
	PrivateKeyPassword string `mapstructure:"private_key_password"`
	// KnownHostsFile is a path to a known_hosts file used for host key
	// verification. When empty, the system default (~/.ssh/known_hosts) is
	// used unless UseInsecureHostKey is true.
	KnownHostsFile string `mapstructure:"known_hosts_file"`
	// UseInsecureHostKey disables host key verification.
	// WARNING: this makes the connection vulnerable to MITM attacks.
	UseInsecureHostKey bool `mapstructure:"insecure_ignore_host_key"`
	// UseAgent authenticates via the running SSH agent instead of a key file.
	UseAgent bool `mapstructure:"use_agent"`
	// Username is the SSH user. Defaults to "git" when empty.
	Username string `mapstructure:"username"`
}

// isConfigured reports whether any SSH authentication has been requested.
func (s *SSHAuthConfig) isConfigured() bool {
	return s.UseAgent || s.PrivateKeyPath != ""
}

func (s *SSHAuthConfig) username() string {
	if s.Username == "" {
		return "git"
	}
	return s.Username
}

// Validate validates the SSHAuthConfig using ozzo-validation rules.
func (s *SSHAuthConfig) Validate() error {
	validation.ErrorTag = "mapstructure"
	return validation.ValidateStruct(s,
		// PrivateKeyPath is required when SSH is configured without the agent
		validation.Field(&s.PrivateKeyPath,
			validation.When(s.isConfigured() && !s.UseAgent, validation.Required),
		),
		// KnownHostsFile and UseInsecureHostKey are mutually exclusive
		validation.Field(&s.KnownHostsFile,
			validation.When(s.UseInsecureHostKey, validation.Length(0, 0).Error("must be empty when insecure_ignore_host_key is true")),
		),
	)
}

// ToAuthMethod builds the go-git transport.AuthMethod for the SSH config.
func (s *SSHAuthConfig) ToAuthMethod() (transport.AuthMethod, error) {
	if s.UseAgent {
		return gitssh.NewSSHAgentAuth(s.username())
	}

	pkAuth, err := gitssh.NewPublicKeysFromFile(s.username(), s.PrivateKeyPath, s.PrivateKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to load SSH private key %q: %v", commonerrors.ErrInvalid, s.PrivateKeyPath, err)
	}

	if s.UseInsecureHostKey {
		pkAuth.HostKeyCallback = ssh.InsecureIgnoreHostKey() //nolint:gosec
	} else if s.KnownHostsFile != "" {
		cb, cbErr := knownhosts.New(s.KnownHostsFile)
		if cbErr != nil {
			return nil, fmt.Errorf("%w: failed to load known_hosts file %q: %v", commonerrors.ErrInvalid, s.KnownHostsFile, cbErr)
		}
		pkAuth.HostKeyCallback = cb
	}
	// When KnownHostsFile is empty and UseInsecureHostKey is false,
	// go-git's ssh.PublicKeys falls back to the system known_hosts file.

	return pkAuth, nil
}

// GitActionConfig describes how a clone or checkout should be performed.
type GitActionConfig struct {
	// The (possibly remote) repository URL to clone from.
	URL string `mapstructure:"url"`
	// Auth credentials, if required, to use with the remote repository (HTTP/HTTPS).
	// Mutually exclusive with SSH.
	Auth http.BasicAuth `mapstructure:"auth"`
	// SSH holds SSH-specific authentication configuration.
	// Takes precedence over Auth when configured.
	SSH SSHAuthConfig `mapstructure:"ssh"`
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

// GetAuth returns the HTTP basic-auth credentials as a transport.AuthMethod.
// For SSH-based authentication use GetSSHAuth or ResolveAuth.
func (c *GitActionConfig) GetAuth() transport.AuthMethod {
	return &c.Auth
}

// GetSSHAuth returns an SSH transport.AuthMethod built from the SSH field.
// Returns (nil, nil) when no SSH configuration is present.
func (c *GitActionConfig) GetSSHAuth() (transport.AuthMethod, error) {
	if !c.SSH.isConfigured() {
		return nil, nil
	}
	return c.SSH.ToAuthMethod()
}

// ResolveAuth returns the appropriate transport.AuthMethod for this config.
// SSH auth takes precedence when the SSH field is populated; otherwise HTTP
// basic auth is returned (nil when no credentials are set).
func (c *GitActionConfig) ResolveAuth() (transport.AuthMethod, error) {
	if c.SSH.isConfigured() {
		return c.SSH.ToAuthMethod()
	}
	if c.Auth.Username != "" || c.Auth.Password != "" {
		return &c.Auth, nil
	}
	return nil, nil
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

// NewSSHGitActionConfig creates a GitActionConfig pre-populated for SSH
// authentication using a private key file.
// Pass an empty knownHostsFile to use the system default known_hosts.
func NewSSHGitActionConfig(url, privateKeyPath, privateKeyPassword, knownHostsFile string) GitActionConfig {
	return GitActionConfig{
		URL: url,
		SSH: SSHAuthConfig{
			PrivateKeyPath:     privateKeyPath,
			PrivateKeyPassword: privateKeyPassword,
			KnownHostsFile:     knownHostsFile,
		},
	}
}

// NewSSHAgentGitActionConfig creates a GitActionConfig pre-populated for SSH
// agent authentication.
func NewSSHAgentGitActionConfig(url string) GitActionConfig {
	return GitActionConfig{
		URL: url,
		SSH: SSHAuthConfig{
			UseAgent: true,
		},
	}
}
