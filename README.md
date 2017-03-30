# dr🍩nutz: An OpenTracing Walktrhough

Welcome to Dronutz! This is a sample application and OpenTracing walkthrough, 
written in Go.

OpenTracing is a vendor-neutral, open standard for distributed tracing. To 
learn more, check out [opentracing.io](http://opentracing.io), and try the 
walkthrough below!

In addition to doing the walkthrough yourself, the results of each step are 
saved in their own git branch.
- `git checkout step-1` to see dronutz with turnkey/network tracing.
- `git checkout step-2` to see an example of application-level client tracing.

## Step 0: Setup Dronutz
### Installation
Dronutz can be installed into your GOPATH via `go get`.

```
go get github.com/tedsuo/ot-walkthrough-go/dronutz/cmd/api
```

### Running
Dronutz has two server components, `API` and `Kitchen`, which can be run in two
separate terminal windows. Both are configured with the same config file.

```
./scripts/run_kitchen.sh -config config_example.yml
./scripts/run_api.sh -config config_example.yml
```

In your web broswer, navigate to http://127.0.0.1:10001 and get yourself 
some donuts.

### Pick a Tracer
The configuration file currently supports three tracers: `zipkin`, `lightstep`, 
and `log`.

#### zipkin
The easiest way to run zipkin locally is with 
[docker](https://github.com/openzipkin/docker-zipkin#running).

```
docker run -d -p 9411:9411 openzipkin/zipkin
```

```yaml
tracer: zipkin
tracer_host: localhost
tracer_port: 9411 
```

#### lightstep
If you're using lightstep, you will need your access token.

```yaml
tracer: lightstep
tracer_host: LS_HOST
tracer_port: LS_PORT
tracer_access_token: LS_ACCESS_TOKEN
```

#### log
Log simply logs the traces, showing off how you can create structured logging 
without being tied to a particular logging framework.

```yaml
tracer: log
```

## Step 1: TurnKey Tracing
When you go to add tracing to a system, the best place to start is installing 
OpenTracing plugins for the OSS components you are using. Instrumenting your 
networking libraries, web frameworks, and service clients quickly gives you a 
lot of information about your distributed system, without requiring you to 
change a lot of code.

To do this, let's change the startup of `dronutz/cmd/api` and 
`dronutz/cmd/kitchen` to include tracing.

### Start the GlobalTracer
In OpenTracing, there is a concept of a global tracer for everyone to access. 
As a convenience, we already have a function `dronutz.ConfigureGlobalTracer` 
that works with the dronutz configuration file. In the `main` of both API and 
Kitchen, add it before initializing any other components.

```go
err = dronutz.ConfigureGlobalTracer(cfg, "api") // <-- be sure to name api and kitchen differently
if err != nil {
	panic(err)
}
```

There's a trick in this first block of code. Make sure to change the second 
parameter to `kitchen` when you add it to the Kitchen component.

### Instrument the HTTP Server
The first thing to always intrument is the inbound RPC server. For API, 
this is an http server. Use the `nethttp` package from `opentracing-contrib` 
to wrap the top-level service `http.Handler` with opentracing middleware.

https://godoc.org/github.com/opentracing-contrib/go-stdlib/nethttp

```go
err = http.ListenAndServe(
	cfg.APIAddress(),
	nethttp.Middleware(
		opentracing.GlobalTracer(),
		service.ServeMux(),
		nethttp.OperationNameFunc(func(req *http.Request) string {
			return "/dronutz.API" + req.URL.Path // <-- ensures we use a standard naming convention
		}),
	),
)
```

### Intsrument the gRPC Client
Now that we have inbound RPC taken care of, let's look at our outbound service 
clients. In API, we talk to the Kitchen service over gRPC.

https://godoc.org/github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc

```go
conn, err := grpc.Dial(
	cfg.KitchenAddress(),
	grpc.WithInsecure(),
	grpc.WithUnaryInterceptor(
		otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()),
	),
)
```
At this point, the API service is successfully wrapped.

### Intsrument the gRPC Server
Though we already have the client side wrapped, let's now wrap the server side 
of Kitchen's gRPC service.

```go
server := grpc.NewServer(
	grpc.UnaryInterceptor(
		otgrpc.OpenTracingServerInterceptor(
			opentracing.GlobalTracer(),
			otgrpc.LogPayloads()),
	),
)
```

This will allow us to automatically have traces bound to rpc endpoints, so we 
can start tracing internally, and will let us know if something other than the 
API server were to start talking to this backend service.

Try including the `otgrpc.LogPayloads()` option to have more detailed 
information show up in your traces as logs.

### Check it out in yout Tracer
Now that we're all hooked up, try buying some donuts in the browser. You should 
see the traces appear in your tracer, or in the terminal if you are using the 
`log` tracer.

Search for traces starting with `/dronutz.API` or `/dronutz.Kitchen` to see the 
patterns of requests that occur when you click the order button.

## Step 2: Enhance
Now that the components in our system are linked up at the networking level, we 
can start adding application level tracing by tying multiple network calls 
together into a single trace.

In dronutz, we'd like to know the time and resources involved with buying a 
donut, from the moment it is ordered to when it is delivered. Let's add 
OpenTracing to the client so that we can bundle all of these network calls into 
the same trace.

__NOTE:__ Unfortunately, zipkin does not have opentracing support for javascript at 
this time, so you must use the LightStep tracer for this part of the 
walkthrough. Hopefully this will be changing soon.

### Add a GlobalTracer to the Client
Just like on the server, we start be configuring the global tracer. Do so at 
the begining of document ready in `client/client.js`.

```js
Tracer.initGlobalTracer(LightStep.tracer({
  component_name      : 'client',
  access_token        : Config.tracer_access_token,
  collector_host      : Config.tracer_host,
  collector_port      : Config.tracer_port,
  collector_encryption: "none",
  xhr_instrumentation : false,
}));
```

### Instrument the Order object
On the client, we don't want to trace an individual request, but record a 
logical collection of actions under a single span. The Order object in dronutz 
encapsulates the logic of buying donuts. To start tracing the order process, 
store a span on the order, starting and finishing it when the order 
activates and completes.

```js
Order.prototype.activate = function(){
  this._span = Tracer.startSpan('/dronutz.Client/BuyDonuts');
  this._state = "active";
}

Order.prototype.complete = function(){
  this._span.finish();
  return this.reset();
}
```

### Inject the Span Context
Because we are now doing manual tracing, we have to inject the span context into 
the http requests ourselves, so that the endpoints on the server can extract and 
bind their child spans to the same trace. To do this, add a function to `Order` 
that serializes the span's context into a set of http headers.

```js
Order.prototype.headers = function(){
  var otHeaders = {};
  Tracer.inject(this._span, Tracer.FORMAT_TEXT_MAP, otHeaders);
  return otHeaders;
}
```
Now that we can generate headers, use them as the headers for the `/order` and 
`/status` requests in `client.js`:
```js
$.ajax('/order', {
  headers: order.headers(),
  method:'POST',
  // yadda yadda
}
```

And we're done! Buy some donuts, check out `/dronutz.Client/BuyDonuts`, and 
notice how the order and polling requests are now grouped under a single span, 
with timing information for the entire operation.

### Step 3: Have fun
If you still have time, try to trace other things! For example, maybe we would 
like to extend the trace on the client to record when an order is opened, and 
log when each donut is added to the cart.

Also, that kitchen component is pretty strange. What's going on in there? Think of how you 
might trace it to find the answers. Alternatively, add a `DroneDispatch` service 
to the system and try to create a more complicated trace.

__NOTE:__ you must install `protoc` and run `./scripts/build_proto.sh` if you 
change the protobuf definitions.

## Thanks for playing, and welcome to OpenTracing!
Thanks for joining us in this walkthrough! Hope you enjoyed it. If you did, let 
us know, and consider spreading the love! 

A great way to get the feel for OpenTracing is to try your hand at instrumenting 
the OSS servers, frameworks, and client libraries that we all share. If you make
one, consider adding it to the growing ecosystem at 
http://github.com/opentracing-contrib. If you maintain a library yourself, 
plase consider adding built-in OT support.

We also need walkthroughs for languages other than Golang. Feel free to reuse 
the client, protobufs, and other assets from here if you'd like to make one.

For a more detailed explanation of OSS Instrumentation, check out the 
TurnKey Tracing proposal at http://bit.ly/turnkey-tracing.

_Aloha!_
