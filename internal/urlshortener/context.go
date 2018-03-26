package urlshortener

type contextKey string

func (c contextKey) String() string {
	return "urlshortener context key " + string(c)
}

var (
	contextKeyHTTPAddress      = contextKey("URLShortenerServiceHTTPAddr")
	contextKeyAPIGWHTTPAddress = contextKey("URLShortenerServiceAPIGWAddr")
)
