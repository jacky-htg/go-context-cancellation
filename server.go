package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go-cancel/pb/cities"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RpcServer struct {
	Grpc *grpc.Server
}

func NewServer() *RpcServer {
	gs := grpc.NewServer()
	return &RpcServer{
		Grpc: gs,
	}
}

func main() {
	if err := run(); err != nil {
		log.Printf("error: shutting down: %s", err)
		os.Exit(1)
	}
}

func run() error {
	port := map[string]string{"grpc": "9099", "rest": "8099"}
	errorServer := make(chan error)

	rpcServer := NewServer()
	cities.RegisterCitiesServiceServer(rpcServer.Grpc, &citiesServer{})

	go func() {
		errorServer <- runRpcServer(port["grpc"], rpcServer)
	}()

	go func() {
		errorServer <- runRestServer(port["rest"], rpcServer)
	}()

	select {
	case err := <-errorServer:
		if err != nil {
			return err
		}
	}

	return nil
}

func runRpcServer(port string, rpcServer *RpcServer) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	if err := rpcServer.Grpc.Serve(listener); err != nil {
		return err
	}
	return nil
}

func runRestServer(httpPort string, rpcServer *RpcServer) error {
	handler := http.HandlerFunc(rest)

	if err := http.ListenAndServe(":"+httpPort, handler); err != nil {
		return err
	}

	return nil
}

func rest(w http.ResponseWriter, r *http.Request) {
	list, err := new(citiesServer).List(r.Context(), &cities.EmptyMessage{})
	if st, ok := status.FromError(err); err != nil && ok {
		err = fmt.Errorf(st.Message())
	}

	if err != nil {
		log.Println("error get list city", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(list.City)
	if err != nil {
		log.Println("error marshalling result", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Println("error writing result", err)
	}
}

type citiesServer struct{}

func (u *citiesServer) ListStream(in *cities.EmptyMessage, stream cities.CitiesService_ListStreamServer) error {
	ctx := stream.Context()
	select {
	case <-ctx.Done():
		return contextError(ctx)
	default:
	}

	for i := 1; i < 50; i++ {
		println(i)
		time.Sleep(1 * time.Second)

		res := &cities.CityStream{
			City: &cities.City{Id: uint32(i), Name: randSeq(10)},
		}

		if err := stream.Send(res); err != nil {
			return status.Errorf(codes.Unknown, "cannot send stream response: %v", err)
		}
	}

	println("tes")

	return nil
}

func (u *citiesServer) List(ctx context.Context, in *cities.EmptyMessage) (*cities.Cities, error) {
	/*select {
	case <-ctx.Done():
		return nil, contextError(ctx)
	default:
	} */

	var list []*cities.City
	for i := 1; i < 50; i++ {
		err := contextError(ctx)
		if err != nil {
			return nil, err
		}
		list = append(list, &cities.City{Id: uint32(i), Name: randSeq(10)})
		time.Sleep(100 * time.Millisecond)
		println(i)
	}

	err := contextError(ctx)
	if err != nil {
		return nil, err
	}

	for i := 1; i < 10; i++ {
		println(i)
	}

	return &cities.Cities{City: list}, nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return strings.Title(string(b))
}
