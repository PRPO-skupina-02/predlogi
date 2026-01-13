package models

import (
	"time"

	"github.com/PRPO-skupina-02/common/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecommendationStatus string

const (
	StatusPending RecommendationStatus = "pending"
	StatusSent    RecommendationStatus = "sent"
	StatusOpened  RecommendationStatus = "opened"
	StatusClicked RecommendationStatus = "clicked"
	StatusFailed  RecommendationStatus = "failed"
)

type Recommendation struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID  uuid.UUID `gorm:"type:uuid;not null;index"`
	MovieID uuid.UUID `gorm:"type:uuid;not null"`

	Reason          string  `gorm:"type:text"`
	ConfidenceScore float64 `gorm:"type:float;default:0"`

	SentAt    *time.Time `gorm:"index"`
	OpenedAt  *time.Time
	ClickedAt *time.Time

	Status RecommendationStatus `gorm:"type:varchar(50);default:'pending';index"`

	GenerationContext string `gorm:"type:jsonb"` // Store AI context for debugging

	// Email tracking
	EmailTo      string
	EmailSubject string
}

func (r *Recommendation) Create(tx *gorm.DB) error {
	if err := tx.Create(r).Error; err != nil {
		return err
	}
	return nil
}

func (r *Recommendation) Save(tx *gorm.DB) error {
	if err := tx.Save(r).Error; err != nil {
		return err
	}
	return nil
}

func (r *Recommendation) Delete(tx *gorm.DB) error {
	if err := tx.Delete(r).Error; err != nil {
		return err
	}
	return nil
}

func GetRecommendation(tx *gorm.DB, id uuid.UUID) (Recommendation, error) {
	var recommendation Recommendation
	if err := tx.Where("id = ?", id).First(&recommendation).Error; err != nil {
		return recommendation, err
	}
	return recommendation, nil
}

func GetRecommendationsByUser(tx *gorm.DB, userID uuid.UUID, pagination *request.PaginationOptions, sort *request.SortOptions) ([]Recommendation, int64, error) {
	var recommendations []Recommendation
	var total int64

	query := tx.Model(&Recommendation{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return recommendations, 0, err
	}

	if err := query.Scopes(request.PaginateScope(pagination), request.SortScope(sort)).Find(&recommendations).Error; err != nil {
		return recommendations, 0, err
	}

	return recommendations, total, nil
}

func GetRecommendations(tx *gorm.DB, pagination *request.PaginationOptions, sort *request.SortOptions) ([]Recommendation, int64, error) {
	var recommendations []Recommendation
	var total int64

	query := tx.Model(&Recommendation{})

	if err := query.Count(&total).Error; err != nil {
		return recommendations, 0, err
	}

	if err := query.Scopes(request.PaginateScope(pagination), request.SortScope(sort)).Find(&recommendations).Error; err != nil {
		return recommendations, 0, err
	}

	return recommendations, total, nil
}

func MarkRecommendationAsSent(tx *gorm.DB, id uuid.UUID) error {
	now := time.Now()
	return tx.Model(&Recommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  StatusSent,
		"sent_at": now,
	}).Error
}

func MarkRecommendationAsFailed(tx *gorm.DB, id uuid.UUID) error {
	return tx.Model(&Recommendation{}).Where("id = ?", id).Update("status", StatusFailed).Error
}
