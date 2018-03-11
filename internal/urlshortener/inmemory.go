package urlshortener

import (
	"github.com/friends-of-scalability/url-shortener/pkg"
)

// shortURLRepository is an in-memory user database.
type shortURLInMemoryRepository struct {
	shortURLRepository map[string]*shortURL
}

func newInMemory() (shortURLStorage, error) {
	return &shortURLInMemoryRepository{
		shortURLRepository: map[string]*shortURL{},
	}, nil
}

// ByShortURL finds and URL in our databse.
func (u *shortURLInMemoryRepository) ByURL(URL string) (*shortURL, error) {

	for _, mapping := range u.shortURLRepository {
		if mapping.URL == URL {
			return mapping, nil
		}
	}
	return nil, errURLNotFound
}

func (u *shortURLInMemoryRepository) IsHealthy() (bool, error) { return true, nil }

func (u *shortURLInMemoryRepository) ByID(id string) (*shortURL, error) {

	key, err := base62.Decode(id)
	if err != nil {
		return nil, errMalformedURL
	}
	for _, mapping := range u.shortURLRepository {
		if mapping.ID == key {
			return mapping, nil
		}
	}
	return nil, errURLNotFound
}

// ByShortURL finds and URL in our databse.
func (u *shortURLInMemoryRepository) Save(item *shortURL) (*shortURL, error) {

	m, err := u.ByURL(item.URL)
	if err != errURLNotFound {
		return m, err
	}
	var mapping shortURL
	var autoInc uint64
	for _, mapping := range u.shortURLRepository {
		if mapping.ID > autoInc {
			autoInc = mapping.ID
		}
	}
	autoInc++
	mapping.URL = item.URL
	mapping.VisitsCounter = 0
	mapping.ID = autoInc
	encodedkey := base62.Encode(mapping.ID)
	u.shortURLRepository[encodedkey] = &mapping
	return &mapping, nil
}
