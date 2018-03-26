package urlshortener

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

func MakeResolverHandler(ctx context.Context, us Service, logger kitlog.Logger) http.Handler {
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

	r.Handle("/healthz", URLHealthzHandler).Methods("GET")
	r.Handle("/{shortURL}", URLRedirectHandler).Methods("GET")
	r.Handle("/info/{shortURL}", URLInfoHandler).Methods("GET")

	return r
}
