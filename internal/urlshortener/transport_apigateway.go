package urlshortener

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/friends-of-scalability/url-shortener/cmd/config"
	endpoint "github.com/go-kit/kit/endpoint"
	sd "github.com/go-kit/kit/sd"
	zipkin "github.com/openzipkin/zipkin-go"

	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	dnssrv "github.com/go-kit/kit/sd/dnssrv"
	"github.com/go-kit/kit/sd/lb"
	kittracing "github.com/go-kit/kit/tracing/zipkin"
	kithttp "github.com/go-kit/kit/transport/http"
)

func getCurrentAddr(r *http.Request) string {
	var scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/"
}

func MakeAPIGWHandler(ctx context.Context, us Service, logger kitlog.Logger, cfg *config.Config, tracer *zipkin.Tracer) http.Handler {
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
	r.Path("/metrics").Handler(stdprometheus.Handler())
	r.Handle("/healthz", URLHealthzHandler).Methods("GET")
	hystrix.ConfigureCommand("shortener Request", hystrix.CommandConfig{Timeout: 100000})
	hystrix.ConfigureCommand("resolver Request", hystrix.CommandConfig{Timeout: 1000})
	hystrix.ConfigureCommand("info Request", hystrix.CommandConfig{Timeout: 1000})

	resolverEndpoint := Hystrix("resolver Request",
		"Resolver service currently unavailable", logger)(newEndpoint(ctx, "resolver", "GET", logger, cfg, tracer))
	r.Handle("/{shortURL}", kithttp.NewServer(resolverEndpoint, decodeURLRedirectRequest, encodeRedirectResponse, opts...)).Methods("GET")

	infoEndpoint := Hystrix("info Request",
		"Info service currently unavailable", logger)(newEndpoint(ctx, "info", "GET", logger, cfg, tracer))
	r.Handle("/info/{shortURL}", kithttp.NewServer(infoEndpoint, decodeURLInfoRequest, encodeResponse, opts...)).Methods("GET")

	shortenerEndpoint := Hystrix("shortener Request",
		"Shortener service currently unavailable", logger)(newEndpoint(ctx, "shortener", "POST", logger, cfg, tracer))
	r.Handle("/", kithttp.NewServer(shortenerEndpoint, decodeURLShortenerRequest, encodeResponse, opts...)).Methods("POST")

	return r
}

func newEndpoint(ctx context.Context, action, method string, logger kitlog.Logger, cfg *config.Config, tracer *zipkin.Tracer) endpoint.Endpoint {
	var instancer *dnssrv.Instancer
	switch action {
	case "resolver":
		instancer = dnssrv.NewInstancer(cfg.ServiceDiscovery.Resolver, 200*time.Millisecond, kitlog.NewNopLogger())
	case "info":
		instancer = dnssrv.NewInstancer(cfg.ServiceDiscovery.Resolver, 200*time.Millisecond, kitlog.NewNopLogger())
	case "shortener":
		instancer = dnssrv.NewInstancer(cfg.ServiceDiscovery.Shortener, 200*time.Millisecond, kitlog.NewNopLogger())

	}
	factory := endpointFactory(ctx, action, method, logger, cfg, tracer)
	endpointer := sd.NewEndpointer(instancer, factory, logger)
	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(3, 500*time.Millisecond, balancer)
	return retry
}

func endpointFactory(ctx context.Context, action, method string, logger kitlog.Logger, cfg *config.Config, tracer *zipkin.Tracer) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		s := strings.Split(instance, ":")
		// removing port since service discovery is getting a wrong one
		if len(s) == 0 || len(s) > 2 {
			return nil, nil, fmt.Errorf("Got wrong address from service discovery, something went wrong")
		}
		instance = s[0]
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance + ":" + cfg.ExposedPort
		}
		tgt, err := url.Parse(instance)

		if err != nil {
			return nil, nil, err
		}
		var (
			enc kithttp.EncodeRequestFunc
			dec kithttp.DecodeResponseFunc
		)
		switch action {
		case "resolver":
			enc, dec = encodeAPIGWRedirectRequest, decodeAPIGWRedirectResponse
		case "info":
			tgt.Path = "/info"
			enc, dec = encodeAPIGWInfoRequest, decodeURLInfoResponse
		case "shortener":
			enc, dec = encodeHTTPGenericRequest, decodeURLShortenerResponse
		default:
			return nil, nil, fmt.Errorf("unknown resolver action %q", action)
		}
		before := kithttp.ClientBefore(func(ctx context.Context, req *http.Request) context.Context {
			logger.Log("TYPE", "HTTP CLIENT", "METHOD", method, "ACTION", action, "HOST", tgt.String(), "URI", req.RequestURI)
			req.Header.Set("X-Forwarded-Host", ctx.Value(contextKeyHTTPAddress).(string))
			return ctx
		})

		// global client middlewares
		options := []kithttp.ClientOption{
			kittracing.HTTPClientTrace(tracer),
			before,
		}
		return kithttp.NewClient(method, tgt, enc, dec, options...).Endpoint(), nil, nil
	}
}

func encodeAPIGWRedirectRequest(ctx context.Context, req *http.Request, request interface{}) error {

	originalRequest, ok := request.(redirectRequest)
	if !ok {
		return fmt.Errorf("Cannot cast request to an redirectRequest")
	}
	url := url.URL{
		Scheme:  req.URL.Scheme,
		Host:    req.URL.Host,
		Path:    "/" + originalRequest.id,
		RawPath: "/" + url.QueryEscape(originalRequest.id),
	}
	req.URL = &url
	return nil

}
func encodeAPIGWInfoRequest(ctx context.Context, req *http.Request, request interface{}) error {
	originalRequest, ok := request.(infoRequest)
	if !ok {
		return fmt.Errorf("Cannot cast request to an inforequest")
	}
	url := url.URL{
		Scheme:  req.URL.Scheme,
		Host:    req.URL.Host,
		Path:    "/info/" + originalRequest.id,
		RawPath: "/info/" + url.QueryEscape(originalRequest.id),
	}
	req.URL = &url

	return nil

}
