package main

import (
	"log"
	"net/http"
	"os"

	"micro-holtye/gen/gateway/v1/gatewayv1connect"
	"micro-holtye/internal/service/gateway"
)

func main() {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8080"
	}

	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://localhost:8081"
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8082"
	}

	store := gateway.NewStore(userServiceURL, orderServiceURL)
	service := gateway.NewService(store)
	handler := gateway.NewConnectHandler(service)

	mux := http.NewServeMux()
	path, h := gatewayv1connect.NewGatewayServiceHandler(handler)
	mux.Handle(path, h)

	log.Printf("Starting gateway service on %s", serverAddress)
	log.Printf("User service URL: %s", userServiceURL)
	log.Printf("Order service URL: %s", orderServiceURL)

	if err := http.ListenAndServe(serverAddress, mux); err != nil {
		log.Fatal(err)
	}
}
