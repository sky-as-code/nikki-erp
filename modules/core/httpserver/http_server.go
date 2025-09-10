package httpserver

import (
	"context"
	stdErr "errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	httpHost := params.Config.GetStr(c.HttpHost, "127.0.0.1")
	httpPort := params.Config.GetInt32(c.HttpPort, "80")
	httpServer := HttpServer{
		Name:       params.Config.GetStr(c.AppName),
		EchoServer: echoServer,
		Logger:     params.Logger,
		Host:       httpHost,
		Port:       httpPort,
	}

	httpServer.Use(middleware.Logger())
	httpServer.Use(middleware.Recover())
	httpServer.Use(middleware.CORSWithConfig(configCors(params.Config)))
	httpServer.Use(m.RequestContextMiddleware)

	return httpServerResult{
		HttpServer: &httpServer,
		RootRoute:  initRoutes(httpServer, params.Config),
	}
}

func configCors(config config.ConfigService) middleware.CORSConfig {
	corsOrigins := config.GetStrArr(c.HttpCorsOrigins, "*")
	corsHeaders := config.GetStrArr(c.HttpCorsHeaders, "")
	if len(corsHeaders) == 0 {
		corsHeaders = []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}
	}
	corsMethods := config.GetStrArr(c.HttpCorsMethods, "")
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

func initRoutes(mainServer HttpServer, config config.ConfigService) *echo.Group {
	basePath := config.GetStr(c.HttpBasePath, "") // or "/api"
	routeGroup := mainServer.EchoServer.Group(basePath)
	return routeGroup
}

func newEchoServer() *echo.Echo {
	echoServer := echo.New()
	echoServer.HideBanner = true // Not logging startup banner on start
	echoServer.HidePort = true   // Not logging port information on start
	echoServer.HTTPErrorHandler = CustomHttpErrorHandler(echoServer.DefaultHTTPErrorHandler)
	// echoServer.Validator = validator.TagBased
	return echoServer
}

type HttpServer struct {
	Name       string
	EchoServer *echo.Echo
	Logger     logging.LoggerService
	Host       string
	Port       int32
}

func (this HttpServer) Start() error {
	address := fmt.Sprintf("%s:%d", this.Host, this.Port)
	this.Logger.Infof("Starting HTTP server %s at %s", this.Name, address)
	err := this.EchoServer.Start(address)

	if err != nil && !stdErr.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (this HttpServer) Shutdown() {
	this.Logger.Infof("Stopping HTTP server %s ...", this.Name)
	_ = this.EchoServer.Shutdown(context.Background()) //nolint:errcheck
	this.Logger.Infof("HTTP server %s exited", this.Name)
}

func (this HttpServer) Use(middlewares ...echo.MiddlewareFunc) {
	this.EchoServer.Use(middlewares...)
}
