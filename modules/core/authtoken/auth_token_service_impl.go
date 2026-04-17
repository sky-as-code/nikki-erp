package authtoken

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const (
	JwtAlgoHs256             = "HS256"
	JwtAlgoRs256             = "RS256"
	JwtEdDsa                 = "EdDSA"
	JwtTimeToleranceMins     = 3
	JwtAccessTokenExpiryMax  = 60 * 24     // 1 day
	JwtRefreshTokenExpiryMax = 60 * 24 * 7 // 7 days
	JwtRefreshTokenExpiryMin = 15          // 15 minutes
)

type AuthTokenService interface {
	CreateJwt(ctx corectx.Context, param CreateJwtParam) (*CreateJwtResult, error)
	VerifyJwt(ctx corectx.Context, param VerifyJwtParam) (*VerifyJwtResult, error)
}

type NewAuthTokenServiceImplParam struct {
	dig.In
	Logger    logging.LoggerService
	ConfigSvc config.ConfigService
}

func NewAuthTokenServiceImpl(param NewAuthTokenServiceImplParam) AuthTokenService {
	svc := &AuthTokenServiceImpl{
		logger:    param.Logger,
		configSvc: param.ConfigSvc,
	}
	svc.validateConfig()
	return svc
}

type AuthTokenServiceImpl struct {
	logger    logging.LoggerService
	configSvc config.ConfigService
}

type JwtPurpose string

const (
	JwtPurposeAccessToken  = JwtPurpose("access_token")
	JwtPurposeRefreshToken = JwtPurpose("refresh_token")
)

type CreateJwtParam struct {
	// Custom claims other than the registered claims.
	// Any claims collision with the registered claims will be overridden.
	CustomClaims map[string]any

	Purpose JwtPurpose

	// Value of the "jti" claim.
	Jti *string

	// Value of the "sub" claim.
	Sub string
}

type CreateJwtResult struct {
	Claims jwt.MapClaims
	Token  string
}

func (this *AuthTokenServiceImpl) CreateJwt(ctx corectx.Context, param CreateJwtParam) (*CreateJwtResult, error) {
	if param.Purpose == "" {
		return nil, errors.New("purpose is required")
	}

	algorithm := this.configSvc.GetStr(c.RequestGuardAccessTokenAlgorithm)
	signingKey, err := this.signingKey(algorithm)
	if err != nil {
		return nil, err
	}

	var jwtId string
	if param.Jti != nil {
		jwtId = strings.TrimSpace(*param.Jti)
	} else {
		jwtId, err = this.generateJwtId()
		if err != nil {
			return nil, err
		}
	}

	claims := make(jwt.MapClaims, len(param.CustomClaims)+7)
	for key, value := range param.CustomClaims {
		claims[key] = value
	}

	var expiry uint
	if param.Purpose == JwtPurposeAccessToken {
		expiry = this.configSvc.GetUint(c.RequestGuardAccessTokenExpiryMinutes)
	} else {
		expiry = this.configSvc.GetUint(c.RequestGuardRefreshTokenExpiryMinutes)
	}

	now := model.NewModelDateTime().GoTime()
	expiryDuration := time.Duration(expiry) * time.Minute
	audience := this.configSvc.GetStrArr(c.RequestGuardAccessTokenAudience, "")

	if len(audience) > 0 {
		claims["aud"] = audience
	}
	claims["exp"] = jwt.NewNumericDate(now.Add(expiryDuration)).Unix()
	claims["iss"] = this.configSvc.GetStr(c.RequestGuardAccessTokenIssuer)
	claims["iat"] = jwt.NewNumericDate(now).Unix()
	claims["jti"] = jwtId
	claims["nbf"] = jwt.NewNumericDate(now).Unix()
	claims["sub"] = param.Sub

	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod(algorithm), claims)
	tokenString, err := jwtToken.SignedString(signingKey)
	if err != nil {
		return nil, err
	}

	// Need this so MapClams.GetExpirationTime() can work properly
	claims["exp"] = float64(claims["exp"].(int64))
	claims["iat"] = float64(claims["iat"].(int64))
	claims["nbf"] = float64(claims["nbf"].(int64))

	return &CreateJwtResult{
		Claims: claims,
		Token:  tokenString,
	}, nil
}

func (this *AuthTokenServiceImpl) generateJwtId() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate random JWT ID")
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

type VerifyJwtParam struct {
	Token string
}

type VerifyJwtResult struct {
	IsOk   bool
	Claims jwt.MapClaims
	// Value of the "jti" from `Claims`.
	Jti string
}

func (this *AuthTokenServiceImpl) VerifyJwt(ctx corectx.Context, param VerifyJwtParam) (*VerifyJwtResult, error) {
	inputToken := strings.TrimSpace(param.Token)
	if inputToken == "" {
		return &VerifyJwtResult{IsOk: false}, nil
	}

	algorithm := this.configSvc.GetStr(c.RequestGuardAccessTokenAlgorithm)
	issuer := this.configSvc.GetStr(c.RequestGuardAccessTokenIssuer)
	audiences := nonEmptyStrings(this.configSvc.GetStrArr(c.RequestGuardAccessTokenAudience, ""))

	verifyKey, err := this.verificationKey(algorithm)
	if err != nil {
		return nil, err
	}

	parser := jwt.NewParser(jwtParserOptions(algorithm, issuer, audiences)...)
	jwtToken, err := parser.ParseWithClaims(inputToken, make(jwt.MapClaims, 7), func(_ *jwt.Token) (any, error) {
		return verifyKey, nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "malformed") {
			return nil, err
		}
		return &VerifyJwtResult{IsOk: false}, nil
	}

	claims := jwtToken.Claims.(jwt.MapClaims)
	jwtId := claims["jti"].(string)
	return &VerifyJwtResult{
		IsOk:   true,
		Claims: claims,
		Jti:    jwtId,
	}, nil
}

