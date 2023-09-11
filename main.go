package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	//"github.com/helloworld-grpc-gateway/proto/services"

	"github.com/iamrajiv/helloworld-grpc-gateway/postgres"
	services "github.com/iamrajiv/helloworld-grpc-gateway/proto/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server struct representing our service implementation
type server struct {
	// Embed the unimplemented server
	services.UnimplementedBookServer
}

// SayHello is the implementation of the SayHello method defined in the proto file
//func (*server) SayHello(_ context.Context, in *pbHelloWorld.HelloRequest) (*pbHelloWorld.HelloReply, error) {
//	return &pbHelloWorld.HelloReply{Message: in.Name + " world"}, nil
//}

func (*server) Create(context.Context, *services.CreateRequest) (*services.Response, error) {
	out := new(services.Response)
	var err error
	return out, err
}

func (*server) Get(con context.Context, req *services.Request) (*services.Response, error) {
	conn := postgres.GetConnection()
	log.Println(req.Id[6:])
	book := conn.GetBook(req.Id[6:])
	resp := services.Response{}
	resp.Book = &book
	var err error
	return &resp, err
}

func (*server) List(context.Context, *services.ListRequest) (*services.ListResponse, error) {
	conn := postgres.GetConnection()
	books := conn.ListBooks()
	resp := services.ListResponse{}
	resp.Books = books //append(resp.Books, books)
	var err error
	return &resp, err
}

func main() {

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	services.RegisterBookServer(s, &server{})
	//pbHelloWorld.RegisterGreeterServer(s, &server{})
	// Serve gRPC server
	log.Println("Serving gRPC on 0.0.0.0:8080")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8080",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	// Create a new ServeMux for the gRPC-Gateway
	gwmux := runtime.NewServeMux()
	// Register the Greeter service with the gRPC-Gateway
	err = services.RegisterBookHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	// Create a new HTTP server for the gRPC-Gateway
	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(gwServer.ListenAndServe())
}
