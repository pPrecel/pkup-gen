package token

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

const (
	testCodeBody = `{
	"device_code": "3584d83530557fdd1f46af8289938c8ef79f9dc5",
	"user_code": "WDJB-MJHT",
	"verification_uri": "https://github.com/login/device",
	"expires_in": 900,
	"interval": 5
	  }`
	testAccessTokenBody = `{
		"access_token": "test-token"
	}`
)

type storageMock struct {
	getErr error
	setErr error
	token  string
}

func (ts storageMock) Get(_, _ string) (string, error) {
	return ts.token, ts.getErr
}

func (ts storageMock) Set(_, _, _ string) error {
	return ts.setErr
}

func (ts storageMock) Delete(_, _ string) error {
	return nil
}

func Test_tokenGetter_newTokenFlow(t *testing.T) {
	t.Run("create new token", func(t *testing.T) {
		testServiceName := randString()
		testUsername := randString()
		defer func() { _ = keyring.Delete(testServiceName, testUsername) }()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(200)

			switch r.URL.Path {
			case "/login/device/code":
				fmt.Fprint(w, testCodeBody)
			case "/login/oauth/access_token":
				fmt.Fprint(w, testAccessTokenBody)
			default:
				t.Errorf("unexpected path '%s'", r.URL.Path)
			}
		}))
		defer testServer.Close()

		tg := tokenGetter{
			client:         testServer.Client(),
			logger:         pterm.DefaultLogger.WithWriter(io.Discard),
			serviceName:    testServiceName,
			username:       testUsername,
			githubHostname: testServer.URL,
			clientID:       "testID",
			tokenStorage: &storageMock{
				getErr: errors.New("password not found"),
			},
		}

		token, err := tg.do()
		require.NoError(t, err)
		require.Equal(t, "test-token", token)
	})

	t.Run("/login/device/code error", func(t *testing.T) {
		testServiceName := randString()
		testUsername := randString()
		defer func() { _ = keyring.Delete(testServiceName, testUsername) }()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(404)
		}))
		defer testServer.Close()

		tg := tokenGetter{
			client:         testServer.Client(),
			logger:         pterm.DefaultLogger.WithWriter(io.Discard),
			serviceName:    testServiceName,
			username:       testUsername,
			githubHostname: testServer.URL,
			clientID:       "testID",
			tokenStorage: &storageMock{
				getErr: errors.New("password not found"),
			},
		}

		token, err := tg.do()
		require.Error(t, err)
		require.Empty(t, token)
	})

	t.Run("create new token", func(t *testing.T) {
		testServiceName := randString()
		testUsername := randString()
		defer func() { _ = keyring.Delete(testServiceName, testUsername) }()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/login/device/code":
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(200)
				fmt.Fprint(w, testCodeBody)
			case "/login/oauth/access_token":
				w.WriteHeader(404)
			default:
				t.Errorf("unexpected path '%s'", r.URL.Path)
			}
		}))
		defer testServer.Close()

		tg := tokenGetter{
			client:         testServer.Client(),
			logger:         pterm.DefaultLogger.WithWriter(io.Discard),
			serviceName:    testServiceName,
			username:       testUsername,
			githubHostname: testServer.URL,
			clientID:       "testID",
			tokenStorage: &storageMock{
				getErr: errors.New("password not found"),
			},
		}

		token, err := tg.do()
		require.Error(t, err)
		require.Empty(t, token)
	})
}

func Test_tokenGetter_tokenFromKeyringFlow(t *testing.T) {
	t.Run("get proper token from cache", func(t *testing.T) {
		testServiceName := randString()
		testUsername := randString()
		_ = keyring.Set(testServiceName, testUsername, "test-token")
		defer func() { _ = keyring.Delete(testServiceName, testUsername) }()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)

			switch r.URL.Path {
			case "/octocat":
				require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				fmt.Fprint(w, "test response body")
			default:
				t.Errorf("unexpected path '%s'", r.URL.Path)
			}
		}))
		defer testServer.Close()

		tg := tokenGetter{
			client:            testServer.Client(),
			logger:            pterm.DefaultLogger.WithWriter(io.Discard),
			serviceName:       testServiceName,
			username:          testUsername,
			githubHostname:    testServer.URL,
			githubAPIHostname: testServer.URL,
			clientID:          "testID",
			tokenStorage: &storageMock{
				token: "test-token",
			},
		}

		token, err := tg.do()
		require.NoError(t, err)
		require.Equal(t, "test-token", token)
	})

	t.Run("stored token is not valid and request for new one", func(t *testing.T) {
		testServiceName := randString()
		testUsername := randString()
		_ = keyring.Set(testServiceName, testUsername, "test-token")
		defer func() { _ = keyring.Delete(testServiceName, testUsername) }()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/octocat":
				require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				w.WriteHeader(401)
				fmt.Fprint(w, "test response body")
			case "/login/device/code":
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(200)
				fmt.Fprint(w, testCodeBody)
			case "/login/oauth/access_token":
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(200)
				fmt.Fprint(w, testAccessTokenBody)
			default:
				t.Errorf("unexpected path '%s'", r.URL.Path)
			}
		}))
		defer testServer.Close()

		tg := tokenGetter{
			client:            testServer.Client(),
			logger:            pterm.DefaultLogger.WithWriter(io.Discard),
			serviceName:       testServiceName,
			username:          testUsername,
			githubHostname:    testServer.URL,
			githubAPIHostname: testServer.URL,
			clientID:          "testID",
			tokenStorage: &storageMock{
				token: "test-token",
			},
		}

		token, err := tg.do()
		require.NoError(t, err)
		require.Equal(t, "test-token", token)
	})
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
