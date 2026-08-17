package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"mitre_red_team/internal/model"
)

// Decision 描述 LLM 对 TTP 执行结果分析后给出的下一步建议。
// NextTechniqueID 为空表示建议停止，不再推进。
type Decision struct {
	NextTechniqueID string `json:"next_technique_id"`
	Rationale       string `json:"rationale"`
}

// BuildDecisionPrompt 构造要求 LLM 分析 TTP 执行结果的系统与用户提示词。
// 系统提示词限定 LLM 只能从目录中的技术里选择下一步，且必须输出指定 JSON 结构；
// 用户提示词携带目标、执行历史与当前轮数，供 LLM 结合上下文决策。
func BuildDecisionPrompt(
	target model.Target,
	techniques []model.Technique,
	history []model.ExecutionResult,
	round int,
	maxRounds int,
) (systemPrompt string, userPrompt string) {
	systemPrompt = "你是授权红队评估的战术决策助手。" +
		"你会收到某次安全技术（TTP）对指定目标执行后的结果。" +
		"请分析结果并只推荐下一步最合理的单条技术，技术必须从给定列表中选取。" +
		"如果结果已充分或没有合适的下一步，next_technique_id 返回空字符串。" +
		"只输出 JSON，不要输出任何其他文字，格式：{\"next_technique_id\":\"BBxx.xxx\",\"rationale\":\"简短理由\"}。" +
		"候选技术列表：\n"
	for _, technique := range techniques {
		line := fmt.Sprintf("- %s %s", technique.ID, technique.Name)
		if strings.TrimSpace(technique.Description) != "" {
			line += "：" + technique.Description
		}
		systemPrompt += line + "\n"
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "目标: %s://%s", target.Scheme, target.Host)
	if target.Port != 0 {
		fmt.Fprintf(&builder, ":%d", target.Port)
	}
	fmt.Fprintf(&builder, "\n\n已执行的技术与结果（当前第 %d 轮，最多 %d 轮）:\n", round, maxRounds)
	for _, result := range history {
		fmt.Fprintf(&builder, "- [%s] %s (%s): %s\n", result.TechniqueID, result.Status, result.Duration, result.Summary)
	}
	if len(history) == 0 {
		builder.WriteString("- 暂无执行结果\n")
	}
	return systemPrompt, builder.String()
}

// ParseDecision 从 LLM 输出中提取 Decision。
// 输出可能被 ```json 代码块包裹，这里截取最外层花括号内容后再解析，
// 容忍常见的格式噪声；解析失败返回带原文的错误，便于排查。
func ParseDecision(output string) (Decision, error) {
	trimmed := strings.TrimSpace(output)
	start := strings.IndexByte(trimmed, '{')
	end := strings.LastIndexByte(trimmed, '}')
	if start < 0 || end < start {
		return Decision{}, fmt.Errorf("输出中未找到 JSON 对象: %s", truncate(output, 200))
	}
	var decision Decision
	if err := json.Unmarshal([]byte(trimmed[start:end+1]), &decision); err != nil {
		return Decision{}, fmt.Errorf("解析建议 JSON 失败: %w，原文: %s", err, truncate(output, 200))
	}
	return decision, nil
}

// truncate 截断文本到指定长度，防止超长输出进入错误信息。
func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
