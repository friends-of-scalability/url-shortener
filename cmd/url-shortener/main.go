package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/friends-of-scalability/url-shortener/internal/urlshortener"
	"github.com/go-kit/kit/log"

	"context"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8080", "HTTP listen address")
		makeFakeLoad = flag.Bool("fakeLoad", false, "enable to generate fake load using stress")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}

	var ctx context.Context
	{
		ctx = context.Background()
	}

	var s urlshortener.Service
	{
		s = urlshortener.NewService(*makeFakeLoad)
		s = urlshortener.NewLoggingService(logger, s)
	}

	var h http.Handler
	{
		h = urlshortener.MakeHandler(ctx, s, log.With(logger, "component", "HTTP"))
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)

	}()

	logger.Log("exit", <-errs)
}
