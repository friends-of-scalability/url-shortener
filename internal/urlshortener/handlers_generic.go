package urlshortener

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/lb"
)

func debugRequest(req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%q\n", dump)
}
func debugResponse(resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%q", dump)
}

func encodeHTTPGenericRequest(ctx context.Context, r *http.Request, request interface{}) error {

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)

	return nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	if e, ok := response.(fallbacker); ok {
		if e.fallback() == errFallbackGracefully {
			fb := e.(fallbackResponse)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			return json.NewEncoder(w).Encode(fb.FallbackMessage)
		}
		encodeError(ctx, e.fallback(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

type fallbacker interface {
	fallback() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	logger := kitlog.NewLogfmtLogger(os.Stderr)
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	logger.Log("ERROR", err.Error())
	switch err {
	case errURLNotFound:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
	case errMalformedURL:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
	case lb.ErrNoEndpoints:
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Service unavailable, try again later",
		})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "something went wrong",
		})
	}

}
