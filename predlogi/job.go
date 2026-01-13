package predlogi

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/PRPO-skupina-02/predlogi/clients/auth"
	"github.com/PRPO-skupina-02/predlogi/clients/nakup"
	"github.com/PRPO-skupina-02/predlogi/clients/spored"
	"github.com/PRPO-skupina-02/predlogi/services"
	"gorm.io/gorm"
)

func RunRecommendationJob(db *gorm.DB) {
	slog.Info("Starting recommendation generation job")
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Initialize clients
	authHost := os.Getenv("AUTH_HOST")
	nakupHost := os.Getenv("NAKUP_HOST")
	sporedHost := os.Getenv("SPORED_HOST")
	rabbitmqURL := os.Getenv("RABBITMQ_URL")

	authClient := auth.NewClient(authHost)
	nakupClient := nakup.NewClient(nakupHost)
	sporedClient := spored.NewClient(sporedHost)

	// Initialize OpenAI service
	openaiService, err := services.NewOpenAIService()
	if err != nil {
		slog.Error("Failed to initialize OpenAI service", "error", err)
		return
	}

	// Initialize recommendation generator
	generator, err := services.NewRecommendationGenerator(
		db,
		authClient,
		nakupClient,
		sporedClient,
		openaiService,
		rabbitmqURL,
	)
	if err != nil {
		slog.Error("Failed to initialize recommendation generator", "error", err)
		return
	}
	defer generator.Close()

	// Generate recommendations for all users
	if err := generator.GenerateForAllUsers(ctx); err != nil {
		slog.Error("Recommendation job failed", "error", err, "duration", time.Since(startTime))
		return
	}

	slog.Info("Recommendation job completed successfully", "duration", time.Since(startTime))
}
