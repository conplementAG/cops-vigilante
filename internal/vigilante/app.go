package vigilante

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/cli"
	"github.com/conplementag/cops-vigilante/internal/vigilante/errors"
	"github.com/conplementag/cops-vigilante/internal/vigilante/http"
	"github.com/conplementag/cops-vigilante/internal/vigilante/metrics"
	"github.com/conplementag/cops-vigilante/internal/vigilante/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Run() {
	if viper.GetBool("verbose") {
		gin.SetMode(gin.DebugMode)
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		logrus.SetLevel(logrus.InfoLevel)
	}

	metrics.Init()
	scheduler.InitializeAndStart(viper.GetInt(cli.IntervalFlag))
	err := http.Start()
	errors.PanicOnError(err)
}
