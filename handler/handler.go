package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jgkawell/reddit-api-demo/controller"
	"github.com/jgkawell/reddit-api-demo/models"

	"github.com/steady-bytes/draft/pkg/chassis"
)

type (
	// Handler implements the chassis RPCRegistrar interface so its lifecycle can be
	// managed automatically by the chassis. On the network it will expose the /api/stats
	// path for returning stats to the caller.
	Handler interface {
		chassis.RPCRegistrar
	}

	handler struct {
		logger     chassis.Logger
		controller controller.Controller
	}
)

func NewHandler(logger chassis.Logger, ctrl controller.Controller) Handler {
	return &handler{
		logger,
		ctrl,
	}
}

func (h *handler) RegisterRPC(server chassis.Rpcer) {
	server.AddHandler("/api/stats", http.HandlerFunc(h.statsHandler), false)
}

// params:
//   - sub <string>: the subreddit to return the stats for
//   - limit <int>: the limit of posts and users to return (optional)
// returns:
//   - models.Stats{}
func (h *handler) statsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		h.logger.WithError(err).Error("failed to parse query params")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		h.logger.WithError(err).Warn("failed to parse limit param")
		limit = 5
	}

	links, users, err := h.controller.Stats(ctx, params.Get("sub"), limit)
	if err != nil {
		h.logger.WithError(err).Error("failed to collect stats")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Stats{
		Posts: links,
		Users: users,
	})
}
