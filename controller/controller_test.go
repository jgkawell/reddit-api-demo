package controller

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jgkawell/reddit-api-demo/mocks"
	"github.com/jgkawell/reddit-api-demo/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/steady-bytes/draft/pkg/loggers/zerolog"
)

var (
	l1 = models.Link{
		Data: models.LinkData{
			Name: "l1",
			Ups:  1,
		},
	}
	l2 = models.Link{
		Data: models.LinkData{
			Name: "l2",
			Ups:  2,
		},
	}
	l3 = models.Link{
		Data: models.LinkData{
			Name: "l3",
			Ups:  3,
		},
	}

	u1 = user{
		name: "u1",
		links: map[string]models.Link{
			l1.Data.Name: l1,
		},
	}
	u2 = user{
		name: "u2",
		links: map[string]models.Link{
			l2.Data.Name: l2,
			l3.Data.Name: l3,
		},
	}
)

func Test_Processor(t *testing.T) {
	logger := zerolog.New()
	client := mocks.NewClient(t)
	config := subredditConfig{}
	proc := &processor{
		logger: logger,
		client: client,
		config: config,

		linksMu: sync.RWMutex{},
		links:   make(map[string]models.Link),
		usersMu: sync.RWMutex{},
		users:   make(map[string]user),
	}

	tests := []struct {
		name          string
		processor     *processor
		limit         int
		links         map[string]models.Link
		users         map[string]user
		expectedLinks []models.LinkStats
		expectedUsers []models.UserStats
		expectedErr   error
	}{
		{
			name:          "no data",
			processor:     proc,
			limit:         5,
			links:         map[string]models.Link{},
			users:         map[string]user{},
			expectedLinks: []models.LinkStats{},
			expectedUsers: []models.UserStats{},
			expectedErr:   nil,
		},
		{
			name:      "happy",
			processor: proc,
			limit:     5,
			links: map[string]models.Link{
				l1.Data.Name: l1,
				l2.Data.Name: l2,
				l3.Data.Name: l3,
			},
			users: map[string]user{
				u1.name: u1,
				u2.name: u2,
			},
			expectedLinks: []models.LinkStats{
				{
					Name:    l3.Data.Name,
					UpVotes: l3.Data.Ups,
				},
				{
					Name:    l2.Data.Name,
					UpVotes: l2.Data.Ups,
				},
				{
					Name:    l1.Data.Name,
					UpVotes: l1.Data.Ups,
				},
			},
			expectedUsers: []models.UserStats{
				{
					Name:      u2.name,
					PostCount: 2,
				},
				{
					Name:      u1.name,
					PostCount: 1,
				},
			},
			expectedErr: nil,
		},
		{
			name:      "one of each",
			processor: proc,
			limit:     1,
			links: map[string]models.Link{
				l1.Data.Name: l1,
			},
			users: map[string]user{
				u1.name: u1,
			},
			expectedLinks: []models.LinkStats{
				{
					Name:    l1.Data.Name,
					UpVotes: l1.Data.Ups,
				},
			},
			expectedUsers: []models.UserStats{
				{
					Name:      u1.name,
					PostCount: 1,
				},
			},
			expectedErr: nil,
		},
		{
			name:      "less than limit",
			processor: proc,
			limit:     5,
			links: map[string]models.Link{
				l1.Data.Name: l1,
			},
			users: map[string]user{
				u1.name: u1,
			},
			expectedLinks: []models.LinkStats{
				{
					Name:    l1.Data.Name,
					UpVotes: l1.Data.Ups,
				},
			},
			expectedUsers: []models.UserStats{
				{
					Name:      u1.name,
					PostCount: 1,
				},
			},
			expectedErr: nil,
		},
		{
			name:      "more than limit",
			processor: proc,
			limit:     1,
			links: map[string]models.Link{
				l1.Data.Name: l1,
				l2.Data.Name: l2,
				l3.Data.Name: l3,
			},
			users: map[string]user{
				u1.name: u1,
				u2.name: u2,
			},
			expectedLinks: []models.LinkStats{
				{
					Name:    l3.Data.Name,
					UpVotes: l3.Data.Ups,
				},
			},
			expectedUsers: []models.UserStats{
				{
					Name:      u2.name,
					PostCount: 2,
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.processor.links = tc.links
			tc.processor.users = tc.users
			links, users, err := tc.processor.Stats(ctx, tc.limit)

			assert.Equal(t, links, tc.expectedLinks)
			assert.Equal(t, users, tc.expectedUsers)
			assert.Equal(t, err, tc.expectedErr)

		})
	}
}

func Test_ProcessorInit(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New()
	client := mocks.NewClient(t)
	config := subredditConfig{
		Name: "test",
	}
	proc := &processor{
		logger: logger,
		client: client,
		config: config,
	}

	tests := []struct {
		name            string
		processor       *processor
		listingResponse models.Listing
		expectedStart   string
		expectedErr     error
	}{
		{
			name:      "happy",
			processor: proc,
			listingResponse: models.Listing{
				Data: models.ListingData{
					Children: []models.Link{
						l1,
					},
				},
			},
			expectedStart: "l1",
			expectedErr:   nil,
		},
		{
			name:            "error",
			processor:       proc,
			listingResponse: models.Listing{},
			expectedStart:   "",
			expectedErr:     errors.New("failed to call api"),
		},
		{
			name:      "no data",
			processor: proc,
			listingResponse: models.Listing{
				Data: models.ListingData{
					Children: []models.Link{},
				},
			},
			expectedStart: "",
			expectedErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client.On("GetLinkListing", ctx, mock.Anything, mock.Anything).Once().Return(tc.listingResponse, tc.expectedErr)

			start, err := tc.processor.init(ctx)

			assert.Equal(t, start, tc.expectedStart)
			assert.Equal(t, err, tc.expectedErr)

		})
	}
}

func Test_ProcessorProcess(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New()
	client := mocks.NewClient(t)
	config := subredditConfig{
		Name: "test",
		Start: "example",
	}
	proc := &processor{
		logger: logger,
		client: client,
		config: config,

		linksMu: sync.RWMutex{},
		links:   make(map[string]models.Link),
		usersMu: sync.RWMutex{},
		users:   make(map[string]user),
	}

	tests := []struct {
		name            string
		processor       *processor
		before          string
		listingResponse models.Listing
		expectedStart   string
		expectedErr     error
	}{
		{
			name:      "happy",
			processor: proc,
			listingResponse: models.Listing{
				Data: models.ListingData{
					Children: []models.Link{
						l1,
					},
				},
			},
			expectedErr:   nil,
		},
		{
			name:      "more data",
			processor: proc,
			listingResponse: models.Listing{
				Data: models.ListingData{
					Children: []models.Link{
						l1,
						l2,
						l3,
					},
				},
			},
			expectedErr:   nil,
		},
		{
			name:      "duplicate data",
			processor: proc,
			listingResponse: models.Listing{
				Data: models.ListingData{
					Children: []models.Link{
						l1,
						l2,
						l3,
						l1,
						l2,
						l3,
					},
				},
			},
			expectedErr:   nil,
		},
		{
			name:            "error",
			processor:       proc,
			listingResponse: models.Listing{},
			expectedErr:     errors.New("failed to call api"),
		},
		{
			name:      "no data",
			processor: proc,
			listingResponse: models.Listing{
				Data: models.ListingData{
					Children: []models.Link{},
				},
			},
			expectedErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client.On("GetLinkListing", ctx, mock.Anything, mock.Anything).Once().Return(tc.listingResponse, tc.expectedErr)

			tc.processor.process(ctx)
		})
	}
}
