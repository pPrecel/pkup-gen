// Code generated by mockery v2.33.3. DO NOT EDIT.

package automock

import (
	github "github.com/pPrecel/PKUP/pkg/github"
	mock "github.com/stretchr/testify/mock"

	v53github "github.com/google/go-github/v53/github"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// GetLatestReleaseOrZero provides a mock function with given fields: _a0, _a1
func (_m *Client) GetLatestReleaseOrZero(_a0 string, _a1 string) (string, error) {
	ret := _m.Called(_a0, _a1)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPRContentDiff provides a mock function with given fields: _a0, _a1, _a2
func (_m *Client) GetPRContentDiff(_a0 *v53github.PullRequest, _a1 string, _a2 string) (string, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(*v53github.PullRequest, string, string) (string, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(*v53github.PullRequest, string, string) string); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(*v53github.PullRequest, string, string) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUserPRsForRepo provides a mock function with given fields: _a0, _a1
func (_m *Client) ListUserPRsForRepo(_a0 github.Options, _a1 []github.FilterFunc) ([]*v53github.PullRequest, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*v53github.PullRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(github.Options, []github.FilterFunc) ([]*v53github.PullRequest, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(github.Options, []github.FilterFunc) []*v53github.PullRequest); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v53github.PullRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(github.Options, []github.FilterFunc) error); ok {
		r1 = rf(_a0, _a1)
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
