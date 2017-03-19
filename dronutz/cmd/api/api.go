package main

import (
	"flag"
	"fmt"
	"net/http"

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

	conn, err := grpc.Dial(
		cfg.KitchenAddress(),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}

	kitchenClient := dronutz.NewKitchenClient(conn)
	service := dronutz.NewAPIService(cfg, kitchenClient)

	fmt.Println("Api server listening on", cfg.APIAddress())

	err = http.ListenAndServe(
		cfg.APIAddress(),
		service.ServeMux(),
	)

	fmt.Println("Api server exited:", err)
}
