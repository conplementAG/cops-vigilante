package http_test

import (
	. "github.com/conplementag/cops-vigilante/internal/vigilante/http"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("HealthController", func() {
	var router *gin.Engine

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = CreateServer()
	})

	It("Should serve the health endpoint", func() {
		// Act
		req, _ := http.NewRequest("GET", "/health", nil)

		// Assert
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, req)

		Expect(responseRecorder.Code).To(Equal(http.StatusOK))

		body, err := io.ReadAll(responseRecorder.Body)
		Expect(err).To(BeNil())
		Expect(string(body)).To(ContainSubstring("up and running"))
	})
})
