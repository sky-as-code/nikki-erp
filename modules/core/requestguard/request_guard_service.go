package requestguard

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

type StaticRequestGuardServiceParams struct {
	dig.In

	ConfigSvc config.ConfigService
}

func NewStaticRequestGuardServiceImpl(params StaticRequestGuardServiceParams) RequestGuardService {
	svc := &StaticRequestGuardServiceImpl{
		configSvc: params.ConfigSvc,
	}
	svc.validateConfig()
	return svc
}

type StaticRequestGuardServiceImpl struct {
	configSvc      config.ConfigService
	corsMiddleware echo.MiddlewareFunc
}

func (this *StaticRequestGuardServiceImpl) CalcRequestFingerprint(request *http.Request) (fingerprint string, err error) {
	if this.configSvc.GetBool(c.RequestGuardAccessTokenEnabled) {
		rawToken := this.bearerAccessToken(request)
		parts := strings.Split(rawToken, ".")
		return parts[0] + "." + parts[1], nil
	}
	return "", errors.New("not implemented")
}

func (this *StaticRequestGuardServiceImpl) GetCorsMiddleware() (echo.MiddlewareFunc, error) {
	if !this.configSvc.GetBool(c.RequestGuardCorsEnabled) {
		return nil, nil
	}

	if this.corsMiddleware == nil {
		this.corsMiddleware = middleware.CORSWithConfig(this.configCors())
	}
	return this.corsMiddleware, nil
}

func (this *StaticRequestGuardServiceImpl) configCors() middleware.CORSConfig {
	corsOrigins := this.configSvc.GetStrArr(c.HttpCorsOrigins)
	corsHeaders := this.configSvc.GetStrArr(c.HttpCorsHeaders, "")
	if len(corsHeaders) == 0 {
		corsHeaders = []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}
	}
	corsMethods := this.configSvc.GetStrArr(c.HttpCorsMethods, "")
	if len(corsMethods) == 0 {
		corsMethods = []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE}
	}

	return middleware.CORSConfig{
		// TODO: Allow config CORS from database
		AllowOrigins: corsOrigins,
		AllowHeaders: corsHeaders,
		AllowMethods: corsMethods,
	}
}

func (this *StaticRequestGuardServiceImpl) VerifyTrustedConnection(request *http.Request) (result *VerifyRequestResult, err error) {
	return &VerifyRequestResult{
		IsOk:       true,
		HttpStatus: http.StatusOK,
	}, nil
}

func (this *StaticRequestGuardServiceImpl) VerifyJwt(request *http.Request) (*VerifyRequestResult, error) {
	cfg := this.configSvc
	if !cfg.GetBool(c.RequestGuardAccessTokenEnabled) {
		return &VerifyRequestResult{
			IsOk:       true,
			HttpStatus: http.StatusOK,
		}, nil
	}

	rawToken := this.bearerAccessToken(request)
	if rawToken == "" {
		return jwtVerifyFailure(), nil
	}

	algo := cfg.GetStr(c.RequestGuardAccessTokenAlgorithm)
	issuer := cfg.GetStr(c.RequestGuardAccessTokenIssuer)
	audiences := nonEmptyStrings(cfg.GetStrArr(c.RequestGuardAccessTokenAudience, ""))

	var secretBytes any
	if algo == JWT_ALGO_RS256 {
		var err error
		secretBytes, err = rsaPublicKeyFromPem(cfg.GetStr(c.RequestGuardAccessTokenRsaPublicKey))
		if err != nil {
			return nil, err
		}
	}

	secret := cfg.GetStr(c.RequestGuardAccessTokenSha256Secret)
	secretBytes = []byte(secret)
	parser := jwt.NewParser(jwtParserOptions(algo, issuer, audiences)...)
	claims := &jwt.RegisteredClaims{}
	jwtToken, err := parser.ParseWithClaims(rawToken, claims, func(_ *jwt.Token) (any, error) {
		return secretBytes, nil
	})
	if err != nil {
		return jwtVerifyFailure(), nil
	}

	if this.configSvc.GetBool(c.RequestGuardAccessTokenDpopEnabled) {
		result, dpopErr := this.VerifyJwtDpop(request)
		if result != nil || dpopErr != nil {
			return result, dpopErr
		}
	}
	return &VerifyRequestResult{
		IsOk:       true,
		HttpStatus: http.StatusOK,
		JwtClaims:  jwtToken.Claims,
	}, nil
}

