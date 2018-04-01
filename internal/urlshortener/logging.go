package urlshortener

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

type loggingService struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

// Login to the system.
func (s *loggingService) Shortify(ctx context.Context, longURL string) (mapping *shortURL, err error) {
	defer func(begin time.Time) {
		s.logger.Log("method", "shortify", "url", longURL, "took", time.Since(begin), "err", err)
	}(time.Now())
	return s.Service.Shortify(ctx, longURL)
}

func (s *loggingService) Resolve(ctx context.Context, shortURL string) (mapping *shortURL, err error) {
	defer func(begin time.Time) {
		s.logger.Log("method", "Resolve", "shortURLId", shortURL, "took", time.Since(begin), "err", err)
	}(time.Now())
	return s.Service.Resolve(ctx, shortURL)
}

func (s *loggingService) GetInfo(ctx context.Context, shortURL string) (mapping *shortURL, err error) {
	defer func(begin time.Time) {
		s.logger.Log("method", "GetInfo", "shortURLId", shortURL, "took", time.Since(begin), "err", err)
	}(time.Now())
	return s.Service.GetInfo(ctx, shortURL)
}
