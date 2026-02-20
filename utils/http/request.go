package http

import (
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/http/schemes"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	AuthorisationSchemeToken        = schemes.AuthorisationSchemeToken
	AuthorisationSchemeBasic        = schemes.AuthorisationSchemeBasic
	AuthorisationSchemeBearer       = schemes.AuthorisationSchemeBearer
	AuthorisationSchemeConcealed    = schemes.AuthorisationSchemeConcealed
	AuthorisationSchemeDigest       = schemes.AuthorisationSchemeDigest
	AuthorisationSchemeDPoP         = schemes.AuthorisationSchemeDPoP
	AuthorisationSchemeGNAP         = schemes.AuthorisationSchemeGNAP
	AuthorisationSchemeHOBA         = schemes.AuthorisationSchemeHOBA
	AuthorisationSchemeMutual       = schemes.AuthorisationSchemeMutual
	AuthorisationSchemeNegotiate    = schemes.AuthorisationSchemeNegotiate
	AuthorisationSchemeOAuth        = schemes.AuthorisationSchemeOAuth
	AuthorisationSchemePrivateToken = schemes.AuthorisationSchemePrivateToken
	AuthorisationSchemeSCRAMSSHA1   = schemes.AuthorisationSchemeSCRAMSHA1
	AuthorisationSchemeSCRAMSHA256  = schemes.AuthorisationSchemeSCRAMSHA256
	AuthorisationSchemeVapid        = schemes.AuthorisationSchemeVapid
)

var (
	HTTPAuthorisationSchemes = schemes.HTTPAuthorisationSchemes
	inAuthSchemes            = schemes.InAuthSchemes
)

// Auth defines a typical HTTP client authentication/authorisation configuration
// See https://datatracker.ietf.org/doc/html/rfc7235
type Auth struct {
	Enforced    bool   `mapstructure:"enforced"`
	Scheme      string `mapstructure:"scheme"`
	AccessToken string `mapstructure:"token"` //nolint:gosec //G117: Exported struct field
}

const (
	missingScheme      = "!!MISSING_SCHEME!!"
	missingAccessToken = "!!MISSING_ACCESS_TOKEN!!"
)

func (cfg *Auth) GetAuthorizationHeader() string {
	if !cfg.Enforced {
		return ""
	}

	scheme, accessToken := cfg.Scheme, cfg.AccessToken
	if scheme == "" {
		scheme = missingScheme
	}
	if accessToken == "" {
		accessToken = missingAccessToken
	}

	return fmt.Sprintf("%v %v", scheme, accessToken)
}

func (cfg *Auth) Validate() (err error) {
	err = configUtils.ValidateEmbedded(cfg)
	if err != nil {
		return
	}
	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Scheme, validation.When(cfg.Enforced, validation.Required, validation.In(inAuthSchemes...))),
		validation.Field(&cfg.AccessToken, validation.Required.When(cfg.Enforced)),
	)
}

// NewAuthConfiguration returns a configuration based on an authorization header
func NewAuthConfiguration(authorizationHeader *string) (cfg *Auth, err error) {
	if reflection.IsEmpty(authorizationHeader) {
		cfg = &Auth{}
		return
	}
	elem := strings.Split(*authorizationHeader, " ")
	if len(elem) < 2 {
		err = commonerrors.New(commonerrors.ErrInvalid, "authorization header does not comply with the header syntax (https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Authorization)")
		return
	}
	cfg = &Auth{
		Enforced:    true,
		Scheme:      elem[0],
		AccessToken: elem[1],
	}
	err = cfg.Validate()
	return
}

type Target struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

func (cfg *Target) GetTargetAddress() string {
	port := strings.TrimSpace(cfg.Port)
	if port == "" {
		return cfg.Host
	}
	return fmt.Sprintf("%v:%v", strings.TrimSpace(cfg.Host), port)
}

func (cfg *Target) Validate() (err error) {
	err = configUtils.ValidateEmbedded(cfg)
	if err != nil {
		return
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Host, validation.Required, is.URL),
		validation.Field(&cfg.Port, is.UTFNumeric, is.Port),
	)
}

// RequestConfiguration defines the typical configuration for an HTTP requests so it can contact a particular target.
type RequestConfiguration struct {
	Target        `mapstructure:"target"`
	UserAgent     string                   `mapstructure:"user_agent"`
	Authorisation Auth                     `mapstructure:"authorisation"`
	Retries       RetryPolicyConfiguration `mapstructure:"retries"`
}

func (cfg *RequestConfiguration) Validate() (err error) {
	err = configUtils.ValidateEmbedded(cfg)
	if err != nil {
		return
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.UserAgent, validation.Required),
		validation.Field(&cfg.Authorisation, validation.Required),
		validation.Field(&cfg.Retries, validation.Required),
	)
}

func DefaultHTTPRequestConfiguration(userAgent string) *RequestConfiguration {
	return &RequestConfiguration{
		UserAgent: userAgent,
	}
}

func DefaultHTTPRequestWithAuthorisationConfigurationEnforced(userAgent string) *RequestConfiguration {
	return &RequestConfiguration{
		UserAgent: userAgent,
		Authorisation: Auth{
			Enforced: true,
		},
		Retries: *DefaultExponentialBackoffRetryPolicyConfiguration(),
	}
}
