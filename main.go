package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/PRPO-skupina-02/common/database"
	"github.com/PRPO-skupina-02/common/logging"
	"github.com/PRPO-skupina-02/common/validation"
	"github.com/PRPO-skupina-02/predlogi/api"
	"github.com/PRPO-skupina-02/predlogi/db"
	"github.com/PRPO-skupina-02/predlogi/predlogi"
	"github.com/gin-gonic/gin"
)

func main() {
	err := run()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run() error {
	slog.Info("Starting predlogi service")

	logger := logging.GetDefaultLogger()
	slog.SetDefault(logger)

	database, err := database.OpenAndMigrateProd(db.MigrationsFS)
	if err != nil {
		return err
	}

	trans, err := validation.RegisterValidation()
	if err != nil {
		return err
	}

	// Setup cron scheduler for recommendation job
	err = predlogi.SetupCron(database)
	if err != nil {
		return err
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api.Register(router, database, trans)

	slog.Info("Server startup complete")
	err = router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}
