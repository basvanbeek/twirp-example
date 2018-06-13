# twirp-example with opencensus instrumentation

This is an example Twirp service with [OpenCensus] instrumentation for
educational purposes. Learn more about Twirp at its
[website](https://twitchtv.github.io/twirp/docs/intro.html) or
[repo](https://github.com/twitchtv/twirp).

## Example rationale

This is a fork from the standard twirp-example project as provided by twitchtv.
The demo highlights how easy it is to instrument a Twirp service with
[OpenCensus]. Since Twirp uses the standard HTTP server and client libraries
from Go and also has a different path for each service method, it is possible to
use the standard ochttp middleware for instrumenting Twirp services. This demo
highlights that feature.

### OpenCensus
[OpenCensus] is a single distribution of libraries for metrics and distributed
tracing. It supports multiple tracing and metrics backends. For more information
see the [website](https://opencensus.io).

### Zipkin
For tracing we have chosen the [Zipkin] backend in this example, as it is very
easy to get it going. You can run the [Zipkin] jar if having Java installed or
use the official Docker container. For more information on getting started with
[Zipkin] see the [Zipkin Quickstart](https://zipkin.io/pages/quickstart).

### Prometheus
For metrics we have chosen the [Prometheus] backend in this example, again
because it is easy to get started. You will need to configure [Prometheus] to
scrape the service. A typical simple config scraping [Prometheus] itself and our
haberdasher service can look like this if running [Prometheus] locally:

```yaml
global:
  scrape_interval:     10s
  evaluation_interval: 10s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
    - targets: ['localhost:9090']
  - job_name: 'haberdasher'
    static_configs:
    - targets: ['localhost:8080']
```

[OpenCensus]: (https://opencensus.io)
[Zipkin]: (https://zipkin.io)
[Prometheus]: (https://prometheus.io/)

## Try it out

First, download this repo with the Go tool:
```
go get github.com/basvanbeek/twirp-example/...
cd $GOPATH/src/github.com/basvanbeek/twirp-example
```

Next, try building the client and server binaries:
```
go build ./cmd/client
go build ./cmd/server
```

And run them. In one terminal session:
```
./server
```

And in another:
```
./client
```

In the client, you should see something like this:
```
-> % ./client
size:12 color:"red" name:"baseball cap"
```

In the server, something like this:
```% ./server
received req svc="Haberdasher" method="MakeHat"
response sent svc="Haberdasher" method="MakeHat" time="109.01µs"
```

## Code structure

The protobuf definition for the service lives in
`rpc/haberdasher/haberdasher.proto`. The `rpc` directory name is a good way to
signal where your service definitions reside.

The generated Twirp and Go protobuf code is in the same directory. This makes it
easy to import for both internal and external users - internally, we need to
import it to have the right types for our implmentation of the service
interface, and externally it needs to be available so clients can import it.

The implementation of the server is in `internal/haberdasherserver`. Putting it
in `internal` means that it can't be imported from outside this repository,
which is nice because we don't have to think about API stability nearly as much.

In addition, `internal/hooks/logging.go` has a file which provides
[`ServerHooks`](https://twitchtv.github.io/twirp/docs/hooks.html) which can log
requests. This is a good demo of how you can use hooks to extend Twirp's basic
functionality - you can use hooks to add instrumentation or even for
authentication.

Finally, `cmd/server` and `cmd/client` wrap things together into executable main
packages.

## License
This library is licensed under the Apache 2.0 License.
