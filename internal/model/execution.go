package model

// ExecutionRequest 描述用户发起的一次执行请求。
// TechniqueID 与 TacticID 二选一：指定单条技术，或指定整个战术下的所有技术。
type ExecutionRequest struct {
	Target      Target
	TechniqueID string
	TacticID    string
}

// ExecutionPlan 描述解析后的执行计划：针对目标要运行的技术列表。
type ExecutionPlan struct {
	Target     Target
	Techniques []Technique
}
