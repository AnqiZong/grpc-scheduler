package utils

import (
	"grpc-scheduler/pkg/entity"

	v1 "k8s.io/api/core/v1"
)

func GetReward(State [][]float64, Services []entity.Service, pod *v1.Pod) (float64, error) {

}
