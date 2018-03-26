package urlshortener

import "errors"

var (
	errURLNotFound        = errors.New("This URL does not exist yet")
	errMalformedURL       = errors.New("This URL is not valid")
	errFallbackFail       = errors.New("uops! something is really wrong")
	errFallbackGracefully = errors.New("Something went wrong trying to recover")
)
