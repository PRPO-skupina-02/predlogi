package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRecommendationCreation(t *testing.T) {
	recommendation := Recommendation{
		UserID:          uuid.New(),
		MovieID:         uuid.New(),
		Reason:          "Test recommendation",
		ConfidenceScore: 0.85,
		Status:          StatusPending,
		EmailTo:         "test@example.com",
		EmailSubject:    "Test Subject",
	}

	assert.NotEqual(t, uuid.Nil, recommendation.UserID)
	assert.NotEqual(t, uuid.Nil, recommendation.MovieID)
	assert.Equal(t, StatusPending, recommendation.Status)
	assert.Equal(t, 0.85, recommendation.ConfidenceScore)
}

func TestRecommendationStatus(t *testing.T) {
	tests := []struct {
		name   string
		status RecommendationStatus
	}{
		{"Pending", StatusPending},
		{"Sent", StatusSent},
		{"Opened", StatusOpened},
		{"Clicked", StatusClicked},
		{"Failed", StatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := Recommendation{Status: tt.status}
			assert.Equal(t, tt.status, rec.Status)
		})
	}
}
