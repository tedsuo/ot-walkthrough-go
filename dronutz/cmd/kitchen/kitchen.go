package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/tedsuo/ot-walkthrough-go/dronutz"

	"google.golang.org/grpc"
)

var configPath = flag.String("config", "config_example.yml", "path to configuration file")

func init() {
	flag.Parse()
}

func main() {
	fmt.Println("🍩🍩🍩🍩 Kitchen 🍩🍩🍩🍩")

	cfg, err := dronutz.NewConfigFromPath(*configPath)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()

	service := dronutz.NewKitchenService(cfg)
	dronutz.RegisterKitchenServer(server, service)

	lis, err := net.Listen("tcp", cfg.KitchenAddress())
	if err != nil {
		panic(err)
	}

	fmt.Println("Kitchen server listening on", cfg.KitchenAddress())
	err = server.Serve(lis)
	fmt.Println("Kitchen server exited:", err)
}
