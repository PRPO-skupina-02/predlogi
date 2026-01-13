package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/healthcheck", healthcheck)

	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	w := performRequest(router, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func performRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
