package httpserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	stdErr "errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type httpServerParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService
}

type httpServerResult struct {
	dig.Out

	HttpServer *HttpServer
	RootRoute  *echo.Group
}

func initHttpServer(params httpServerParams) httpServerResult {
	echoServer := newEchoServer()

	httpServer := HttpServer{
		echoServer: echoServer,
		logger:     params.Logger,
		config:     params.Config,
	}

	httpServer.Use(middleware.RequestLogger())
	httpServer.Use(middleware.Recover())
	httpServer.Use(m.RequestContextMiddleware3)
	applyGlobalLazywares(&httpServer)
	httpServer.UseLazy(m.Lazyware(m.CorsEchoMiddleware))
	// httpServer.Use(m.EnsureAuthorized)

	return httpServerResult{
		HttpServer: &httpServer,
		RootRoute:  initRoutes(httpServer, params.Config),
	}
}

func initRoutes(mainServer HttpServer, config config.ConfigService) *echo.Group {
	basePath := config.GetStr(c.HttpBasePath) // or "/api"
	routeGroup := mainServer.echoServer.Group(basePath)
	return routeGroup
}

func newEchoServer() *echo.Echo {
	return echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: CustomHttpErrorHandler(echo.DefaultHTTPErrorHandler(false)),
	})
}

type MiddlewareCreator func() echo.MiddlewareFunc

var globalLazywares = []MiddlewareCreator{}

func applyGlobalLazywares(server *HttpServer) {
	for _, lazyware := range globalLazywares {
		server.UseLazy(m.Lazyware(lazyware))
	}
}

func AppendGlobalLazywares(middlewares ...MiddlewareCreator) {
	globalLazywares = append(globalLazywares, middlewares...)
}

type HttpServer struct {
	Name            string
	echoServer      *echo.Echo
	httpServer      *http.Server
	logger          logging.LoggerService
	LazyMiddlewares []*m.LazyMiddleware
	config          config.ConfigService
}

func (this *HttpServer) Start() error {
	this.Name = this.config.GetStr(c.AppName)
	httpsEnabled := this.config.GetBool(c.HttpsEnabled)
	if httpsEnabled {
		return this.startHttps()
	}
	return this.startHttp()
}

func (this *HttpServer) startHttp() error {
	httpHost := this.config.GetStr(c.HttpHost)
	httpPort := this.config.GetInt32(c.HttpPort)
	address := fmt.Sprintf("%s:%d", httpHost, httpPort)

	this.logger.Infof("Starting HTTP server %s at %s", this.Name, address)

	for _, lazyware := range this.LazyMiddlewares {
		lazyware.Enable()
	}

	httpServer := &http.Server{
		Addr:    address,
		Handler: this.echoServer,
	}

	this.httpServer = httpServer
	err := httpServer.ListenAndServe()
	if err != nil && !stdErr.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (this *HttpServer) startHttps() error {
	httpHost := this.config.GetStr(c.HttpHost)
	httpsPort := this.config.GetInt32(c.HttpsPort)
	httpsRsaPublicKey := this.config.GetStr(c.HttpsRsaPublicKey)
	httpsRsaPrivateKey := this.config.GetStr(c.HttpsRsaPrivateKey)
	mtlsEnabled := this.config.GetBool(c.MtlsEnabled)
	address := fmt.Sprintf("%s:%d", httpHost, httpsPort)

	this.logger.Infof("Starting HTTPS server %s at %s", this.Name, address)

	for _, lazyware := range this.LazyMiddlewares {
		lazyware.Enable()
	}

	cert, err := tls.X509KeyPair([]byte(httpsRsaPublicKey), []byte(httpsRsaPrivateKey))
	if err != nil {
		return fmt.Errorf("load server certificate and key from config: %w", err)
	}

	var caPool *x509.CertPool
	// clientAuthType := tls.NoClientCert
	clientAuthType := tls.RequireAnyClientCert
	if mtlsEnabled {
		clientAuthType = tls.RequireAndVerifyClientCert

		mtlsClientCaCert := this.config.GetStr(c.MtlsClientCaCert)
		caPool = x509.NewCertPool()
		if !caPool.AppendCertsFromPEM([]byte(mtlsClientCaCert)) {
			return fmt.Errorf("append client CA certificate from config to pool")
		}
	}

	tlsConfig := &tls.Config{
		GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			// TODO: Dynamically load certificate and key from config based on client hostname

			return &tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   clientAuthType,
				ClientCAs:    caPool,
				MinVersion:   tls.VersionTLS13,
			}, nil
		},
	}

	tlsServer := &http.Server{
		Addr:      address,
		Handler:   this.echoServer,
		TLSConfig: tlsConfig,
	}

	this.httpServer = tlsServer
	err = tlsServer.ListenAndServeTLS("", "")

	if err != nil && !stdErr.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (this HttpServer) Shutdown() {
	this.logger.Infof("Stopping HTTP(S) server %s ...", this.Name)
	if this.httpServer != nil {
		_ = this.httpServer.Shutdown(context.Background()) //nolint:errcheck
	}
	this.logger.Infof("HTTP(S) server %s exited", this.Name)
}

func (this HttpServer) Use(middlewares ...echo.MiddlewareFunc) {
	this.echoServer.Use(middlewares...)
}

func (this *HttpServer) UseLazy(lazywares ...*m.LazyMiddleware) {
	for _, lazyware := range lazywares {
		this.LazyMiddlewares = append(this.LazyMiddlewares, lazyware)
		this.Use(lazyware.Middleware())
	}
}
