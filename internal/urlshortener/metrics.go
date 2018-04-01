package urlshortener

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

func NewMetricsService(requestCount metrics.Counter,
	requestLatency metrics.Histogram, s Service) Service {
	return &metricsMiddleware{
		s,
		requestCount,
		requestLatency,
	}

}

type metricsMiddleware struct {
	Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func (mw *metricsMiddleware) Shortify(ctx context.Context, longURL string) (mapping *shortURL, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Shortify"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.Shortify(ctx, longURL)
}

func (mw *metricsMiddleware) Resolve(ctx context.Context, shortURL string) (mapping *shortURL, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Resolve"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.Resolve(ctx, shortURL)
}

func (mw *metricsMiddleware) GetInfo(ctx context.Context, shortURL string) (mapping *shortURL, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Info"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.GetInfo(ctx, shortURL)
}
