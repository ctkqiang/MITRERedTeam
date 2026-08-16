package engine

import (
	"context"
	"fmt"

	"mitre_red_team/internal/catalog"
	"mitre_red_team/internal/model"
	"mitre_red_team/internal/technique"
)

// Engine 编排目录查询与技术执行。
type Engine struct {
	catalogData *catalog.Catalog
}

// New 创建引擎。
func New(catalogData *catalog.Catalog) *Engine {
	return &Engine{catalogData: catalogData}
}

// Execute 按请求执行技术，返回结果列表。
// 指定 TechniqueID 执行单条；指定 TacticID 执行该战术下已实现的技术。
func (e *Engine) Execute(ctx context.Context, request model.ExecutionRequest) ([]model.ExecutionResult, error) {
	if request.TechniqueID != "" {
		result, err := e.executeTechnique(ctx, request.Target, request.TechniqueID)
		if err != nil {
			return nil, err
		}
		return []model.ExecutionResult{result}, nil
	}
	if request.TacticID != "" {
		return e.executeTactic(ctx, request.Target, request.TacticID)
	}
	return nil, fmt.Errorf("执行请求缺少技术或战术")
}

// ExecuteByMitre 执行映射到指定 MITRE ATT&CK ID 的所有已实现技术。
func (e *Engine) ExecuteByMitre(ctx context.Context, target model.Target, mitreID string) ([]model.ExecutionResult, error) {
	techniques := e.catalogData.TechniquesByMitreID(mitreID)
	if len(techniques) == 0 {
		return nil, fmt.Errorf("未找到映射到 MITRE %s 的技术", mitreID)
	}
	return e.executeTechniques(ctx, target, techniques)
}

// executeTechnique 执行单条技术：目录查询 + 注册表解析 + 执行。
func (e *Engine) executeTechnique(ctx context.Context, target model.Target, techniqueID string) (model.ExecutionResult, error) {
	metadata, found := e.catalogData.GetTechnique(techniqueID)
	if !found {
		return model.ExecutionResult{}, fmt.Errorf("技术 %s 不在目录中", techniqueID)
	}
	implementation, registered := technique.Get(metadata.Executor)
	if !registered {
		return model.ExecutionResult{}, fmt.Errorf("技术 %s 的执行器 %s 尚未实现", techniqueID, metadata.Executor)
	}
	return implementation.Execute(ctx, target)
}

// executeTactic 执行某战术下所有已实现的技术。
func (e *Engine) executeTactic(ctx context.Context, target model.Target, tacticID string) ([]model.ExecutionResult, error) {
	techniques := e.catalogData.TechniquesByTactic(tacticID)
	if len(techniques) == 0 {
		return nil, fmt.Errorf("战术 %s 下没有已实现的技术", tacticID)
	}
	return e.executeTechniques(ctx, target, techniques)
}

// executeTechniques 顺序执行一组已实现的技术。
// 规划中的技术允许未实现，跳过不中断；全部未实现时返回错误，不静默成功。
func (e *Engine) executeTechniques(ctx context.Context, target model.Target, techniques []model.Technique) ([]model.ExecutionResult, error) {
	var results []model.ExecutionResult
	for _, metadata := range techniques {
		implementation, registered := technique.Get(metadata.Executor)
		if !registered {
			continue
		}
		result, err := implementation.Execute(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("技术 %s 执行失败: %w", metadata.ID, err)
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("所选技术均未实现执行器，无法执行")
	}
	return results, nil
}
