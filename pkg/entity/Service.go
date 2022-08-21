package entity

//  保存相同服务的冲突值
type ServiceRole struct {
	RoleName string
	Values   []float64
}

// 服务名 从标签中获取该信息
type Service struct {
	ServiceName string
	Roles       []ServiceRole
}
