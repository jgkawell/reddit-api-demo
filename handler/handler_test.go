package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jgkawell/reddit-api-demo/mocks"
	"github.com/jgkawell/reddit-api-demo/models"

	"github.com/steady-bytes/draft/pkg/loggers/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ls1 = models.LinkStats{
		Name:    "ls1",
		UpVotes: 1,
	}
	ls2 = models.LinkStats{
		Name:    "ls2",
		UpVotes: 2,
	}
	ls3 = models.LinkStats{
		Name:    "ls3",
		UpVotes: 3,
	}

	us1 = models.UserStats{
		Name:      "us1",
		PostCount: 1,
	}
	us2 = models.UserStats{
		Name:      "us2",
		PostCount: 2,
	}

	stats1 = models.Stats{
		Posts: []models.LinkStats{
			ls3,
			ls2,
			ls1,
		},
		Users: []models.UserStats{
			us2,
			us1,
		},
	}
	stats2 = models.Stats{
		Posts: []models.LinkStats{
			ls1,
		},
		Users: []models.UserStats{
			us1,
		},
	}
)

func Test_HandlerStatsHandler(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New()
	ctrl := mocks.NewController(t)
	handler := &handler{
		logger,
		ctrl,
	}

	tests := []struct {
		name           string
		rawQuery       string
		expectedStatus int
		expectedStats  models.Stats
		expectedErr    error
	}{
		{
			name:           "happy",
			rawQuery:       "?sub=example&limit=10",
			expectedStatus: http.StatusOK,
			expectedStats:  stats1,
			expectedErr:    nil,
		},
		{
			name:           "warning: bad limit",
			rawQuery:       "?sub=example&limit=abc",
			expectedStatus: http.StatusOK,
			expectedStats:  stats1,
		},
		{
			name:           "bad query params",
			rawQuery:       "?sub=example;limit=10",
			expectedStatus: http.StatusBadRequest,
			expectedStats:  models.Stats{},
			expectedErr:    errors.New("invalid semicolon separator in query"),
		},
		{
			name:           "error",
			rawQuery:       "?sub=example&limit=10",
			expectedStatus: http.StatusInternalServerError,
			expectedStats:  models.Stats{},
			expectedErr:    errors.New("subreddit not configured"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus != http.StatusBadRequest {
				ctrl.On("Stats", ctx, mock.Anything, mock.Anything).Once().Return(tc.expectedStats.Posts, tc.expectedStats.Users, tc.expectedErr)
			}

			req, err := http.NewRequest("GET", "/api/stats", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.URL.RawQuery = tc.rawQuery

			rr := httptest.NewRecorder()
			h := http.HandlerFunc(handler.statsHandler)
			h.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedErr != nil {
				body, err := io.ReadAll(rr.Body)
				if err != nil {
					t.Error("failed to read response body on error")
				}
				assert.Equal(t, tc.expectedErr, errors.New(strings.Trim(string(body), "\n")))
			} else {
				body, err := io.ReadAll(rr.Body)
				if err != nil {
					t.Error("failed to read response body on success")
				}

				stats := models.Stats{}
				err = json.Unmarshal(body, &stats)
				if err != nil {
					t.Error("failed to unmarshal body")
				}

				assert.Equal(t, tc.expectedStats, stats)
			}

		})
	}
}