func jwtParserOptions(algo, issuer string, audiences []string) []jwt.ParserOption {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{algo}),
		jwt.WithLeeway(time.Duration(JWT_NBF_DRIFT_MINS) * time.Minute),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	}
	if issuer != "" {
		opts = append(opts, jwt.WithIssuer(issuer))
	}
	if len(audiences) > 0 {
		opts = append(opts, jwt.WithAllAudiences(audiences...))
	}
	return opts
}

func nonEmptyStrings(values []string) []string {
	var out []string
	for _, s := range values {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (this *StaticRequestGuardServiceImpl) bearerAccessToken(request *http.Request) string {
	headerName := strings.TrimSpace(this.configSvc.GetStr(c.RequestGuardAccessTokenHttpHeaderName))
	prefix := strings.TrimSpace(this.configSvc.GetStr(c.RequestGuardAccessTokenHttpHeaderPrefix))
	auth := strings.TrimSpace(request.Header.Get(headerName))

	if auth == "" {
		return ""
	}
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	raw := strings.TrimSpace(auth[len(prefix):])
	if raw == "" {
		return ""
	}
	return raw
}

func jwtVerifyFailure() *VerifyRequestResult {
	return &VerifyRequestResult{
		IsOk:       false,
		HttpStatus: http.StatusUnauthorized,
		ClientError: &ft.ClientErrorItem{
			Key:     ft.ErrorKey("err_invalid_access_token", "authorize"),
			Message: "invalid or expired access token",
		},
	}
}

func rsaPublicKeyFromPem(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("RSA public key PEM decode failed")
	}
	if strings.Contains(block.Type, "RSA PUBLIC KEY") {
		pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "RSA PKCS1 public key parse failed")
		}
		return pub, nil
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "RSA PKIX public key parse failed")
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}
	return pub, nil
}

// Verify JWT DPoP (OAuth2 Demonstraing Proof of Possession)
func (this *StaticRequestGuardServiceImpl) VerifyJwtDpop(request *http.Request) (*VerifyRequestResult, error) {

	return nil, nil
}

func (this *StaticRequestGuardServiceImpl) VerifySessionBlacklist(request *http.Request) (*VerifyRequestResult, error) {

	return nil, nil
}

func (this *StaticRequestGuardServiceImpl) validateConfig() {
	cfg := this.configSvc
	headerName := strings.TrimSpace(cfg.GetStr(c.RequestGuardAccessTokenHttpHeaderName))
	if headerName == "" {
		panic(errors.Errorf("config '%s' and '%s' are required", c.RequestGuardAccessTokenHttpHeaderName, c.RequestGuardAccessTokenHttpHeaderPrefix))
	}

	prefix := strings.TrimSpace(cfg.GetStr(c.RequestGuardAccessTokenHttpHeaderPrefix))
	if prefix == "" {
		panic(errors.Errorf("config '%s' is required", c.RequestGuardAccessTokenHttpHeaderPrefix))
	}

	algo := cfg.GetStr(c.RequestGuardAccessTokenAlgorithm)
	expiryMins := cfg.GetUint(c.RequestGuardAccessTokenExpiryMinutes)

	if algo != JWT_ALGO_HS256 && algo != JWT_ALGO_RS256 {
		panic(errors.Errorf("config '%s' must be either '%s' or '%s'", c.RequestGuardAccessTokenAlgorithm, JWT_ALGO_HS256, JWT_ALGO_RS256))
	}

	if algo == JWT_ALGO_HS256 {
		secret := cfg.GetStr(c.RequestGuardAccessTokenSha256Secret)
		if secret == "" {
			panic(errors.Errorf("config '%s' is required when '%s' is '%s'", c.RequestGuardAccessTokenSha256Secret, c.RequestGuardAccessTokenAlgorithm, JWT_ALGO_HS256))
		}
	} else if algo == JWT_ALGO_RS256 {
		publicKey := cfg.GetStr(c.RequestGuardAccessTokenRsaPublicKey)
		privateKey := cfg.GetStr(c.RequestGuardAccessTokenRsaPrivateKey)
		if publicKey == "" || privateKey == "" {
			panic(errors.Errorf("config '%s' and '%s' are required when '%s' is '%s'", c.RequestGuardAccessTokenRsaPublicKey, c.RequestGuardAccessTokenRsaPrivateKey, c.RequestGuardAccessTokenAlgorithm, JWT_ALGO_RS256))
		}
	}

	if expiryMins == 0 || expiryMins > JWT_EXP_MAX_MINS {
		panic(errors.Errorf("config '%s' must be > 0 and <= %d", c.RequestGuardAccessTokenExpiryMinutes, JWT_EXP_MAX_MINS))
	}
}
