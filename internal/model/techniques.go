package model

// TechniqueMode 定义技术的执行模式。
type TechniqueMode string

const (
	TechniquePassive TechniqueMode = "passive"
	TechniqueActive  TechniqueMode = "active"
	TechniqueManual  TechniqueMode = "manual"
)

// Technique 描述一个漏洞赏金测试技术。
type Technique struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	TacticID    string        `json:"tactic_id"`
	Description string        `json:"description"`
	Executor    string        `json:"executor"`
	Mode        TechniqueMode `json:"mode"`
	Tools       []string      `json:"tools"`
	MITRE       []string      `json:"mitre"`
	Tags        []string      `json:"tags"`
}
