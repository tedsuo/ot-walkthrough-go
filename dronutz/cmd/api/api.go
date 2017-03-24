package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/tedsuo/ot-walkthrough-go/dronutz"

	"google.golang.org/grpc"
)

var configPath = flag.String("config", "config_example.yml", "path to configuration file")

func init() {
	flag.Parse()
}

func main() {
	fmt.Println("🍩🍩🍩🍩🍩 API 🍩🍩🍩🍩🍩")

	cfg, err := dronutz.NewConfigFromPath(*configPath)
	if err != nil {
		panic(err)
	}

	err = dronutz.ConfigureGlobalTracer(cfg, "api")
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(
		cfg.KitchenAddress(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()),
		),
	)
	if err != nil {
		panic(err)
	}

	kitchenClient := dronutz.NewKitchenClient(conn)
	service := dronutz.NewAPIService(cfg, kitchenClient)

	fmt.Println("Api server listening on", cfg.APIAddress())

	err = http.ListenAndServe(
		cfg.APIAddress(),
		nethttp.Middleware(
			opentracing.GlobalTracer(),
			service.ServeMux(),
			nethttp.OperationNameFunc(func(req *http.Request) string {
				return "/dronutz.API" + req.URL.Path
			}),
		),
	)

	fmt.Println("Api server exited:", err)
}
