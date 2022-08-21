package utils

import (
	"grpc-scheduler/pkg/entity"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientConfig 获取client-go config,同时兼容集群内和集群外
func GetClientConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

// 自定义三目运算符函数
func Ternary(a bool, b, c float64) float64 {
	if a {
		return b
	}
	return c
}

func GetServiceIndex(Services []entity.Service, pod *v1.Pod) (serviceIndex int, roleIndex int, err error) {
	serviceName, ok := pod.ObjectMeta.Labels["servicename"]
	if !ok {
		return -1, -1, err
	}
	serviceIndex = -1
	for i, service := range Services {
		if service.ServiceName == serviceName {
			serviceIndex = i
			break
		}
	}
	if serviceIndex == -1 {
		var tmp entity.Service
		tmp.ServiceName = serviceName
		tmp.Roles = make([]entity.ServiceRole, 1)
		Services = append(Services, tmp)
		serviceIndex = len(Services) - 1
	}
	roleName := pod.ObjectMeta.Labels["rolename"]
	roleIndex = -1
	for i, role := range Services[serviceIndex].Roles {
		if role.RoleName == roleName {
			roleIndex = i
			break
		}
	}
	if roleIndex == -1 {
		var tmp entity.ServiceRole
		tmp.RoleName = roleName
		tmp.Values = make([]float64, 1)
		Services[serviceIndex].Roles = append(Services[serviceIndex].Roles, tmp)
		roleIndex = len(Services[serviceIndex].Roles) - 1
	}
	return serviceIndex, roleIndex, nil
}
