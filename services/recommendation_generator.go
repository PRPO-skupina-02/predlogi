package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/PRPO-skupina-02/common/messaging"
	"github.com/PRPO-skupina-02/predlogi/clients/auth"
	"github.com/PRPO-skupina-02/predlogi/clients/nakup"
	"github.com/PRPO-skupina-02/predlogi/clients/spored"
	"github.com/PRPO-skupina-02/predlogi/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecommendationGenerator struct {
	db            *gorm.DB
	authClient    *auth.Client
	nakupClient   *nakup.Client
	sporedClient  *spored.Client
	openaiService *OpenAIService
	publisher     *messaging.Publisher
	lookaheadDays int
}

func NewRecommendationGenerator(
	db *gorm.DB,
	authClient *auth.Client,
	nakupClient *nakup.Client,
	sporedClient *spored.Client,
	openaiService *OpenAIService,
	rabbitmqURL string,
) (*RecommendationGenerator, error) {
	publisher, err := messaging.NewPublisher(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	lookaheadDays := 7
	if ld := os.Getenv("RECOMMENDATION_LOOKAHEAD_DAYS"); ld != "" {
		if parsed, err := strconv.Atoi(ld); err == nil {
			lookaheadDays = parsed
		}
	}

	return &RecommendationGenerator{
		db:            db,
		authClient:    authClient,
		nakupClient:   nakupClient,
		sporedClient:  sporedClient,
		openaiService: openaiService,
		publisher:     publisher,
		lookaheadDays: lookaheadDays,
	}, nil
}

func (rg *RecommendationGenerator) Close() error {
	if rg.publisher != nil {
		return rg.publisher.Close()
	}
	return nil
}

func (rg *RecommendationGenerator) GenerateForUser(ctx context.Context, user *auth.User) error {
	slog.Info("Generating recommendation for user", "user_id", user.ID, "email", user.Email)

	// 1. Fetch user's reservation history
	reservations, err := rg.nakupClient.GetUserReservations(user.ID)
	if err != nil {
		slog.Error("Failed to fetch user reservations", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to fetch reservations: %w", err)
	}

	slog.Info("Fetched reservations", "user_id", user.ID, "count", len(reservations))

	// 2. Extract unique movie IDs and fetch movie details
	movieMap := make(map[uuid.UUID]*spored.Movie)
	var userHistory []MovieHistory

	for _, reservation := range reservations {
		// Fetch timeslot to get movie ID
		timeSlot, err := rg.sporedClient.GetTimeSlot(reservation.TimeSlotID)
		if err != nil {
			slog.Warn("Failed to fetch timeslot", "timeslot_id", reservation.TimeSlotID, "error", err)
			continue
		}

		// Skip if we already have this movie
		if _, exists := movieMap[timeSlot.MovieID]; exists {
			continue
		}

		movieMap[timeSlot.MovieID] = &timeSlot.Movie
		userHistory = append(userHistory, MovieHistory{
			Title:  timeSlot.Movie.Title,
			Rating: timeSlot.Movie.Rating,
		})
	}

	slog.Info("Extracted user history", "user_id", user.ID, "unique_movies", len(userHistory))

	// 3. Fetch upcoming schedule
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, rg.lookaheadDays)

	upcomingTimeSlots, err := rg.sporedClient.GetUpcomingTimeSlots(startDate, endDate)
	if err != nil {
		slog.Error("Failed to fetch upcoming schedule", "error", err)
		return fmt.Errorf("failed to fetch upcoming schedule: %w", err)
	}

	slog.Info("Fetched upcoming timeslots", "count", len(upcomingTimeSlots))

	// 4. Extract unique upcoming movies
	upcomingMoviesMap := make(map[uuid.UUID]*spored.Movie)
	var upcomingMovies []UpcomingMovie

	for _, timeSlot := range upcomingTimeSlots {
		if _, exists := upcomingMoviesMap[timeSlot.MovieID]; !exists {
			upcomingMoviesMap[timeSlot.MovieID] = &timeSlot.Movie
			upcomingMovies = append(upcomingMovies, UpcomingMovie{
				ID:          timeSlot.MovieID.String(),
				Title:       timeSlot.Movie.Title,
				Description: timeSlot.Movie.Description,
				Rating:      timeSlot.Movie.Rating,
			})
		}
	}

	if len(upcomingMovies) == 0 {
		slog.Warn("No upcoming movies available", "user_id", user.ID)
		return fmt.Errorf("no upcoming movies available")
	}

	slog.Info("Extracted upcoming movies", "count", len(upcomingMovies))

	// 5. Generate recommendation using OpenAI
	aiReq := RecommendationRequest{
		UserHistory:    userHistory,
		UpcomingMovies: upcomingMovies,
	}

	aiResp, err := rg.openaiService.GenerateRecommendation(ctx, aiReq)
	if err != nil {
		slog.Error("Failed to generate AI recommendation", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to generate AI recommendation: %w", err)
	}

	slog.Info("AI recommendation generated",
		"user_id", user.ID,
		"movie_id", aiResp.MovieID,
		"confidence", aiResp.ConfidenceScore)

	// 6. Parse movie ID
	movieID, err := uuid.Parse(aiResp.MovieID)
	if err != nil {
		slog.Error("Failed to parse movie ID", "movie_id", aiResp.MovieID, "error", err)
		return fmt.Errorf("failed to parse movie ID: %w", err)
	}

	// Get full movie details
	recommendedMovie, err := rg.sporedClient.GetMovie(movieID)
	if err != nil {
		slog.Error("Failed to fetch recommended movie", "movie_id", movieID, "error", err)
		return fmt.Errorf("failed to fetch recommended movie: %w", err)
	}

	// 7. Store recommendation in database
	contextJSON, _ := json.Marshal(map[string]interface{}{
		"user_history":    userHistory,
		"upcoming_movies": upcomingMovies,
		"ai_response":     aiResp,
	})

	recommendation := models.Recommendation{
		UserID:            user.ID,
		MovieID:           movieID,
		Reason:            aiResp.Reason,
		ConfidenceScore:   aiResp.ConfidenceScore,
		Status:            models.StatusPending,
		GenerationContext: string(contextJSON),
		EmailTo:           user.Email,
		EmailSubject:      fmt.Sprintf("Perfect Movie for You: %s", recommendedMovie.Title),
	}

	if err := recommendation.Create(rg.db); err != nil {
		slog.Error("Failed to save recommendation", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to save recommendation: %w", err)
	}

	slog.Info("Recommendation saved", "recommendation_id", recommendation.ID)

	// 8. Send email notification via RabbitMQ
	reservationURL := fmt.Sprintf("https://cinema.example.com/reserve?movie=%s", movieID.String())

	emailMsg := messaging.NewEmailMessage(
		user.Email,
		"recommendation",
		map[string]interface{}{
			"UserName":             user.FirstName,
			"MovieTitle":           recommendedMovie.Title,
			"MovieDescription":     recommendedMovie.Description,
			"MovieRating":          fmt.Sprintf("%.1f/10", recommendedMovie.Rating),
			"RecommendationReason": aiResp.Reason,
			"ReservationURL":       reservationURL,
			"ImageURL":             recommendedMovie.ImageURL,
		},
	)

	if err := rg.publisher.PublishEmail(ctx, emailMsg); err != nil {
		slog.Error("Failed to publish email", "user_id", user.ID, "error", err)
		// Mark as failed but don't return error - recommendation is still saved
		_ = models.MarkRecommendationAsFailed(rg.db, recommendation.ID)
		return fmt.Errorf("failed to publish email: %w", err)
	}

	// Mark as sent
	if err := models.MarkRecommendationAsSent(rg.db, recommendation.ID); err != nil {
		slog.Warn("Failed to mark recommendation as sent", "recommendation_id", recommendation.ID, "error", err)
	}

	slog.Info("Recommendation sent successfully",
		"user_id", user.ID,
		"recommendation_id", recommendation.ID,
		"email", user.Email)

	return nil
}

func (rg *RecommendationGenerator) GenerateForAllUsers(ctx context.Context) error {
	slog.Info("Starting recommendation generation for all users")

	users, err := rg.authClient.GetActiveUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch active users: %w", err)
	}

	slog.Info("Fetched active users", "count", len(users))

	successCount := 0
	failureCount := 0

	for i, user := range users {
		slog.Info("Processing user", "index", i+1, "total", len(users), "user_id", user.ID, "email", user.Email)

		if err := rg.GenerateForUser(ctx, &user); err != nil {
			slog.Error("Failed to generate recommendation for user", "user_id", user.ID, "error", err)
			failureCount++
			continue
		}

		successCount++
	}

	slog.Info("Recommendation generation completed",
		"total_users", len(users),
		"success", successCount,
		"failure", failureCount)

	return nil
}
