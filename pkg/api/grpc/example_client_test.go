package grpc_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/mszostok/job-runner/pkg/api/grpc"
)

const (
	svrAddrEnv = "JOB_GRPC_BACKEND_ADDR"
	jobName    = "example-run"
)

// This test illustrates how to use gRPC Go client against real gRPC Agent server.
//
// NOTE: Before running this test, make sure that the Agent is running under the `srvAddr` address.
//
// To run this test, execute:
// JOB_GRPC_BACKEND_ADDR=":50051" go test ./pkg/api/grpc/ -run "^ExampleJobServiceClient" -v -count 1
func ExampleJobServiceClient() {
	srvAddr := os.Getenv(svrAddrEnv)
	if srvAddr == "" {
		return
	}

	conn, err := grpc.Dial(srvAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	fatalOnErr(err)
	defer conn.Close()

	ctx := context.TODO()
	client := pb.NewJobServiceClient(conn)

	_, err = client.Run(ctx, &pb.RunRequest{
		Name:    jobName,
		Command: "sh",
		Args:    []string{"-c", "sleep 1 && echo $MOTTO"},
		Env:     []string{"MOTTO=hakuna_matata"},
	})
	fatalOnErr(err)

	getOut, err := client.Get(ctx, &pb.GetRequest{Name: jobName})
	fatalOnErr(err)

	fmt.Printf("'Get' of just started Job: Job create by %q status %q, exit code %d\n\n", getOut.CreatedBy, getOut.Status, getOut.ExitCode)

	stream, err := client.StreamLogs(ctx, &pb.StreamLogsRequest{Name: jobName})
	fatalOnErr(err)

	fmt.Println("Stream logs:")
	err = pb.ForwardStreamLogs(os.Stdout, stream)
	fatalOnErr(err)

	getOut, err = client.Get(ctx, &pb.GetRequest{Name: jobName})
	fatalOnErr(err)

	fmt.Println()
	fmt.Printf("'Get' of finished Job: Job create by %q status %q, exit code %d\n", getOut.CreatedBy, getOut.Status, getOut.ExitCode)
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
