package urlshortener

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func decodeURLShortenerResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response shortenerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
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
