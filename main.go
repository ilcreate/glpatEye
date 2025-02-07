package main

import (
	"context"
	"glpatEye/internal/config"
	gitlabClient "glpatEye/internal/gitlab"
	"glpatEye/internal/metrics"
	"glpatEye/internal/scheduler"
	"log"
	"strconv"

	"github.com/go-co-op/gocron/v2"
)

func main() {
	cfg := initConfig()
	client := initGitlabClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cron := cfg.DefaultConfig("CRON", cfg.Gitlab.Cron).(string)
	port := cfg.DefaultConfig("SERVER_PORT", strconv.Itoa(cfg.Server.Port)).(string)
	pool := cfg.DefaultConfig("POOL_SIZE", cfg.Gitlab.PoolSize).(int)
	objectsPerPage := cfg.DefaultConfig("OBJECTS_PER_PAGE", cfg.Gitlab.ResponseObjectSize).(int)

	s, _ := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()
	scheduler.ScheduleTasks(s, ctx, client, cron, objectsPerPage, pool)
	s.Start()

	log.Println("Metrics exposed to endpoint '/metrics'")
	metrics.InitAppMetrics(port)
}

func initConfig() *config.Config {
	cfg := &config.Config{}
	cfg.InitAppConfig("./configs/config.yaml")
	return cfg
}

func initGitlabClient(cfg *config.Config) *gitlabClient.GitlabClient {
	token := cfg.DefaultConfig("GITLAB_TOKEN", "").(string)
	baseUrl := cfg.DefaultConfig("GITLAB_URL", cfg.Gitlab.BaseUrl).(string)
	pattern := cfg.DefaultConfig("GITLAB_PATTERN", cfg.Gitlab.Pattern).(string)
	cron := cfg.DefaultConfig("CRON", cfg.Gitlab.Cron).(string)

	client, err := gitlabClient.NewGitlabClient(token, baseUrl, pattern, cron)
	if err != nil {
		log.Fatalf("Error to initialize Gitlab client: %v", err)
	}
	return client
}
