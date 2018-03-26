package urlshortener

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func decodeURLInfoResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response infoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func decodeURLInfoRequest(c context.Context, r *http.Request) (interface{}, error) {
	shURL := mux.Vars(r)
	if val, ok := shURL["shortURL"]; ok {
		//do something here
		return infoRequest{id: val}, nil
	}
	return nil, errMalformedURL
}
