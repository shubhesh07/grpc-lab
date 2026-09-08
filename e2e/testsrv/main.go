// Minimal reflecting gRPC server for grpc-lab end-to-end tests: the interop
// TestService (unary + all three streaming shapes) and Health.
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/interop"
	testgrpc "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50077")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	testgrpc.RegisterTestServiceServer(s, interop.NewTestServer())
	healthpb.RegisterHealthServer(s, health.NewServer())
	reflection.Register(s)
	log.Println("testsrv on 127.0.0.1:50077")
	log.Fatal(s.Serve(lis))
}
