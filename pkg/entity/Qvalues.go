package entity

import "k8s.io/kubernetes/pkg/scheduler/framework"

type QValues struct {
	Values []int32
}

func (q *QValues) Clone() framework.StateData {
	return q
}

func NewQvalues(values []int32) *QValues {
	return &QValues{
		Values: values,
	}
}
