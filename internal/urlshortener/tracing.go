package urlshortener

import (
	"context"
	"time"

	zipkin "github.com/openzipkin/zipkin-go"
	model "github.com/openzipkin/zipkin-go/model"
)

func NewTracingService(tracer *zipkin.Tracer, s Service) Service {
	return &tracingMiddleware{
		s,
		tracer,
	}

}

type tracingMiddleware struct {
	Service
	tracer *zipkin.Tracer
}

func (tw *tracingMiddleware) Shortify(ctx context.Context, longURL string) (mapping *shortURL, err error) {
	var (
		spanContext model.SpanContext
		serviceName = "Shortener"
		serviceHost string
		queryLabel  = "Shortener"
		query       string
	)
	value := ctx.Value(contextKeyHTTPAddress)
	host, ok := value.(string)
	if ok {
		serviceHost = host
		query = longURL
	} else {
		serviceHost = ""
	}

	// add interesting timed event to our span

	// do the actual query...

	// let's annotate the end...

	// we're done with this span.
	// retrieve the parent span from context to use as parent if available.
	if parentSpan := zipkin.SpanFromContext(ctx); parentSpan != nil {
		spanContext = parentSpan.Context()
	}

	// create the remote Zipkin endpoint
	ep, _ := zipkin.NewEndpoint(serviceName, serviceHost)

	// create a new span to record the resource interaction
	zipkin.Parent(spanContext)
	span := tw.tracer.StartSpan(
		queryLabel,
	)
	span.Annotate(time.Now(), "Server Receive")
	defer func() {

		span.Tag("query", query)
		span.SetRemoteEndpoint(ep)
		span.Annotate(time.Now(), "Server Send")
		span.Finish()
	}()
	return tw.Service.Shortify(ctx, longURL)
}

func (tw *tracingMiddleware) Resolve(ctx context.Context, shortURL string) (mapping *shortURL, err error) {
	var (
		spanContext model.SpanContext
		serviceName = "Resolve"
		serviceHost string
		queryLabel  = "Resolve"
		query       string
	)
	value := ctx.Value(contextKeyHTTPAddress)
	host, ok := value.(string)
	if ok {
		serviceHost = host
		query = shortURL
	} else {
		serviceHost = ""
	}
	if parentSpan := zipkin.SpanFromContext(ctx); parentSpan != nil {
		spanContext = parentSpan.Context()
	}

	// create the remote Zipkin endpoint
	ep, _ := zipkin.NewEndpoint(serviceName, serviceHost)

	// create a new span to record the resource interaction
	zipkin.Parent(spanContext)
	span := tw.tracer.StartSpan(
		queryLabel,
	)
	span.Annotate(time.Now(), "Server Receive")
	defer func() {

		span.Tag("query", query)
		span.SetRemoteEndpoint(ep)
		span.Annotate(time.Now(), "Server Send")
		span.Finish()
	}()

	return tw.Service.Resolve(ctx, shortURL)
}

func (tw *tracingMiddleware) GetInfo(ctx context.Context, shortURL string) (mapping *shortURL, err error) {
	var (
		spanContext model.SpanContext
		serviceName = "Info"
		serviceHost string
		queryLabel  = "Info"
		query       string
	)
	value := ctx.Value(contextKeyHTTPAddress)
	host, ok := value.(string)
	if ok {
		serviceHost = host
		query = shortURL
	} else {
		serviceHost = ""
	}
	if parentSpan := zipkin.SpanFromContext(ctx); parentSpan != nil {
		spanContext = parentSpan.Context()
	}

	// create the remote Zipkin endpoint
	ep, _ := zipkin.NewEndpoint(serviceName, serviceHost)

	// create a new span to record the resource interaction
	zipkin.Parent(spanContext)
	span := tw.tracer.StartSpan(
		queryLabel,
	)
	span.Annotate(time.Now(), "Server Receive")
	defer func() {

		span.Tag("query", query)
		span.SetRemoteEndpoint(ep)
		span.Annotate(time.Now(), "Server Send")
		span.Finish()
	}()
	return tw.Service.GetInfo(ctx, shortURL)
}
