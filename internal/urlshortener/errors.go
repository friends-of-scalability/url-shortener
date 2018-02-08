package urlshortener

import "errors"

var (
	errURLNotFound  = errors.New("This URL has not been found in our database")
	errMalformedURL = errors.New("This URL is not valid")
)
