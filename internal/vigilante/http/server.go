package http

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/cli"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Start() error {
	router := CreateServer()

	if viper.GetBool(cli.TLSFlag) {
		port := "8443"
		logrus.Info("Starting the server with TLS on " + port)
		return router.RunTLS(":"+port, "/etc/vigilante/certs/tls.crt", "/etc/vigilante/certs/tls.key")
	} else {
		port := "8000"
		logrus.Info("Starting the server (plain HTTP) on " + port)
		return router.Run(":" + port)
	}
}

func CreateServer() *gin.Engine {
	router := gin.Default()

	logrus.Info("We don't trust any proxies by default.")
	router.SetTrustedProxies([]string{})

	logrus.Info("Adding controller routes")
	addRoutes(router)

	return router
}

func addRoutes(router *gin.Engine) {
	healthController := HealthController{}
	router.GET("/health", healthController.Check)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
