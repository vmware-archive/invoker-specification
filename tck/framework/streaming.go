package framework

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/projectriff/invoker-specification/tck/framework/rpc"
)

var Streaming = Suite{
	Name:        "s",
	Description: "Streaming Interaction",
	Port:        8081,
	Cases: []*Testcase{
		{
			Name:        "s-0001",
			Description: "MUST fail if first InputFrame is not StartFrame",
			Image:       "upper",
			T: func(port int) {
				timeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				conn, err := grpc.DialContext(timeout, fmt.Sprintf(":%d", port), grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					panic(err)
				}
				riffClient := rpc.NewRiffClient(conn)
				client, err := riffClient.Invoke(context.Background())
				if err != nil {
					panic(err)
				}
				sendData(client, 0, "text/plain", []byte("hello"))

				outputSignal, err := client.Recv()
				if err != nil {
					if grpcError, ok := status.FromError(err); ok && grpcError.Message() == "Expected first frame to be of type Start" {
					} else {
						panic(err)
					}
				} else {
					panic(fmt.Sprintf("Expected error, got data: %v", outputSignal))
				}

				// A subsequent invocation should not fail (checks that the whole handler did not crash)
				client, err = riffClient.Invoke(context.Background())
				if err != nil {
					panic(err)
				}
				firstSignal := rpc.InputSignal{
					Frame: &rpc.InputSignal_Start{
						Start: &rpc.StartFrame{
							ExpectedContentTypes: []string{"application/json"},
							InputNames:           []string{"words", "numbers"},
							OutputNames:          []string{"repeated"},
						},
					},
				}
				if err := client.Send(&firstSignal); err != nil {
					panic(err)
				}
				sendEOF(client)
				expectEOF(client)
			},
		},
		{
			Name:        "s-0002",
			Description: "MUST fail if subsequent InputFrame is a StartFrame",
			Image:       "upper",
			T: func(port int) {
				timeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				conn, err := grpc.DialContext(timeout, fmt.Sprintf(":%d", port), grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					panic(err)
				}
				riffClient := rpc.NewRiffClient(conn)
				client, err := riffClient.Invoke(context.Background())
				if err != nil {
					panic(err)
				}
				firstSignal := rpc.InputSignal{
					Frame: &rpc.InputSignal_Start{
						Start: &rpc.StartFrame{
							ExpectedContentTypes: []string{"application/json"},
							InputNames:           []string{"words", "numbers"},
							OutputNames:          []string{"repeated"},
						},
					},
				}
				if err := client.Send(&firstSignal); err != nil {
					panic(err)
				}
				secondSignal := rpc.InputSignal{
					Frame: &rpc.InputSignal_Start{
						Start: &rpc.StartFrame{
							ExpectedContentTypes: []string{"application/json"},
							InputNames:           []string{"words", "numbers"},
							OutputNames:          []string{"repeated"},
						},
					},
				}
				if err := client.Send(&secondSignal); err != nil {
					panic(err)
				}
				outputSignal, err := client.Recv()
				if err != nil {
					if grpcError, ok := status.FromError(err); ok && grpcError.Message() == "Expected first frame to be of type Start" {
					} else {
						panic(err)
					}
				} else {
					panic(fmt.Sprintf("Expected error, got data: %v", outputSignal))
				}

				// A subsequent invocation should not fail (checks that the whole handler did not crash)
				client, err = riffClient.Invoke(context.Background())
				if err != nil {
					panic(err)
				}
				firstSignal = rpc.InputSignal{
					Frame: &rpc.InputSignal_Start{
						Start: &rpc.StartFrame{
							ExpectedContentTypes: []string{"application/json"},
							InputNames:           []string{"words", "numbers"},
							OutputNames:          []string{"repeated"},
						},
					},
				}
				if err := client.Send(&firstSignal); err != nil {
					panic(err)
				}
				outputSignal, err = client.Recv()
				if err != nil {
					panic(err)
				}
			},
		},
		{
			Name:        "s-0003",
			Description: "MUST honor the expectedContentTypes header",
			Image:       "repeater",
			T: func(port int) {
				timeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				cancel()
				conn, err := grpc.DialContext(timeout, fmt.Sprintf(":%d", port), grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					panic(err)
				}
				riffClient := rpc.NewRiffClient(conn)
				client, err := riffClient.Invoke(context.Background())
				if err != nil {
					panic(err)
				}
				firstSignal := rpc.InputSignal{
					Frame: &rpc.InputSignal_Start{
						Start: &rpc.StartFrame{
							ExpectedContentTypes: []string{"text/plain"},
							InputNames:           []string{"words", "numbers"},
							OutputNames:          []string{"repeated"},
						},
					},
				}
				if err := client.Send(&firstSignal); err != nil {
					panic(err)
				}

				sendData(client, 0, "text/plain", []byte("hello"))
				sendData(client, 1, "application/json", []byte("2"))

				expectData(client, "text/plain", []byte("hello"))
				expectData(client, "text/plain", []byte("hello"))
				sendEOF(client)
				expectEOF(client)

				client, err = riffClient.Invoke(context.Background())
				if err != nil {
					panic(err)
				}
				firstSignal = rpc.InputSignal{
					Frame: &rpc.InputSignal_Start{
						Start: &rpc.StartFrame{
							ExpectedContentTypes: []string{"application/json"},
							InputNames:           []string{"words", "numbers"},
							OutputNames:          []string{"repeated"},
						},
					},
				}
				if err := client.Send(&firstSignal); err != nil {
					panic(err)
				}

				sendData(client, 0, "text/plain", []byte("hello"))
				sendData(client, 1, "application/json", []byte("2"))

				expectData(client, "application/json", []byte(`"hello"`))
				expectData(client, "application/json", []byte(`"hello"`))
				sendEOF(client)
				expectEOF(client)

			},
		},
	},
}

func expectEOF(client rpc.Riff_InvokeClient) {
	if f, err := client.Recv(); err == nil || err != io.EOF {
		panic(fmt.Sprintf("Expected to receive EOF, either got data (%v) or a different error (%v)", f, err))
	}

}

func sendEOF(client rpc.Riff_InvokeClient) {
	if err := client.CloseSend(); err != nil {
		panic(err)
	}
}

func sendData(client rpc.Riff_InvokeClient, index int32, ct string, b []byte) {
	signal := rpc.InputSignal{
		Frame: &rpc.InputSignal_Data{
			Data: &rpc.InputFrame{
				Payload:     b,
				ContentType: ct,
				Headers:     nil,
				ArgIndex:    index,
			},
		},
	}
	if err := client.Send(&signal); err != nil {
		panic(err)
	}
}

func expectData(client rpc.Riff_InvokeClient, ct string, b []byte) {
	outputSignal, err := client.Recv()
	if err != nil {
		panic(err)
	} else if outputSignal.GetData().ContentType != ct {
		panic(fmt.Sprintf("Expected DataFrame contentType to be %v, got %v", ct, outputSignal.GetData().ContentType))
	} else if !reflect.DeepEqual(outputSignal.GetData().Payload, b) {
		panic(fmt.Sprintf("Expected DataFrame payload to be %v, got %v", outputSignal.GetData().Payload, b))
	}
}
