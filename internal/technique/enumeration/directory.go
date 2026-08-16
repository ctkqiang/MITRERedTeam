package enumeration

import (
	"context"
	"strconv"
	"time"

	"mitre_red_team/internal/model"
	"mitre_red_team/tools"
)

// directoryEnumerationResultLimit 限制结果摘要长度，防止工具输出无限增长。
const directoryEnumerationResultLimit = 2000

// DirectoryEnumeration 使用 ffuf 枚举目标 Web 目录。
type DirectoryEnumeration struct {
	runner   *tools.Runner
	wordlist string
}

// NewDirectoryEnumeration 创建目录枚举技术。
// executablePath 为 ffuf 可执行文件路径，wordlist 为字典文件路径。
func NewDirectoryEnumeration(executablePath string, wordlist string) *DirectoryEnumeration {
	return &DirectoryEnumeration{
		runner:   tools.NewRunner(executablePath, 30*time.Second),
		wordlist: wordlist,
	}
}

// Execute 对 target 执行目录枚举。
// 命中状态码限定为 200/301/302/403，结果摘要截断以防无界增长。
func (d *DirectoryEnumeration) Execute(ctx context.Context, target model.Target) (model.ExecutionResult, error) {
	url := target.Scheme + "://" + target.Host
	if target.Port != 0 {
		url += ":" + strconv.Itoa(target.Port)
	}

	result, err := d.runner.Run(ctx, []string{
		"-u", url + "/FUZZ",
		"-w", d.wordlist,
		"-mc", "200,301,302,403",
	})

	status := model.StatusSucceeded
	summary := "目录枚举完成"
	if err != nil {
		status = model.StatusFailed
		summary = "目录枚举失败"
	} else if len(result.Stdout) > directoryEnumerationResultLimit {
		summary = "目录枚举完成（输出已截断）"
	}

	return model.ExecutionResult{
		TechniqueID: "BB05.001",
		Status:      status,
		Summary:     summary,
	}, err
}
