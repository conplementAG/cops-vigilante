package vigilante

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/cli"
	"github.com/conplementag/cops-vigilante/internal/vigilante/http"
	"github.com/conplementag/cops-vigilante/internal/vigilante/metrics"
	"github.com/conplementag/cops-vigilante/internal/vigilante/scheduler"
	"github.com/spf13/viper"
)

func Run() {
	metrics.Init()
	scheduler.InitializeAndStart(viper.GetInt(cli.IntervalInSecondsFlag))
	http.Start()
}
