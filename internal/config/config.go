package config

import (
	"os/exec"
	"sort"
)

// Config 描述应用运行配置。
// Tools 保存外部工具名到可执行文件路径的映射；Wordlists 保存字典名到文件路径的映射；
// Notifications 描述通知偏好。
type Config struct {
	Tools         map[string]string  `json:"tools"`
	Wordlists     map[string]string  `json:"wordlists"`
	Notifications NotificationConfig `json:"notifications"`
}

// CheckToolsAvailable 检查配置中所有工具路径是否可执行。
// 对含分隔符的路径直接校验文件存在性与执行权限；对命令名经 PATH 查找。
// 返回缺失的工具名列表（按名称排序），全部可用时返回空切片。
func (c *Config) CheckToolsAvailable() []string {
	var missing []string
	for name, path := range c.Tools {
		if _, err := exec.LookPath(path); err != nil {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

// NotificationConfig 描述通知偏好。
// Platform 取值：telegram / wechat / none。平台凭据不存放在此配置中，
// 一律从环境变量读取，避免敏感信息进入版本仓库。
// Local 控制本地系统通知开关，配置缺省时视为开启。
type NotificationConfig struct {
	Platform string `json:"platform"`
	Local    *bool  `json:"local"`
}

// LocalEnabled 返回本地通知是否开启。配置未显式声明 local 字段时默认开启。
func (n NotificationConfig) LocalEnabled() bool {
	if n.Local == nil {
		return true
	}
	return *n.Local
}
