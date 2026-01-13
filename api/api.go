package api

import (
	"net/http"

	"github.com/PRPO-skupina-02/common/middleware"
	_ "github.com/PRPO-skupina-02/predlogi/api/docs"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// @title						Predlogi API
// @version					1.0
// @description				Movie recommendation service for the PRPO project
// @host						localhost:8080
// @BasePath					/api/v1/predlogi
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func Register(router *gin.Engine, db *gorm.DB, trans ut.Translator) {
	// Healthcheck
	router.GET("/healthcheck", healthcheck)

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Admin API
	admin := router.Group("/api/v1/predlogi/admin")
	admin.Use(middleware.TransactionMiddleware(db))
	admin.Use(middleware.TranslationMiddleware(trans))
	admin.Use(middleware.ErrorMiddleware)
	admin.Use(middleware.RequireAdmin())
	admin.POST("/trigger-job", TriggerRecommendationJob)
}

func healthcheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
