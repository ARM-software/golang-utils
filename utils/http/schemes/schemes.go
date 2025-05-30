package schemes

const (
	AuthorisationSchemeToken        = "Token"
	AuthorisationSchemeBasic        = "Basic"
	AuthorisationSchemeBearer       = "Bearer"
	AuthorisationSchemeConcealed    = "Concealed"
	AuthorisationSchemeDigest       = "Digest"
	AuthorisationSchemeDPoP         = "DPoP"
	AuthorisationSchemeGNAP         = "GNAP"
	AuthorisationSchemeHOBA         = "HOBA"
	AuthorisationSchemeMutual       = "Mutual"
	AuthorisationSchemeNegotiate    = "Negotiate"
	AuthorisationSchemeOAuth        = "OAuth"
	AuthorisationSchemePrivateToken = "PrivateToken"
	AuthorisationSchemeSCRAMSHA1    = "SCRAM-SHA-1"
	AuthorisationSchemeSCRAMSHA256  = "SCRAM-SHA-256"
	AuthorisationSchemeVapid        = "vapid"
)

var (
	// HTTPAuthorisationSchemes lists all supported authorisation schemes. See https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml
	HTTPAuthorisationSchemes = []string{
		AuthorisationSchemeToken,
		AuthorisationSchemeBasic,
		AuthorisationSchemeBearer,
		AuthorisationSchemeConcealed,
		AuthorisationSchemeDigest,
		AuthorisationSchemeDPoP,
		AuthorisationSchemeGNAP,
		AuthorisationSchemeHOBA,
		AuthorisationSchemeMutual,
		AuthorisationSchemeNegotiate,
		AuthorisationSchemeOAuth,
		AuthorisationSchemePrivateToken,
		AuthorisationSchemeSCRAMSHA1,
		AuthorisationSchemeSCRAMSHA256,
		AuthorisationSchemeVapid,
	}
	InAuthSchemes = []any{
		AuthorisationSchemeToken,
		AuthorisationSchemeBasic,
		AuthorisationSchemeBearer,
		AuthorisationSchemeConcealed,
		AuthorisationSchemeDigest,
		AuthorisationSchemeDPoP,
		AuthorisationSchemeGNAP,
		AuthorisationSchemeHOBA,
		AuthorisationSchemeMutual,
		AuthorisationSchemeNegotiate,
		AuthorisationSchemeOAuth,
		AuthorisationSchemePrivateToken,
		AuthorisationSchemeSCRAMSHA1,
		AuthorisationSchemeSCRAMSHA256,
		AuthorisationSchemeVapid,
	}
)
