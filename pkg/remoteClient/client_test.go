package remoteClient

import (
	"fmt"
	"grpc-scheduler/pkg/entity"
	"testing"
)

func TestGrpc(t *testing.T) {
	var grpcAddress = "127.0.0.1:50051"
	GRPCClient, err := NewGRPCClient(grpcAddress)
	if err != nil {
		panic(err)
	}
	var dqnargs entity.DQNArgs
	qValues, err := GRPCClient.GetQValues(dqnargs, nil, nil)
	if err != nil {
		panic(err)
	}
	for _, q := range qValues {
		fmt.Println(q)
	}
}
