package tests

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"mitre_red_team/internal/agent"
	"mitre_red_team/internal/catalog"
	"mitre_red_team/internal/engine"
	"mitre_red_team/internal/llm"
	"mitre_red_team/internal/model"
	"mitre_red_team/internal/technique"
	"mitre_red_team/internal/utilities"
)

// stubTechnique 是测试用的假技术实现，注册到目录枚举执行器后由 engine 调度。
type stubTechnique struct{}

// Execute 返回固定成功结果，不产生任何副作用。
func (stubTechnique) Execute(context.Context, model.Target) (model.ExecutionResult, error) {
	return model.ExecutionResult{
		TechniqueID: "BB05.001",
		Status:      model.StatusSucceeded,
		Summary:     "stub 执行完成",
	}, nil
}

// scriptedClient 按预定顺序返回 LLM 输出，用于验证 agent 循环的分支。
type scriptedClient struct {
	outputs []string
	calls   int
}

// Complete 依次返回脚本中的输出，耗尽后报错。
func (c *scriptedClient) Complete(context.Context, string, string) (string, error) {
	if c.calls >= len(c.outputs) {
		return "", errors.New("脚本输出已用尽")
	}
	output := c.outputs[c.calls]
	c.calls++
	return output, nil
}

// Provider 固定返回 DeepSeek，测试不关心具体值。
func (c *scriptedClient) Provider() model.LLMModel {
	return model.DeepSeek
}

// 验证纯 JSON 输出能被正确解析。
func TestParseDecisionPlain(t *testing.T) {
	decision, err := agent.ParseDecision(`{"next_technique_id":"BB05.003","rationale":"继续探测端点"}`)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if decision.NextTechniqueID != "BB05.003" {
		t.Errorf("期望 BB05.003，实际 %s", decision.NextTechniqueID)
	}
	if !strings.Contains(decision.Rationale, "端点") {
		t.Errorf("rationale 应保留原文，实际 %s", decision.Rationale)
	}
}

// 验证被代码块包裹的输出也能解析。
func TestParseDecisionCodeFence(t *testing.T) {
	output := "```json\n{\"next_technique_id\":\"BB05.001\"}\n```"
	decision, err := agent.ParseDecision(output)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if decision.NextTechniqueID != "BB05.001" {
		t.Errorf("期望 BB05.001，实际 %s", decision.NextTechniqueID)
	}
}

// 验证空 next_technique_id 表示建议停止。
func TestParseDecisionStop(t *testing.T) {
	decision, err := agent.ParseDecision(`{"next_technique_id":"","rationale":"已充分"}`)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if decision.NextTechniqueID != "" {
		t.Errorf("期望空建议，实际 %s", decision.NextTechniqueID)
	}
}

// 验证无法解析的输出返回错误。
func TestParseDecisionInvalid(t *testing.T) {
	if _, err := agent.ParseDecision("这不是 JSON"); err == nil {
		t.Fatal("期望解析失败，实际无错误")
	}
}

// 验证决策提示词包含目标、候选技术与执行历史。
func TestBuildDecisionPrompt(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}
	systemPrompt, userPrompt := agent.BuildDecisionPrompt(
		model.Target{Host: "example.com", Scheme: "https"},
		catalogData.Techniques,
		[]model.ExecutionResult{{TechniqueID: "BB05.001", Status: model.StatusSucceeded, Summary: "发现 40 个命中"}},
		1, 3,
	)
	if !strings.Contains(userPrompt, "example.com") {
		t.Error("用户提示词应包含目标")
	}
	if !strings.Contains(systemPrompt, "BB05.001") {
		t.Error("系统提示词应包含候选技术列表")
	}
	if !strings.Contains(userPrompt, "BB05.001") {
		t.Error("用户提示词应包含执行历史")
	}
	if !strings.Contains(userPrompt, "第 1 轮") {
		t.Error("用户提示词应包含当前轮数")
	}
}

