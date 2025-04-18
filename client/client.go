package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jgkawell/reddit-api-demo/models"

	"github.com/steady-bytes/draft/pkg/chassis"
	"golang.org/x/time/rate"
)

type (
	// Client wraps an http.Client to provide access to the Reddit API with some additional helpers built
	// in. Included are:
	// - Bearer authorization using the configured access token
	// - Applies a custom User-Agent (required by Reddit API terms)
	// - Sets the Accept type to 'application/json'
	// - Abides by the current rate limit imposed by the Reddit API
	Client interface {
		// Get will send a GET request to the provided url with the given values as url params.
		Get(ctx context.Context, url string, values url.Values) (response *http.Response, err error)
		// GetLinkListing wraps Get() and automatically unmarshals the result into a Listing type with
		// a Children type of Link.
		GetLinkListing(ctx context.Context, url string, values url.Values) (listing models.Listing, err error)
	}
	client struct {
		logger      chassis.Logger
		token       string
		latestLimit time.Time
		limiter     *rate.Limiter
	}
)

const (
	limitPercentage = 0.9
	maxRate         = 2
)

func NewClient(logger chassis.Logger) Client {
	token := chassis.GetConfig().GetString("reddit.accessToken")
	if token == "" {
		logger.Panic("no token provided in config")
	}

	return &client{
		logger:      logger,
		token:       token,
		latestLimit: time.Now(),
		limiter:     rate.NewLimiter(100/60, 1),
	}
}

func (c *client) Get(ctx context.Context, url string, values url.Values) (response *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "golang:reddit-api-demo:v0.0.1 (by /u/jgkawell)")
	// abide by the current rate limit
	err = c.limiter.Wait(ctx)
	if err != nil {
		return
	}
	c.logger.Trace("calling Reddit API")
	response, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	go c.setRateLimit(response.Header)
	return
}

func (c *client) GetLinkListing(ctx context.Context, url string, values url.Values) (listing models.Listing, err error) {
	resp, err := c.Get(ctx, url, values)
	if err != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &listing)
	if err != nil {
		return
	}

	return
}

func (c *client) setRateLimit(header http.Header) {
	date, err := time.Parse(time.RFC1123, header.Get("Date"))
	if err != nil {
		c.logger.WithError(err).Error("failed to parse date from header")
		return
	}

	// if this date is before the latest event we can ignore it
	if date.Before(c.latestLimit) {
		c.logger.Info("ignoring older rate limit event")
		return
	}

	reset, err := strconv.Atoi(header.Get("X-RateLimit-Reset"))
	if err != nil {
		c.logger.WithError(err).Error("failed to parse reset from header")
		return
	}
	remaining, err := strconv.ParseFloat(header.Get("X-RateLimit-Remaining"), 64)
	if err != nil {
		c.logger.WithError(err).Error("failed to parse remaining from header")
		return
	}

	// new limit is the remaining requests divided by the reset time (in seconds) multiplied by the
	// configured percentage limit (with a configured maximum)
	c.limiter.SetLimit(rate.Limit(math.Min(maxRate, limitPercentage*remaining/float64(reset))))
}
