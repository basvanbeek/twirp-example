// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/exporter/prometheus"
	oczipkin "go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"github.com/basvanbeek/twirp-example/internal/haberdasherserver"
	"github.com/basvanbeek/twirp-example/internal/hooks"
	"github.com/basvanbeek/twirp-example/rpc/haberdasher"
)

func main() {
	// Initialize OpenCensus with Zipkin Tracing and Prometheus metrics.
	var (
		reporter         = zipkinhttp.NewReporter("http://localhost:9411/api/v2/spans")
		localEndpoint, _ = zipkin.NewEndpoint("haberdasher", "localhost:8080")
		zipkinExporter   = oczipkin.NewExporter(reporter, localEndpoint)
	)
	defer reporter.Close()

	// Always trace for this demo.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(zipkinExporter)

	// using Prometheus for metrics
	prometheusExporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "haberdasher",
	})
	if err != nil {
		log.Fatalf("error configuring prometheus: %s\n", err.Error())
	}

	// add default ochttp server views
	view.Register(ochttp.DefaultServerViews...)

	// Report stats at every second.
	view.SetReportingPeriod(1 * time.Second)
	view.RegisterExporter(prometheusExporter)

	hook := hooks.LoggingHooks(os.Stderr)
	service := haberdasherserver.New()
	server := haberdasher.NewHaberdasherServer(service, hook)

	// allow us to both handle prometheus scaping as well as twirp service on
	// the same HTTP server.
	router := mux.NewRouter()
	// attach Haberdasher Service
	router.PathPrefix(haberdasher.HaberdasherPathPrefix).Handler(server)
	// attach prometheus
	router.Handle("/metrics", prometheusExporter)

	// wrap our router with ochttp tracing and metrics
	handler := &ochttp.Handler{Handler: router}

	log.Fatal(http.ListenAndServe(":8080", handler))
}
