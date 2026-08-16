package model

// Tactic 描述一个漏洞赏金测试战术。
type Tactic struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Techniques  []string `json:"techniques"`
}
