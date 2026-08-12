package http

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRegisterRoutesNoConflict memastikan registrasi route akademik tidak
// panic karena konflik wildcard gin (mis. :id vs :periodId pada posisi sama).
func TestRegisterRoutesNoConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	dummy := func(c *gin.Context) {}

	RegisterRoutes(engine.Group("/"), &AkademikHandler{}, dummy, dummy)
}
