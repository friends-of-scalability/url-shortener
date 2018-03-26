package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/friends-of-scalability/url-shortener/cmd/config"
	"github.com/friends-of-scalability/url-shortener/internal/urlshortener"
	"github.com/go-kit/kit/log"
)

func TestHTTP(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}
	var cfg config.Config
	{
		cfg.StorageType = "inmemory"
		cfg.Role = "full"
	}
	var ctx context.Context
	{
		ctx = context.Background()
	}

	var s urlshortener.Service
	{
		var err error
		s, err = urlshortener.NewService(&cfg)
		if err != nil {
			logger.Log("fatal", err)
			os.Exit(1)
		}
		s = urlshortener.NewLoggingService(logger, s)
	}

	h := urlshortener.MakeHandler(ctx, s, log.With(logger, "component", "HTTP"))
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, testcase := range []struct {
		method, url, body, want string
	}{
		{"POST", srv.URL + "/", `{"url":"https://crazy.url"}`, fmt.Sprintf(`{"shortURL":"%s/1","URL":"https://crazy.url"}`, srv.URL)},
		{"GET", srv.URL + "/1", "", `{"URL":"https://crazy.url"}`},
		{"GET", srv.URL + "/info/1", "", fmt.Sprintf(`{"URL":"https://crazy.url","shortURL":"%s/1","visitsCount":1}`, srv.URL)},
		{"GET", srv.URL + "/1", "", `{"URL":"https://crazy.url"}`},
		{"GET", srv.URL + "/info/1", "", fmt.Sprintf(`{"URL":"https://crazy.url","shortURL":"%s/1","visitsCount":2}`, srv.URL)},
	} {
		req, _ := http.NewRequest(testcase.method, testcase.url, strings.NewReader(testcase.body))
		resp, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		want, have := testcase.want, strings.TrimSpace(string(body))
		if resp.StatusCode != 200 {
			t.Errorf("%s %s %s: bad status code got %d", testcase.method, testcase.url, testcase.body, resp.StatusCode)
		} else if want != have {
			t.Errorf("%s %s %s: want %q, have %q", testcase.method, testcase.url, testcase.body, want, have)

		}
	}
}
