package urlshortener

import (
	"context"
	"fmt"
	"strings"

	"github.com/friends-of-scalability/url-shortener/pkg"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd/lb"
)

type shortenerRequest struct {
	URL string
}

type shortenerResponse struct {
	ShortURL string `json:"shortURL,omitempty"`
	URL      string `json:"URL,omitempty"`
	Err      string `json:"error,omitempty"`
}

type redirectRequest struct {
	id string
}

type redirectResponse struct {
	URL string `json:"URL,omitempty"`
	id  string
	Err string `json:"error,omitempty"`
}

type infoRequest struct {
	id string
}

type infoResponse struct {
	URL      string `json:"URL,omitempty"`
	ShortURL string `json:"shortURL,omitempty"`
	Visits   uint64 `json:"visitsCount,omitempty"`
	Err      string `json:"error,omitempty"`
}
type fallbackResponse struct {
	FallbackMessage string `json:"fallback,omitempty"`
	UpstreamError   string `json:"upstreamError"`
}

type healthzResponse struct {
	Msg string `json:"msg,omitempty"`
	Err string `json:"error,omitempty"`
}

func makeURLShortifyEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(shortenerRequest)
		m, err := s.Shortify(req.URL)
		if err != nil {
			return shortenerResponse{Err: err.Error()}, nil
		}
		var host string
		if ctx.Value(contextKeyAPIGWHTTPAddress) != nil {
			host = ctx.Value(contextKeyAPIGWHTTPAddress).(string)
		} else {
			host = ctx.Value(contextKeyHTTPAddress).(string)
		}
		return shortenerResponse{ShortURL: host + base62.Encode(m.ID), URL: m.URL}, nil
	}
}

func makeURLHealthzEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// if db ok and service ok
		status, err := s.IsHealthy()
		if !status {
			return healthzResponse{Msg: "Nope! Something went wrong :(", Err: err.Error()}, nil
		}
		return healthzResponse{Msg: "Always look at the bright side of life :)"}, nil
	}
}

func makeURLRedirectEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(redirectRequest)
		m, err := s.Resolve(req.id)
		if err != nil {
			return redirectResponse{Err: err.Error()}, nil
		}
		var host string
		if ctx.Value(contextKeyAPIGWHTTPAddress) != nil {
			host = ctx.Value(contextKeyAPIGWHTTPAddress).(string)
		} else {
			host = ctx.Value(contextKeyHTTPAddress).(string)
		}
		return redirectResponse{URL: m.URL, id: host + req.id}, nil
	}
}

func makeURLInfoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(infoRequest)
		m, err := s.GetInfo(req.id)
		if err != nil {
			return infoResponse{Err: err.Error()}, nil
		}
		var host string
		if ctx.Value(contextKeyAPIGWHTTPAddress) != nil {
			host = ctx.Value(contextKeyAPIGWHTTPAddress).(string)
		} else {
			host = ctx.Value(contextKeyHTTPAddress).(string)
		}
		return infoResponse{URL: m.URL, ShortURL: host + req.id, Visits: m.VisitsCounter}, nil
	}
}

func (r redirectResponse) error() error { return dealWithErrors(r.Err) }

func (r shortenerResponse) error() error { return dealWithErrors(r.Err) }

func (r healthzResponse) error() error { return dealWithErrors(r.Err) }

func (r infoResponse) error() error { return dealWithErrors(r.Err) }

func (r fallbackResponse) fallback() error { return doFallback(r) }

func dealWithErrors(errorReason string) error {
	if errorReason != "" {
		switch errorReason {
		case errMalformedURL.Error():
			return errMalformedURL
		case errURLNotFound.Error():
			return errURLNotFound
		case lb.ErrNoEndpoints.Error():
			return lb.ErrNoEndpoints
		default:
			return fmt.Errorf(errorReason)
		}
	}
	return nil
}

func doFallback(fallbackMessage fallbackResponse) error {
	if strings.Contains(fallbackMessage.UpstreamError, "no endpoints available") {
		return errFallbackFail
	}
	return errFallbackGracefully
}
