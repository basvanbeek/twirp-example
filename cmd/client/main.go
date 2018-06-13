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
	"context"
	"fmt"
	"log"
	"net/http"

	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/twitchtv/twirp"
	oczipkin "go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"

	"github.com/basvanbeek/twirp-example/rpc/haberdasher"
)

const (
	zipkinURL = "http://localhost:9411/api/v2/spans"
)

func main() {
	// Initialize OpenCensus with Zipkin Tracing Backend
	var (
		reporter         = zipkinhttp.NewReporter(zipkinURL)
		localEndpoint, _ = zipkin.NewEndpoint("haberdasher-client", "localhost:0")
		zipkinExporter   = oczipkin.NewExporter(reporter, localEndpoint)
	)
	defer reporter.Close()

	// Always trace for this demo.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(zipkinExporter)

	// customize a generic http client with the OpenCensus tracing Roundtripper.
	httpClient := &http.Client{Transport: &ochttp.Transport{}}

	// Create a client capable of talking to a Haberdasher server running on
	// localhost. This is a generated function call.
	client := haberdasher.NewHaberdasherJSONClient("http://localhost:8080", httpClient)

	var (
		hat *haberdasher.Hat
		err error
	)

	// add an app span for our Make Hat call
	ctx, span := trace.StartSpan(context.Background(), "do MakeHat")
	defer span.End()

	// Call the client's 'MakeHat' method, retrying up to five times.
	for i := 0; i < 5; i++ {
		hat, err = client.MakeHat(ctx, &haberdasher.Size{Inches: 12})
		if err != nil {
			// We got an error. Is it a twirp Error?
			if twerr, ok := err.(twirp.Error); ok {
				// Twirp errors support custom, arbitrary metadata. For example, a
				// server could inform a client that a particular error is retryable.
				if twerr.Meta("retryable") != "" {
					log.Printf("got error %q, retrying", twerr)
					continue
				}
			}
			log.Fatal(err)
		} else {
			break
		}
	}

	// Print out the response.
	fmt.Printf("%+v\n", hat)
}
