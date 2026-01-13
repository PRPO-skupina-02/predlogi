package api

import (
	"net/http"

	"github.com/PRPO-skupina-02/common/middleware"
	"github.com/PRPO-skupina-02/predlogi/predlogi"
	"github.com/gin-gonic/gin"
)

// TriggerRecommendationJob godoc
//
//	@Summary		Manually trigger recommendation generation job
//	@Description	Triggers the recommendation generation process for all users
//	@Tags			admin
//	@Security		BearerAuth
//	@Success		202	{object}	map[string]string
//	@Failure		401	{object}	middleware.HttpError
//	@Failure		403	{object}	middleware.HttpError
//	@Failure		500	{object}	middleware.HttpError
//	@Router			/api/v1/predlogi/admin/trigger-job [post]
func TriggerRecommendationJob(c *gin.Context) {
	db := middleware.GetContextTransaction(c)

	// Trigger job asynchronously
	go predlogi.RunRecommendationJob(db)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Recommendation job triggered successfully",
		"status":  "processing",
	})
}
