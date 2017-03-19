package dronutz

import (
	"fmt"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

// NewTracer currently supports zipkin and lightstep
func ConfigureGlobalTracer(cfg Config, component string) error {
	var tracer opentracing.Tracer

	switch cfg.Tracer {

	case "zipkin":
		collector, err := zipkin.NewHTTPCollector(
			fmt.Sprintf("http://%s:%d/api/v1/spans", cfg.TracerHost, cfg.TracerPort))
		if err != nil {
			return err
		}
		tracer, err = zipkin.NewTracer(
			zipkin.NewRecorder(collector, false, "127.0.0.1:0", component))
		if err != nil {
			return err
		}

	case "lightstep":
		tracer = lightstep.NewTracer(lightstep.Options{
			AccessToken: cfg.TracerAccessToken,
			Collector: lightstep.Endpoint{
				Host:      cfg.TracerHost,
				Port:      cfg.TracerPort,
				Plaintext: true,
			},
			Tags: opentracing.Tags{
				lightstep.ComponentNameKey: component,
			},
		})

	case "log":
		tracer = basictracer.NewWithOptions(basictracer.Options{
			NewSpanEventListener: logSpanEvents,
			ShouldSample:         alwaysSample,
			Recorder:             noopRecorder{},
		})

	default:
		cfg.Tracer = "false"
		tracer = opentracing.NoopTracer{}
	}

	fmt.Println("Tracing enabled:", cfg.Tracer)
	opentracing.SetGlobalTracer(tracer)
	return nil
}

func logSpanEvents() func(basictracer.SpanEvent) {
	return func(event basictracer.SpanEvent) {
		switch event := event.(type) {
		case basictracer.EventCreate:
			fmt.Printf("\nSPAN START  \"%s\" \n", event.OperationName)
		case basictracer.EventFinish:
			fmt.Printf("\nSPAN FINISH \"%s\" %#v\n", event.Operation, event)
		}
	}
}

func alwaysSample(traceID uint64) bool {
	return true
}

type noopRecorder struct{}

func (noopRecorder) RecordSpan(span basictracer.RawSpan) {}
