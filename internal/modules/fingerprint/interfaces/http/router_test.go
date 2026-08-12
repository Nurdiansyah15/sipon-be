package http

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesSandboxEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	dummy := func(c *gin.Context) {}

	RegisterRoutes(engine.Group("/"), &FingerprintHandler{}, dummy, dummy, true)

	found := false
	for _, r := range engine.Routes() {
		if r.Path == "/api/v1/web/fingerprint/sandbox/scan" && r.Method == "POST" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected sandbox scan route when sandbox enabled")
	}
}

func TestRegisterRoutesSandboxDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	dummy := func(c *gin.Context) {}

	RegisterRoutes(engine.Group("/"), &FingerprintHandler{}, dummy, dummy, false)

	for _, r := range engine.Routes() {
		if r.Path == "/api/v1/web/fingerprint/sandbox/scan" {
			t.Fatalf("sandbox route should not exist when sandbox disabled, found %s %s", r.Method, r.Path)
		}
	}
}
