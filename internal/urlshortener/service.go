package urlshortener

// Service provides operations on Users.
type Service interface {
	//Creates a new shortURL from a longURL
	Shortify(longURL string) (*shortURL, error)
	//Retrieves a long URL from a short one
	Resolve(shortURL string) (*shortURL, error)
	GetInfo(shortURL string) (*shortURL, error)
	IsHealthy() (bool, error)
}