// 验证未配置任何供应商时 Run 返回错误。
func TestRunNoProviders(t *testing.T) {
	_, err := agent.Run(context.Background(), agent.RunParams{Providers: nil})
	if err == nil {
		t.Fatal("期望无供应商报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "未配置任何 LLM 供应商") {
		t.Errorf("错误信息应提示配置 API Key，实际: %v", err)
	}
}

// 验证完整循环：初始执行 → LLM 建议下一步 → 自动执行 → LLM 建议停止。
func TestRunLoop(t *testing.T) {
	technique.Register("directory-enumeration", stubTechnique{})
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}
	executionEngine := engine.New(catalogData)

	var logBuffer bytes.Buffer
	logger := utilities.New(&logBuffer)
	client := &scriptedClient{outputs: []string{
		`{"next_technique_id":"BB05.001","rationale":"继续验证"}`,
		`{"next_technique_id":"","rationale":"已充分"}`,
	}}

	results, err := agent.Run(context.Background(), agent.RunParams{
		Engine:      executionEngine,
		CatalogData: catalogData,
		Target:      model.Target{Host: "example.com", Scheme: "https"},
		Request:     model.ExecutionRequest{TechniqueID: "BB05.001"},
		Providers:   []model.LLMModel{model.Openai, model.DeepSeek},
		Logger:      logger,
		RandIntn:    func(int) int { return 1 },
		NewClient:   func(model.LLMModel) (llm.Client, error) { return client, nil },
	})
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	// 初始 1 条 + LLM 建议后执行 1 条，随后建议停止。
	if len(results) != 2 {
		t.Fatalf("期望 2 条结果，实际 %d", len(results))
	}
	// RandIntn(1) 应选中 providers 中第二个供应商 deepseek。
	if !strings.Contains(logBuffer.String(), "随机选择 LLM 供应商 deepseek") {
		t.Errorf("日志应记录选中的供应商，实际: %s", logBuffer.String())
	}
	if client.calls != 2 {
		t.Errorf("期望 LLM 被调用 2 次，实际 %d", client.calls)
	}
}

// 验证 RandIntn 未注入时使用内置安全随机源选择供应商。
func TestRunUsesSecureRandom(t *testing.T) {
	technique.Register("directory-enumeration", stubTechnique{})
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}
	executionEngine := engine.New(catalogData)
	client := &scriptedClient{outputs: []string{`{"next_technique_id":"","rationale":"停止"}`}}

	results, err := agent.Run(context.Background(), agent.RunParams{
		Engine:      executionEngine,
		CatalogData: catalogData,
		Target:      model.Target{Host: "example.com", Scheme: "https"},
		Request:     model.ExecutionRequest{TechniqueID: "BB05.001"},
		Providers:   []model.LLMModel{model.Anthropic},
		Logger:      utilities.Default,
		NewClient:   func(model.LLMModel) (llm.Client, error) { return client, nil },
	})
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	// 单个供应商 + 内置随机源应正常完成初始执行并建议停止。
	if len(results) != 1 {
		t.Fatalf("期望 1 条结果，实际 %d", len(results))
	}
}

// 验证 LLM 建议的技术执行器未实现时，记录警告并优雅停止，不报错。
func TestRunUnimplementedSuggestion(t *testing.T) {
	technique.Register("directory-enumeration", stubTechnique{})
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}
	executionEngine := engine.New(catalogData)
	var logBuffer bytes.Buffer
	logger := utilities.New(&logBuffer)
	// BB01.003 的子域名枚举执行器 subdomain-enumeration 未注册实现。
	client := &scriptedClient{outputs: []string{
		`{"next_technique_id":"BB01.003","rationale":"枚举子域名"}`,
	}}

	results, err := agent.Run(context.Background(), agent.RunParams{
		Engine:      executionEngine,
		CatalogData: catalogData,
		Target:      model.Target{Host: "example.com", Scheme: "https"},
		Request:     model.ExecutionRequest{TechniqueID: "BB05.001"},
		Providers:   []model.LLMModel{model.Kimi},
		Logger:      logger,
		RandIntn:    func(int) int { return 0 },
		NewClient:   func(model.LLMModel) (llm.Client, error) { return client, nil },
	})
	if err != nil {
		t.Fatalf("未实现执行器不应视为致命错误: %v", err)
	}
	// 只保留初始执行结果，未实现的建议被跳过。
	if len(results) != 1 {
		t.Fatalf("期望 1 条结果，实际 %d", len(results))
	}
	if !strings.Contains(logBuffer.String(), "尚未实现") {
		t.Errorf("日志应记录未实现提示，实际: %s", logBuffer.String())
	}
}

// 验证 LLM 建议不存在于目录的技术时返回错误。
func TestRunInvalidSuggestion(t *testing.T) {
	technique.Register("directory-enumeration", stubTechnique{})
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}
	executionEngine := engine.New(catalogData)
	client := &scriptedClient{outputs: []string{
		`{"next_technique_id":"BB99.999","rationale":"不存在"}`,
	}}

	_, err = agent.Run(context.Background(), agent.RunParams{
		Engine:      executionEngine,
		CatalogData: catalogData,
		Target:      model.Target{Host: "example.com", Scheme: "https"},
		Request:     model.ExecutionRequest{TechniqueID: "BB05.001"},
		Providers:   []model.LLMModel{model.Kimi},
		Logger:      utilities.Default,
		RandIntn:    func(int) int { return 0 },
		NewClient:   func(model.LLMModel) (llm.Client, error) { return client, nil },
	})
	if err == nil {
		t.Fatal("期望无效建议报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "不在目录中") {
		t.Errorf("错误信息应说明技术不在目录，实际: %v", err)
	}
}
