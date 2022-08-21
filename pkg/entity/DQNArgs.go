package entity

type DQNArgs struct {
	State     [][]float64
	Reward    float64
	Action    int32
	NextState [][]float64
	//  定义神经网络的输出与nodeName之间的映射
	NodeMap map[string]int32
}
