package urlshortener

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	zipkin "github.com/openzipkin/zipkin-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	kitlog "github.com/go-kit/kit/log"
	kittracing "github.com/go-kit/kit/tracing/zipkin"
	kithttp "github.com/go-kit/kit/transport/http"
)

func MakeResolverHandler(ctx context.Context, us Service, logger kitlog.Logger, tracer *zipkin.Tracer) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kithttp.PopulateRequestContext, func(c context.Context, r *http.Request) context.Context {
			var scheme = "http"
			if r.TLS != nil {
				scheme = "https"
			}
			c = context.WithValue(c, contextKeyAPIGWHTTPAddress, r.Header.Get("X-Forwarded-Host"))
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
	URLRedirectHandler := kithttp.NewServer(
		makeURLRedirectEndpoint(us),
		decodeURLRedirectRequest,
		encodeRedirectResponse,
		opts...,
	)
	URLInfoHandler := kithttp.NewServer(
		makeURLInfoEndpoint(us),
		decodeURLInfoRequest,
		encodeResponse,
		opts...,
	)
	r.Path("/metrics").Handler(stdprometheus.Handler())
	r.Handle("/healthz", URLHealthzHandler).Methods("GET")
	r.Handle("/{shortURL}", URLRedirectHandler).Methods("GET")
	r.Handle("/info/{shortURL}", URLInfoHandler).Methods("GET")

	return r
}
