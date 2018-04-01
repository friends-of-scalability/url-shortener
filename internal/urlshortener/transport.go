package urlshortener

import (
	"context"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
	zipkin "github.com/openzipkin/zipkin-go"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kittracing "github.com/go-kit/kit/tracing/zipkin"
	kithttp "github.com/go-kit/kit/transport/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// MakeHandler returns a handler for the urlshortener service.
func MakeHandler(ctx context.Context, us Service, logger kitlog.Logger, tracer *zipkin.Tracer) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{

		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kithttp.PopulateRequestContext, func(c context.Context, r *http.Request) context.Context {
			var scheme = "http"
			if r.TLS != nil {
				scheme = "https"
			}
			c = context.WithValue(c, contextKeyHTTPAddress, scheme+"://"+r.Host+"/")
			return c
		}),
		kittracing.HTTPServerTrace(tracer),
	}

	URLHealthzHandler := kithttp.NewServer(
		makeURLHealthzEndpoint(us),
		func(c context.Context, r *http.Request) (interface{}, error) {
			return nil, nil
		},
		encodeResponse,
		opts...,
	)

	hystrix.ConfigureCommand("shortener Request", hystrix.CommandConfig{Timeout: 1000})
	hystrix.ConfigureCommand("resolver Request", hystrix.CommandConfig{Timeout: 1000})
	hystrix.ConfigureCommand("info Request", hystrix.CommandConfig{Timeout: 1000})

	shortenerEndpoint := Hystrix("shortener Request",
		"Service currently unavailable", logger)(makeURLShortifyEndpoint(us))
	resolverEndpoint := Hystrix("resolver Request",
		"Service currently unavailable", logger)(makeURLRedirectEndpoint(us))
	infoEndpoint := Hystrix("info Request",
		"Service currently unavailable", logger)(makeURLInfoEndpoint(us))

	URLShortifyHandler := kithttp.NewServer(
		shortenerEndpoint,
		decodeURLShortenerRequest,
		encodeResponse,
		opts...,
	)
	URLRedirectHandler := kithttp.NewServer(
		resolverEndpoint,
		decodeURLRedirectRequest,
		encodeRedirectResponse,
		opts...,
	)
	URLInfoHandler := kithttp.NewServer(
		infoEndpoint,
		decodeURLInfoRequest,
		encodeResponse,
		opts...,
	)
	r.Path("/metrics").Handler(stdprometheus.Handler())
	r.Handle("/", URLShortifyHandler).Methods("POST")
	r.Handle("/healthz", URLHealthzHandler).Methods("GET")
	r.Handle("/{shortURL}", URLRedirectHandler).Methods("GET")
	r.Handle("/info/{shortURL}", URLInfoHandler).Methods("GET")

	return r
}
