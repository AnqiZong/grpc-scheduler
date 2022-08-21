package remoteClient

import (
	"context"

	"grpc-scheduler/pkg/entity"
	"grpc-scheduler/pkg/proto"

	"google.golang.org/grpc"
)

type GRPCClient interface {
	GetQValues(dqnargs entity.DQNArgs, labels []string, filterNode []string) ([]int32, error)
	//GetQValues(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	//QueryClusterMetrics([]*kv1.Node) ([][]float64, error)
}
type grpcClient struct {
	GrpcClient proto.UseDQNClient
}

func (grpc *grpcClient) GetQValues(dqnargs entity.DQNArgs, labels []string, filterNode []string) ([]int32, error) {
	state := proto.Request{}.State
	nextState := proto.Request{}.NextState
	for i := 0; i < len(dqnargs.State); i++ {
		tmp1 := proto.RequestNodeMatric{}
		tmp2 := proto.RequestNodeMatric{}
		for j := 0; j < len(dqnargs.State[i]); j++ {
			tmp1.Metric[j] = dqnargs.State[i][j]
			tmp2.Metric[j] = dqnargs.NextState[i][j]
		}
		state[i] = &tmp1
		nextState[i] = &tmp2
	}
	request := proto.Request{
		State:       state,
		Reward:      dqnargs.Reward,
		Action:      dqnargs.Action,
		NextState:   nextState,
		Labels:      labels,
		FilterNodes: filterNode,
		NodeMap:     dqnargs.NodeMap,
	}
	res, err := grpc.GrpcClient.GetQValues(context.Background(), &request)
	if err != nil {
		panic(err)
	}
	return res.GetQValues(), nil
}

func NewGRPCClient(grpcAddress string) (GRPCClient, error) {
	grpcConn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	// 保存连接，连接不断开 。。。
	//defer grpcConn.Close()
	// 成功dial grpcServer之后,获取grpcServer的client对象
	return &grpcClient{
		GrpcClient: proto.NewUseDQNClient(grpcConn),
	}, nil
}
