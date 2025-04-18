package controller

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"slices"
	"sync"

	"github.com/jgkawell/reddit-api-demo/client"
	"github.com/jgkawell/reddit-api-demo/models"

	"github.com/steady-bytes/draft/pkg/chassis"
)

type (
	// Processor collects stats for the configured subreddit and provides access to the
	// stats in near real time.
	Processor interface {
		// Start begins stat collection and is meant to be run on a background routine.
		Start()
		// Stats will return the current top
		Stats(ctx context.Context, limit int) (links []models.LinkStats, users []models.UserStats, err error)
	}
	processor struct {
		logger chassis.Logger
		client client.Client
		config subredditConfig

		linksMu sync.RWMutex
		links   map[string]models.Link

		usersMu sync.RWMutex
		users   map[string]user
	}
	user struct {
		name  string
		links map[string]models.Link
	}
)

func NewProcessor(logger chassis.Logger, client client.Client, config subredditConfig) Processor {
	return &processor{
		logger: logger.WithField("subreddit", config.Name),
		client: client,
		config: config,

		linksMu: sync.RWMutex{},
		links:   make(map[string]models.Link),
		usersMu: sync.RWMutex{},
		users:   make(map[string]user),
	}
}

func (p *processor) Stats(ctx context.Context, limit int) (links []models.LinkStats, users []models.UserStats, err error) {
	links = []models.LinkStats{}
	users = []models.UserStats{}

	// collect link stats
	p.linksMu.RLock()
	for _, l := range p.links {
		links = append(links, models.LinkStats{
			Name:    l.Data.Name,
			Title:   l.Data.Title,
			Author:  l.Data.Author,
			UpVotes: l.Data.Ups,
		})
	}
	p.linksMu.RUnlock()
	slices.SortFunc(links, func(a, b models.LinkStats) int {
		return b.UpVotes - a.UpVotes
	})

	// collect user stats
	p.usersMu.RLock()
	for _, u := range p.users {
		users = append(users, models.UserStats{
			Name:      u.name,
			PostCount: len(u.links),
		})
	}
	p.usersMu.RUnlock()
	slices.SortFunc(users, func(a, b models.UserStats) int {
		return b.PostCount - a.PostCount
	})

	// apply limit if needed
	if len(links) > limit {
		links = links[0:limit]
	}
	if len(users) > limit {
		users = users[0:limit]
	}

	return
}

func (p *processor) Start() {
	var (
		ctx = context.Background()
		err error
	)

	if p.config.Start == "" {
		p.config.Start, err = p.init(ctx)
		if err != nil {
			p.logger.WithError(err).Error("failed to find starting link")
			os.Exit(1)
		}
		p.logger.WithField("start", p.config.Start).Info("found starting link")
	} else {
		p.logger.WithField("start", p.config.Start).Info("using configured starting link")
	}

	// run stat collection forever as quickly as the rate limit of the Client will allow
	for {
		p.process(ctx)
	}

}

// init gets the latest link to register where to begin data collection
func (p *processor) init(ctx context.Context) (start string, err error) {
	p.logger.Info("initializing")
	values := url.Values{
		"limit": {"1"},
	}
	listing, err := p.client.GetLinkListing(ctx, subredditURL(p.config.Name, "new"), values)
	if err != nil {
		return
	}
	if len(listing.Data.Children) == 0 {
		return
	}
	return listing.Data.Children[0].Data.Name, nil
}

// process lists the latest links and then processes for new data
func (p *processor) process(ctx context.Context) {
	links, err := p.listLinks(ctx, p.config.Start)
	if err != nil {
		p.logger.WithError(err).Error("failed to list links")
		// just return so that process() will be called again
		return
	}

	// process results concurrently
	for _, link := range links {
		go p.processLink(ctx, link)
		go p.processUser(ctx, link)
	}
}

// listLinks queries the API for the configured subreddit's latest links since the
// provided "before" link.
func (p *processor) listLinks(ctx context.Context, before string) ([]models.Link, error) {
	p.logger.WithField("before", before).Debug("listing links")
	values := url.Values{
		"limit":  {"100"},
		"before": {before},
	}
	listing, err := p.client.GetLinkListing(ctx, subredditURL(p.config.Name, "new"), values)
	if err != nil {
		return nil, err
	}
	return listing.Data.Children, nil
}

// NOTE: this could be converted to a database call
func (p *processor) processLink(_ context.Context, link models.Link) {
	p.linksMu.Lock()
	p.links[link.Data.Name] = link
	p.linksMu.Unlock()
}

// NOTE: this could be converted to a database call
func (p *processor) processUser(_ context.Context, link models.Link) {
	// add link to user (creating user if neccesary)
	p.usersMu.Lock()
	u, ok := p.users[link.Data.AuthorFullname]
	if !ok {
		p.logger.WithField("user", link.Data.Author).Debug("saving user")
		u = user{
			name:  link.Data.Author,
			links: map[string]models.Link{},
		}
	}
	u.links[link.Data.Name] = link
	p.users[link.Data.AuthorFullname] = u
	p.usersMu.Unlock()
}

func subredditURL(subreddit string, sort string) string {
	return fmt.Sprintf("https://oauth.reddit.com/r/%s/%s", subreddit, sort)
}
