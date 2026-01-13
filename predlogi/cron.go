package predlogi

import (
	"log/slog"
	"os"

	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

func SetupCron(db *gorm.DB) error {
	schedule := os.Getenv("RECOMMENDATION_SCHEDULE")
	if schedule == "" {
		// Monday and Thursday at 9:00 AM
		schedule = "0 9 * * 1,4"
	}

	slog.Info("Setting up cron scheduler", "schedule", schedule)

	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	// Run job on startup
	_, err = s.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
		gocron.NewTask(RunRecommendationJob, db),
	)
	if err != nil {
		return err
	}

	// Schedule recurring job
	j, err := s.NewJob(
		gocron.CronJob(schedule, false),
		gocron.NewTask(RunRecommendationJob, db),
	)
	if err != nil {
		return err
	}

	s.Start()

	slog.Info("Cron job started", "id", j.ID(), "schedule", schedule)
	return nil
}
