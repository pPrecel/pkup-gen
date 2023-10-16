// Code generated by mockery v2.33.3. DO NOT EDIT.

package automock

import (
	github "github.com/google/go-github/v53/github"
	mock "github.com/stretchr/testify/mock"

	pkggithub "github.com/pPrecel/PKUP/pkg/github"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// GetFileDiffForPRs provides a mock function with given fields: _a0, _a1, _a2
func (_m *Client) GetFileDiffForPRs(_a0 []*github.PullRequest, _a1 string, _a2 string) (string, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func([]*github.PullRequest, string, string) (string, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func([]*github.PullRequest, string, string) string); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func([]*github.PullRequest, string, string) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUserPRsForRepo provides a mock function with given fields: _a0, _a1
func (_m *Client) ListUserPRsForRepo(_a0 pkggithub.Options, _a1 []pkggithub.FilterFunc) ([]*github.PullRequest, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*github.PullRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(pkggithub.Options, []pkggithub.FilterFunc) ([]*github.PullRequest, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(pkggithub.Options, []pkggithub.FilterFunc) []*github.PullRequest); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.PullRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(pkggithub.Options, []pkggithub.FilterFunc) error); ok {
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