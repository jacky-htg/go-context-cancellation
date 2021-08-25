package main

import (
	"fmt"
	"go-cancel/pb/cities"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	ctx := context.Background()
	// ctx, cancel := context.WithDeadline(ctx, time.Now().Add(3*time.Second))
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9099", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("did not connect: %s", err)
		return
	}
	defer conn.Close()

	city := cities.NewCitiesServiceClient(conn)

	err = callStream(ctx, city)
	if st, ok := status.FromError(err); err != nil && ok {
		err = fmt.Errorf(st.Message())
	}

	if err != nil {
		fmt.Printf("Error when calling grpc service: %s", err)
		return
	}
}

func callStream(ctx context.Context, city cities.CitiesServiceClient) error {
	stream, err := city.ListStream(ctx, &cities.EmptyMessage{})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("end stream")
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("Resp : %v", resp.GetCity())
		println()
	}

	return nil
}
