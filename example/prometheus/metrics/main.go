package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	ctx := context.Background()
	configureOpentelemetry()

	meter := global.MeterProvider().Meter("example")
	counter, err := meter.SyncInt64().Counter(
		"test.my_counter",
		instrument.WithDescription("Just a test counter"),
	)
	if err != nil {
		panic(err)
	}

	for {
		n := rand.Intn(1000)
		time.Sleep(time.Duration(n) * time.Millisecond)

		counter.Add(ctx, 1)
	}
}

func configureOpentelemetry() {
	if err := runtimemetrics.Start(); err != nil {
		panic(err)
	}

	exporter := configureMetrics()

	registry := prometheus.NewRegistry()

	if err := registry.Register(exporter.Collector); err != nil {
		log.Printf("error registering collector: %s", err)
		return
	}

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fmt.Println("listenening on http://localhost:8088/metrics")

	go func() {
		_ = http.ListenAndServe(":8088", nil)
	}()
}

func configureMetrics() otelprom.Exporter {
	exporter := otelprom.New()
	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	global.SetMeterProvider(provider)

	return exporter
}
