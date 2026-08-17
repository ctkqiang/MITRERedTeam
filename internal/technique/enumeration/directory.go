package enumeration

import (
	"context"
	"fmt"
	"mitre_red_team/internal/model"
	"mitre_red_team/tools/ffuf"
	"os"
	"strconv"
	"strings"
	"time"
)

// directoryEnumerationResultLimit 限制成功摘要长度，防止工具输出无限增长。
const directoryEnumerationResultLimit = 2000

// directoryEnumerationErrorLimit 限制失败摘要长度，只保留关键错误信息。
const directoryEnumerationErrorLimit = 300

// directoryEnumerationPreviewLimit 成功摘要中展示的命中条目数量。
const directoryEnumerationPreviewLimit = 5

// DirectoryEnumeration 使用 ffuf 枚举目标 Web 目录。
type DirectoryEnumeration struct {
	fuzzer   *ffuf.Fuzzer
	wordlist string
}

// NewDirectoryEnumeration 创建目录枚举技术。
// executablePath 为 ffuf 可执行文件路径，wordlist 为字典文件路径。
func NewDirectoryEnumeration(executablePath string, wordlist string) *DirectoryEnumeration {
	return &DirectoryEnumeration{
		fuzzer:   ffuf.New(executablePath, 60*time.Second),
		wordlist: wordlist,
	}
}

// Execute 对 target 执行目录枚举。
// 命中状态码限定为 200/301/302/403；silent 模式下 stdout 逐行输出命中条目，据此统计发现。
// 字典缺失或 ffuf 非零退出时返回错误，避免把失败伪装成成功。
func (d *DirectoryEnumeration) Execute(ctx context.Context, target model.Target) (model.ExecutionResult, error) {
	if _, err := os.Stat(d.wordlist); err != nil {
		return model.ExecutionResult{
			TechniqueID: "BB05.001",
			Status:      model.StatusFailed,
			Summary:     "目录枚举失败：字典文件不存在",
		}, fmt.Errorf("目录枚举字典 %s 不存在: %w", d.wordlist, err)
	}

	url := target.Scheme + "://" + target.Host
	if target.Port != 0 {
		url += ":" + strconv.Itoa(target.Port)
	}

	// 参数构造收拢在 ffuf 适配器内，技术层只声明目标、字典与默认选项。
	result, err := d.fuzzer.Fuzz(ctx, url+"/FUZZ", d.wordlist, ffuf.DefaultFuzzOptions())
	if err != nil {
		return model.ExecutionResult{
			TechniqueID: "BB05.001",
			Status:      model.StatusFailed,
			Summary:     "目录枚举失败",
		}, err
	}

	// ffuf 非零退出说明执行异常（如目标不可达），stderr 携带原因，不得吞掉。
	if result.ExitCode != 0 {
		failureReason := strings.TrimSpace(result.Stderr)
		if failureReason == "" {
			failureReason = fmt.Sprintf("ffuf 退出码 %d", result.ExitCode)
		}
		if len(failureReason) > directoryEnumerationErrorLimit {
			failureReason = failureReason[:directoryEnumerationErrorLimit] + "..."
		}
		return model.ExecutionResult{
			TechniqueID: "BB05.001",
			Status:      model.StatusFailed,
			Summary:     "目录枚举失败：" + failureReason,
		}, fmt.Errorf("ffuf 执行失败: %s", failureReason)
	}

	// silent 模式下 stdout 每行一个命中条目，据此统计发现。
	hits := nonEmptyLines(result.Stdout)
	summary := "目录枚举完成，未发现命中（匹配状态码 200/301/302/403）"
	if len(hits) > 0 {
		preview := strings.Join(hits[:min(directoryEnumerationPreviewLimit, len(hits))], ", ")
		summary = fmt.Sprintf("目录枚举完成，发现 %d 个命中: %s", len(hits), preview)
	}
	if len(summary) > directoryEnumerationResultLimit {
		summary = summary[:directoryEnumerationResultLimit] + "..."
	}

	return model.ExecutionResult{
		TechniqueID: "BB05.001",
		Status:      model.StatusSucceeded,
		Summary:     summary,
	}, nil
}

// nonEmptyLines 按行拆分输出并去除空行与首尾空白。
func nonEmptyLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}
