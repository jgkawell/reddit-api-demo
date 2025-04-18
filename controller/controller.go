package controller

import (
	"context"
	"errors"
	"os"

	"github.com/jgkawell/reddit-api-demo/client"
	"github.com/jgkawell/reddit-api-demo/models"
	"github.com/steady-bytes/draft/pkg/chassis"
)

type (
	// Controller provides an interface for running and accessing multiple Processors. Each
	// Processor is assigned the same Client so that they all abide by the same rate limit since
	// the application uses a single token restricted to the same limits.
	Controller interface {
		// Stats will return the current stats for the given subreddit.
		Stats(ctx context.Context, subreddit string, limit int) (links []models.LinkStats, users []models.UserStats, err error)
		// Start will read the configured subreddits and start a single Processor for each.
		Start()
	}
	controller struct {
		logger     chassis.Logger
		processors map[string]Processor
	}
	subredditConfig struct {
		Name  string
		Start string
	}
)

func NewController(logger chassis.Logger) Controller {
	return &controller{
		logger:     logger,
		processors: map[string]Processor{},
	}
}

func (c *controller) Stats(ctx context.Context, subreddit string, limit int) (links []models.LinkStats, users []models.UserStats, err error) {
	p, ok := c.processors[subreddit]
	if !ok {
		return nil, nil, errors.New("subreddit not configured")
	}
	return p.Stats(ctx, limit)
}

func (c *controller) Start() {
	config := []subredditConfig{}
	err := chassis.GetConfig().UnmarshalKey("reddit.subreddits", &config)
	if err != nil {
		c.logger.WithError(err).Error("failed to read subreddit config")
		os.Exit(1)
	}

	client := client.NewClient(c.logger)
	for _, subreddit := range config {
		p := NewProcessor(c.logger, client, subreddit)
		go p.Start()
		c.processors[subreddit.Name] = p
	}
}
