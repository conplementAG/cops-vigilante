package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type HealthController struct {
}

func (c *HealthController) Check(context *gin.Context) {
	context.String(http.StatusOK, fmt.Sprintf("%s up and running", time.Now().String()))
}
