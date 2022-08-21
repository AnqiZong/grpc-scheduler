package plugin

import (
	"context"
	"fmt"
	"grpc-scheduler/pkg/entity"
	"grpc-scheduler/pkg/remoteClient"
	"grpc-scheduler/pkg/utils"
	"math"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	runtime2 "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	metricsClientSet "k8s.io/metrics/pkg/client/clientset/versioned"
)

//  定义使用到的常量字符串
const (
	Name             = "GRPCScheduler"
	preScoreStateKey = "PreScore" + Name
)

//  定义连接prometheus 使用的参数，该值从配置文件获取
type clientArg struct {
	PrometheusAddr   string `json:"prometheus_address,omitempty"`
	RemoteServerAddr string `json:"remote_server_address,omitempty"`
}

// GRPCSchedulerPluginArg DRLScheduler 配置文件参数
type GRPCSchedulerPluginArg struct {
	PrometheusClient   utils.PromClient            // 建立Prometheus client
	MetricsClientSet   *metricsClientSet.Clientset // prometheus未安装时的 metriccs client
	RemoteServerClient remoteClient.GRPCClient
}

// DRLSchedulerPlugin 使用的参数
type GRPCSchedulerPlugin struct {
	handle   framework.Handle
	args     GRPCSchedulerPluginArg
	DQNargs  entity.DQNArgs
	Services []entity.Service
}

func (n *GRPCSchedulerPlugin) Name() string {
	return Name
}

//在PreScore阶段，完成状态的保存、神经网络的训练，并计算出所有动作的Q值保存到数组中，供Score阶段使用
func (n *GRPCSchedulerPlugin) PreScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	if len(nodes) == 0 {
		return nil
	}
	n.DQNargs.NextState, _ = n.args.PrometheusClient.QueryClusterMetrics(n.DQNargs.NodeMap)
	// 获取预测的qValues
	labels := make([]string, 2)
	labels[0] = pod.ObjectMeta.Labels["servicename"]
	labels[1] = pod.ObjectMeta.Labels["servicerolename"]
	filernode := make([]string, len(nodes))
	index := 0
	for _, node := range nodes {
		filernode[index] = node.ObjectMeta.Name
		index++
	}
	prediction, err := n.args.RemoteServerClient.GetQValues(n.DQNargs, labels, filernode)
	qValues := entity.NewQvalues(prediction)
	n.DQNargs.State = n.DQNargs.NextState
	if err != nil {
		panic(err)
	}
	// 把获取的qValues 保存
	cycleState.Write(preScoreStateKey, qValues)
	return nil
}

func (n *GRPCSchedulerPlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	index := n.DQNargs.NodeMap[nodeName]
	prediction, err := state.Read(preScoreStateKey)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("reading %q from cycleState: %w", preScoreStateKey, err))
	}
	qValues, ok := prediction.(*entity.QValues)
	if !ok {
		return 0, framework.AsStatus(fmt.Errorf("cannot convert saved state to tensor.Dense"))
	}
	return int64(qValues.Values[index]), nil
}

func (n *GRPCSchedulerPlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var maxScore int64 = -math.MaxInt64
	var maxIndex int = -1
	for i := range scores {
		score := scores[i].Score
		if score > maxScore {
			maxScore = score
			maxIndex = i
		}
	}
	n.DQNargs.Action = int32(maxIndex)
	n.DQNargs.Reward, _ = utils.GetReward(n.DQNargs.State, n.Services, pod)
	serviceIndex, roleIndex, err := utils.GetServiceIndex(n.Services, pod)
	if err != nil {
		return framework.AsStatus(fmt.Errorf("find serviceIndex err: %w", err))
	}
	n.Services[serviceIndex].Roles[roleIndex].Values = append(n.Services[serviceIndex].Roles[roleIndex].Values, n.DQNargs.Reward)
	for i := range scores {
		scores[i].Score = int64(scores[i].Score / maxScore * 100)
	}
	return nil
}

func (n *GRPCSchedulerPlugin) ScoreExtensions() framework.ScoreExtensions {
	return n
}

func New(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	args := &GRPCSchedulerPluginArg{}
	clientarg := &clientArg{}
	err := runtime2.DecodeInto(configuration, clientarg)
	if err != nil {
		return nil, err
	}
	args.PrometheusClient, _ = utils.NewPromClient(clientarg.PrometheusAddr)
	// 获取K8S .kube/config配置信息
	config, err := utils.GetClientConfig()
	if err != nil {
		return nil, err
	}
	mcs, err := metricsClientSet.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	args.MetricsClientSet = mcs
	args.RemoteServerClient, err = remoteClient.NewGRPCClient(clientarg.RemoteServerAddr)
	if err != nil {
		return nil, err
	}
	//agent := dqn.NewAgent(dqn.DefaultAgentConfig)
	dqnargs := &entity.DQNArgs{}
	//var services []Service
	k8sclient, err := kubernetes.NewForConfig(config)
	nodeList, err := k8sclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	dqnargs.NodeMap = make(map[string]int32)
	var index int32 = 0
	for _, node := range nodeList.Items {
		dqnargs.NodeMap[node.Name] = index
		index++
	}
	dqnargs.Action = -1
	dqnargs.Reward = 0.0
	return &GRPCSchedulerPlugin{
		handle:  f,
		args:    *args,
		DQNargs: *dqnargs,
	}, nil
}
