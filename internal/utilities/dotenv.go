package utilities

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadDotenv 从 path 读取 KEY=VALUE 格式的环境变量文件并写入进程环境。
// 已存在的环境变量不会被覆盖；空行与 # 开头的注释行被忽略；值两侧的引号会被剥离。
// 文件不存在时视为无配置，直接返回 nil。
func LoadDotenv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("打开环境文件 %s: %w", path, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取环境文件 %s: %w", path, err)
	}
	return nil
}
