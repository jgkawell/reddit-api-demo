// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"

	models "github.com/jgkawell/reddit-api-demo/models"

	url "net/url"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, _a1, values
func (_m *Client) Get(ctx context.Context, _a1 string, values url.Values) (*http.Response, error) {
	ret := _m.Called(ctx, _a1, values)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, url.Values) (*http.Response, error)); ok {
		return rf(ctx, _a1, values)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, url.Values) *http.Response); ok {
		r0 = rf(ctx, _a1, values)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, url.Values) error); ok {
		r1 = rf(ctx, _a1, values)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLinkListing provides a mock function with given fields: ctx, _a1, values
func (_m *Client) GetLinkListing(ctx context.Context, _a1 string, values url.Values) (models.Listing, error) {
	ret := _m.Called(ctx, _a1, values)

	if len(ret) == 0 {
		panic("no return value specified for GetLinkListing")
	}

	var r0 models.Listing
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, url.Values) (models.Listing, error)); ok {
		return rf(ctx, _a1, values)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, url.Values) models.Listing); ok {
		r0 = rf(ctx, _a1, values)
	} else {
		r0 = ret.Get(0).(models.Listing)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, url.Values) error); ok {
		r1 = rf(ctx, _a1, values)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
