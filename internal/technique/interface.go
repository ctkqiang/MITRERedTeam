package technique

import (
	"context"

	"mitre_red_team/internal/model"
)

// Technique 描述一个可执行的安全测试技术。
type Technique interface {
	// Execute 对 target 执行技术，返回结构化结果。
	Execute(ctx context.Context, target model.Target) (model.ExecutionResult, error)
}
