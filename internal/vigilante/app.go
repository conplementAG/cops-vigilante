package vigilante

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/cli"
	"github.com/conplementag/cops-vigilante/internal/vigilante/http"
	"github.com/conplementag/cops-vigilante/internal/vigilante/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Run() {
	scheduler.InitializeAndStart(viper.GetInt(cli.IntervalInSecondsFlag))
	http.Start(gin.Default())
}
