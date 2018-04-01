package urlshortener

import "context"

// Service provides operations on Users.
type Service interface {
	//Creates a new shortURL from a longURL
	Shortify(ctx context.Context, longURL string) (*shortURL, error)
	//Retrieves a long URL from a short one
	Resolve(ctx context.Context, shortURL string) (*shortURL, error)
	GetInfo(ctx context.Context, shortURL string) (*shortURL, error)
	IsHealthy() (bool, error)
}
