package server

import (
	"microg/examples/user/client/config"
	"microg/examples/user/client/controller"
	"microg/pkg/common/core"
	"microg/pkg/log"
	"microg/server/restserver"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

type App struct {
	Cfg        *config.Config
	HTTPServer *restserver.Server
}

var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewApp,
)

func NewHTTPServer(cfg *config.Config, userCtrl *controller.UserController) *restserver.Server {
	server := restserver.NewServer(
		restserver.WithPort(cfg.HTTP.Port),
		restserver.WithMode(cfg.HTTP.Mode),
		restserver.WithHealthz(cfg.HTTP.Healthz),
		restserver.WithMetrics(cfg.HTTP.Metrics),
	)
	r := server.Engine
	r.GET("/", func(c *gin.Context) {
		core.WriteResponse(c, nil, gin.H{
			"service":  cfg.Service.Name,
			"version":  "1.0.0",
			"user_srv": cfg.UserSrv.Name,
		})
	})
	v1 := r.Group("/api/v1/user")
	userCtrl.RegisterRoutes(v1)
	return server
}

func NewApp(cfg *config.Config, httpServer *restserver.Server) (*App, func(), error) {
	cleanup := func() {
		log.Infof("清理资源...")
	}
	return &App{Cfg: cfg, HTTPServer: httpServer}, cleanup, nil
}
