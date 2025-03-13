package scheduler

import (
	"context"
	"glpatEye/internal/metrics"
	"log"

	gitlabClient "glpatEye/internal/gitlab"

	"github.com/go-co-op/gocron/v2"
)

func ScheduleTasks(scheduler gocron.Scheduler, ctx context.Context, client *gitlabClient.GitlabClient, cron string, responseObjectsSize, poolSize int) {
	_, _ = scheduler.NewJob(
		gocron.CronJob(
			cron,
			false,
		),
		gocron.NewTask(
			func() {
				log.Println("Scheduler started.")
				metrics.ResetMetrics()
				gitlabClient.CheckMasterToken(ctx, client)
				gitlabClient.ProcessProjects(ctx, client, responseObjectsSize, poolSize)
				gitlabClient.ProcessGroups(ctx, client, responseObjectsSize, poolSize)
				log.Println("Scheduler finished. Metrics were updated.")
			},
		),
	)
}
