package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/sashabaranov/go-openai"
)

type MovieHistory struct {
	Title  string  `json:"title"`
	Rating float64 `json:"rating"`
}

type UpcomingMovie struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
}

type RecommendationRequest struct {
	UserHistory    []MovieHistory  `json:"user_history"`
	UpcomingMovies []UpcomingMovie `json:"upcoming_movies"`
}

type RecommendationResponse struct {
	MovieID         string  `json:"movie_id"`
	MovieTitle      string  `json:"movie_title"`
	Reason          string  `json:"reason"`
	ConfidenceScore float64 `json:"confidence_score"`
}

type OpenAIService struct {
	client    *openai.Client
	model     string
	maxTokens int
}

func NewOpenAIService() (*OpenAIService, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable is required")
	}

	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		return nil, fmt.Errorf("OPENROUTER_MODEL environment variable is required")
	}

	maxTokens := 500
	if mt := os.Getenv("OPENROUTER_MAX_TOKENS"); mt != "" {
		fmt.Sscanf(mt, "%d", &maxTokens)
	}

	// Configure OpenRouter
	config := openai.DefaultConfig(apiKey)
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	slog.Info("OpenRouter service initialized", "model", model, "max_tokens", maxTokens, "base_url", baseURL)

	return &OpenAIService{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}, nil
}

func (s *OpenAIService) GenerateRecommendation(ctx context.Context, req RecommendationRequest) (*RecommendationResponse, error) {
	if len(req.UpcomingMovies) == 0 {
		return nil, fmt.Errorf("no upcoming movies available")
	}

	prompt := s.buildPrompt(req)

	slog.Info("Generating recommendation with OpenAI", "model", s.model)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a movie recommendation assistant for a cinema. Based on user's viewing history and upcoming movies, recommend ONE movie that would best suit this user. Respond ONLY with valid JSON in this exact format: {\"movie_id\": \"<id>\", \"movie_title\": \"<title>\", \"reason\": \"<personalized explanation>\", \"confidence_score\": <0.0-1.0>}. Do not include any other text.",
				},
				{
					Role:    "user",
					Content: prompt,
				},
			},
			MaxTokens:   s.maxTokens,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendation: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no recommendation generated")
	}

	content := resp.Choices[0].Message.Content

	slog.Info("OpenAI response received", "content", content)

	var recommendation RecommendationResponse
	if err := json.Unmarshal([]byte(content), &recommendation); err != nil {
		slog.Error("Failed to parse OpenAI response", "content", content, "error", err)
		return nil, fmt.Errorf("failed to parse recommendation response: %w", err)
	}

	// Validate that the recommended movie is in the upcoming list
	found := false
	for _, movie := range req.UpcomingMovies {
		if movie.ID == recommendation.MovieID {
			found = true
			break
		}
	}

	if !found {
		slog.Warn("OpenAI recommended a movie not in the upcoming list, using first available",
			"recommended_id", recommendation.MovieID)
		// Fallback to first available movie
		recommendation.MovieID = req.UpcomingMovies[0].ID
		recommendation.MovieTitle = req.UpcomingMovies[0].Title
	}

	// Ensure confidence score is between 0 and 1
	if recommendation.ConfidenceScore < 0 {
		recommendation.ConfidenceScore = 0
	}
	if recommendation.ConfidenceScore > 1 {
		recommendation.ConfidenceScore = 1
	}

	return &recommendation, nil
}

func (s *OpenAIService) buildPrompt(req RecommendationRequest) string {
	prompt := "User's viewing history:\n"

	if len(req.UserHistory) == 0 {
		prompt += "No previous viewing history available.\n"
	} else {
		for _, movie := range req.UserHistory {
			prompt += fmt.Sprintf("- %s (Rating: %.1f/10)\n", movie.Title, movie.Rating)
		}
	}

	prompt += "\nUpcoming movies:\n"
	for _, movie := range req.UpcomingMovies {
		prompt += fmt.Sprintf("- ID: %s, Title: %s, Description: %s, Rating: %.1f/10\n",
			movie.ID, movie.Title, movie.Description, movie.Rating)
	}

	prompt += "\nPlease recommend ONE movie from the upcoming list that would best suit this user based on their history. "
	prompt += "Provide a personalized reason for your recommendation."

	return prompt
}
