package agent

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"mitre_red_team/internal/catalog"
	"mitre_red_team/internal/engine"
	"mitre_red_team/internal/llm"
	"mitre_red_team/internal/model"
	"mitre_red_team/internal/technique"
	"mitre_red_team/internal/utilities"
)

// maxRounds 是 AI 模式下自动推进的最大轮数，防止无界编排执行。
const maxRounds = 3

// RunParams 描述一次 AI 辅助执行所需的依赖，便于测试注入随机源与客户端工厂。
type RunParams struct {
	Engine      *engine.Engine
	CatalogData *catalog.Catalog
	Target      model.Target
	Request     model.ExecutionRequest
	// Initial 返回首轮执行结果；为空时回退到 Engine.Execute(Request)。
	// 通过它支持 --mitre 等 Engine.Execute 未覆盖的初始执行方式。
	Initial   func(ctx context.Context) ([]model.ExecutionResult, error)
	Providers []model.LLMModel
	Logger    *utilities.Logger
	RandIntn  func(int) int
	NewClient func(model.LLMModel) (llm.Client, error)
}

// Run 执行一次 AI 辅助 TTP 循环：
// 1. 从已配置的供应商中随机选择一个；
// 2. 执行用户请求的初始技术；
// 3. 把执行结果交给 LLM 分析，解析其建议的下一步技术并自动执行；
// 4. 最多推进 maxRounds 轮，建议为空、无效或执行失败即停止。
// 目标始终限定为用户声明的 Target，不扩大攻击范围。
func Run(ctx context.Context, params RunParams) ([]model.ExecutionResult, error) {
	logger := params.Logger
	if logger == nil {
		logger = utilities.Default
	}
	if params.RandIntn == nil {
		params.RandIntn = secureRandomInt
	}
	if params.NewClient == nil {
		params.NewClient = func(provider model.LLMModel) (llm.Client, error) {
			return llm.FromEnv(provider, logger, nil)
		}
	}

	provider, err := pickRandomProvider(params.Providers, params.RandIntn)
	if err != nil {
		return nil, err
	}
	logger.Info("agent.provider", nil, fmt.Sprintf("随机选择 LLM 供应商 %s", provider))

	client, err := params.NewClient(provider)
	if err != nil {
		return nil, fmt.Errorf("初始化 LLM 客户端: %w", err)
	}

	var history []model.ExecutionResult
	switch {
	case params.Initial != nil:
		history, err = params.Initial(ctx)
	case params.Engine != nil:
		history, err = params.Engine.Execute(ctx, params.Request)
	default:
		err = errors.New("缺少初始执行入口：请提供 Initial 或 Engine")
	}
	if err != nil {
		return nil, fmt.Errorf("初始技术执行失败: %w", err)
	}

	for round := 1; round <= maxRounds; round++ {
		systemPrompt, userPrompt := BuildDecisionPrompt(params.Target, params.CatalogData.Techniques, history, round, maxRounds)
		output, err := client.Complete(ctx, systemPrompt, userPrompt)
		if err != nil {
			return nil, fmt.Errorf("第 %d 轮 LLM 调用失败: %w", round, err)
		}
		decision, err := ParseDecision(output)
		if err != nil {
			return nil, fmt.Errorf("第 %d 轮 LLM 建议解析失败: %w", round, err)
		}
		if strings.TrimSpace(decision.NextTechniqueID) == "" {
			logger.Info("agent.stop", nil, "LLM 建议停止，不再推进")
			break
		}

		metadata, found := params.CatalogData.GetTechnique(decision.NextTechniqueID)
		if !found {
			return nil, fmt.Errorf("LLM 建议的技术 %s 不在目录中", decision.NextTechniqueID)
		}
		// 技术存在于目录但执行器未实现时，记录警告并停止推进，不把缺失实现当致命错误。
		if _, registered := technique.Get(metadata.Executor); !registered {
			logger.Warn("agent.skipped", nil, fmt.Sprintf("技术 %s（执行器 %s）尚未实现，停止推进", metadata.ID, metadata.Executor))
			break
		}
		nextResults, err := params.Engine.Execute(ctx, model.ExecutionRequest{
			Target:      params.Target,
			TechniqueID: decision.NextTechniqueID,
		})
		if err != nil {
			return nil, fmt.Errorf("执行 LLM 建议的技术 %s: %w", decision.NextTechniqueID, err)
		}
		logger.Info("agent.round", nil, fmt.Sprintf("第 %d 轮执行 %s：%s", round, metadata.ID, decision.Rationale))
		history = append(history, nextResults...)
	}
	return history, nil
}

// pickRandomProvider 从已配置的供应商中随机选取一个。
// providers 为空时返回错误，提示先配置 API 凭据。
func pickRandomProvider(providers []model.LLMModel, randIntn func(int) int) (model.LLMModel, error) {
	if len(providers) == 0 {
		return "", errors.New("未配置任何 LLM 供应商，请在 .env 中设置对应 API Key")
	}
	return providers[randIntn(len(providers))], nil
}

// secureRandomInt 基于 crypto/rand 返回 [0, max) 的随机整数，
// 避免使用可预测的伪随机源参与执行决策。
func secureRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	var buffer [8]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return 0
	}
	return int(binary.BigEndian.Uint64(buffer[:]) % uint64(max))
}
