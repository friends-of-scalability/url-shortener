package urlshortener

type shortURLStorage interface {
	//Creates a new shortURL from a longURL
	Save(Item *shortURL) (*shortURL, error)
	ByID(id string) (*shortURL, error)
	ByURL(URL string) (*shortURL, error)
	IsHealthy() (bool, error)
}
