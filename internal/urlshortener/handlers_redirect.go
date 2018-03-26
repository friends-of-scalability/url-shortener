package urlshortener

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func encodeRedirectResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}

	if e, ok := response.(redirectResponse); ok && e.error() == nil {
		if ctx.Value(contextKeyAPIGWHTTPAddress) != nil {
			return json.NewEncoder(w).Encode(e)
		}
		w.Header().Set("Location", e.URL)
		w.WriteHeader(http.StatusPermanentRedirect)
		return nil
	}
	return encodeResponse(ctx, w, response)
}

func decodeURLRedirectRequest(c context.Context, r *http.Request) (interface{}, error) {

	shURL := mux.Vars(r)
	if val, ok := shURL["shortURL"]; ok {
		return redirectRequest{id: val}, nil
	}
	return nil, errMalformedURL

}

func decodeAPIGWRedirectResponse(ctx context.Context, resp *http.Response) (interface{}, error) {

	var response redirectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}
