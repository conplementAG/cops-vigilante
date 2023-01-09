package http

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_HealthEndpoint(t *testing.T) {
	// Arrange
	g := NewGomegaWithT(t)
	gin.SetMode(gin.TestMode)

	router := createServer()

	// Act
	req, _ := http.NewRequest("GET", "/health", nil)

	// Assert
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	g.Expect(responseRecorder.Code).To(Equal(http.StatusOK))

	body, err := io.ReadAll(responseRecorder.Body)
	g.Expect(err).To(BeNil())
	g.Expect(string(body)).To(ContainSubstring("up and running"))
}