func (this *AuthTokenServiceImpl) signingKey(algorithm string) (any, error) {
	if algorithm == JwtAlgoRs256 {
		return rsaPrivateKeyFromPem(this.configSvc.GetStr(c.RequestGuardAccessTokenPrivateKey))
	}
	if algorithm == JwtEdDsa {
		return edDsaPrivateKeyFromPem(this.configSvc.GetStr(c.RequestGuardAccessTokenPrivateKey))
	}
	return []byte(this.configSvc.GetStr(c.RequestGuardAccessTokenSecret)), nil
}

func (this *AuthTokenServiceImpl) verificationKey(algorithm string) (any, error) {
	if algorithm == JwtAlgoRs256 {
		return rsaPublicKeyFromPem(this.configSvc.GetStr(c.RequestGuardAccessTokenPublicKey))
	}
	if algorithm == JwtEdDsa {
		return edDsaPublicKeyFromPem(this.configSvc.GetStr(c.RequestGuardAccessTokenPublicKey))
	}
	return []byte(this.configSvc.GetStr(c.RequestGuardAccessTokenSecret)), nil
}

func (this *AuthTokenServiceImpl) validateConfig() {
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

	switch algo {
	case JwtAlgoHs256:
		secret := cfg.GetStr(c.RequestGuardAccessTokenSecret)
		if secret == "" {
			panic(errors.Errorf("config '%s' is required when '%s' is '%s'", c.RequestGuardAccessTokenSecret, c.RequestGuardAccessTokenAlgorithm, JwtAlgoHs256))
		}
	case JwtAlgoRs256:
	case JwtEdDsa:
		publicKey := cfg.GetStr(c.RequestGuardAccessTokenPublicKey)
		privateKey := cfg.GetStr(c.RequestGuardAccessTokenPrivateKey)
		if publicKey == "" || privateKey == "" {
			panic(errors.Errorf(
				"config '%s' and '%s' are required when '%s' is '%s' or '%s'",
				c.RequestGuardAccessTokenPublicKey, c.RequestGuardAccessTokenPrivateKey, c.RequestGuardAccessTokenAlgorithm, JwtAlgoRs256, JwtEdDsa,
			))
		}
	default:
		panic(errors.Errorf("config '%s' must be one of: '%s', '%s' or '%s'", c.RequestGuardAccessTokenAlgorithm, JwtAlgoHs256, JwtAlgoRs256, JwtEdDsa))
	}

	expiryMinsAccess := cfg.GetUint(c.RequestGuardAccessTokenExpiryMinutes)
	if expiryMinsAccess == 0 || expiryMinsAccess > JwtAccessTokenExpiryMax {
		panic(errors.Errorf("config '%s' must be an integer > 0 and <= %d minutes", c.RequestGuardAccessTokenExpiryMinutes, JwtAccessTokenExpiryMax))
	}

	expiryMinsRefresh := cfg.GetUint(c.RequestGuardRefreshTokenExpiryMinutes)
	if expiryMinsRefresh < JwtRefreshTokenExpiryMin || expiryMinsRefresh > JwtRefreshTokenExpiryMax {
		panic(errors.Errorf("config '%s' must be an integer >= %d and <= %d minutes", c.RequestGuardRefreshTokenExpiryMinutes, JwtRefreshTokenExpiryMin, JwtRefreshTokenExpiryMax))
	}
}

func jwtParserOptions(algo, issuer string, audiences []string) []jwt.ParserOption {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{algo}),
		jwt.WithLeeway(time.Duration(JwtTimeToleranceMins) * time.Minute),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithNotBeforeRequired(),
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
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func rsaPublicKeyFromPem(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("RSA public key PEM decode failed")
	}
	if strings.Contains(block.Type, "RSA PUBLIC KEY") {
		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "RSA PKCS1 public key parse failed")
		}
		return publicKey, nil
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "RSA PKIX public key parse failed")
	}
	publicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}
	return publicKey, nil
}

func rsaPrivateKeyFromPem(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("RSA private key PEM decode failed")
	}

	if strings.Contains(block.Type, "RSA PRIVATE KEY") {
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "RSA PKCS1 private key parse failed")
		}
		return privateKey, nil
	}

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "RSA PKCS8 private key parse failed")
	}
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return privateKey, nil
}

func edDsaPublicKeyFromPem(pemStr string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("EdDSA public key PEM decode failed")
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "EdDSA PKIX public key parse failed")
	}
	publicKey, ok := publicKeyAny.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("public key is not Ed25519")
	}
	return publicKey, nil
}

func edDsaPrivateKeyFromPem(pemStr string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("EdDSA private key PEM decode failed")
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "EdDSA PKCS8 private key parse failed")
	}
	privateKey, ok := privateKeyAny.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not Ed25519")
	}
	return privateKey, nil
}
