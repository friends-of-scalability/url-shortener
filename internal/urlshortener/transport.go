package urlshortener

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

// MakeHandler returns a handler for the urlshortener service.
func MakeHandler(ctx context.Context, us Service, logger kitlog.Logger) http.Handler {
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
	}

	URLHealthzHandler := kithttp.NewServer(
		makeURLHealthzEndpoint(us),
		func(c context.Context, r *http.Request) (interface{}, error) {
			return nil, nil
		},
		encodeResponse,
		opts...,
	)
	URLShortifyHandler := kithttp.NewServer(
		makeURLShortifyEndpoint(us),
		decodeURLShortenerRequest,
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

	r.Handle("/", URLShortifyHandler).Methods("POST")
	r.Handle("/healthz", URLHealthzHandler).Methods("GET")
	r.Handle("/{shortURL}", URLRedirectHandler).Methods("GET")
	r.Handle("/info/{shortURL}", URLInfoHandler).Methods("GET")

	return r
}

func decodeURLShortenerRequest(c context.Context, r *http.Request) (interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	var t shortURL
	if !decoder.More() {
		return nil, errors.New("Empty request, cannot shortify the emptiness")

	}
	err := decoder.Decode(&t)
	if err != nil {
		return nil, err
	}
	if t.URL == "" {
		return nil, errors.New("Empty request, cannot shortify the emptiness")
	}
	return shortenerRequest{URL: t.URL}, nil
}

func decodeURLRedirectRequest(c context.Context, r *http.Request) (interface{}, error) {
	shURL := mux.Vars(r)
	return redirectRequest{id: shURL["shortURL"]}, nil
}

func decodeURLInfoRequest(c context.Context, r *http.Request) (interface{}, error) {
	shURL := mux.Vars(r)
	return infoRequest{id: shURL["shortURL"]}, nil

}

func encodeRedirectResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	if e, ok := response.(redirectResponse); ok && e.error() == nil {
		w.Header().Set("Location", e.URL)
		w.Header().Set("Referer", e.id)
		w.WriteHeader(http.StatusPermanentRedirect)
		return nil
	}
	encodeError(ctx, errMalformedURL, w)
	return nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case errURLNotFound:
		w.WriteHeader(http.StatusNotFound)
	case errMalformedURL:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
