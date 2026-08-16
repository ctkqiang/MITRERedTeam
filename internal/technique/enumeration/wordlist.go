package enumeration

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// wordlistMaxLineLength 单条字典条目最大长度，防止异常行撑爆内存。
const wordlistMaxLineLength = 64 * 1024

// ValidateWordlist 校验字典文件存在、可读且格式合法。
// 格式要求：UTF-8 编码，每行一个条目；空行与 # 开头的注释行被忽略。
// 校验不通过时返回带路径与原因的错误，便于用户定位问题。
func ValidateWordlist(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("字典路径为空")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("字典文件不存在: %s", path)
		}
		return fmt.Errorf("无法访问字典文件 %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("字典路径 %s 是目录而非文件", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法读取字典文件 %s（请检查文件权限）: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), wordlistMaxLineLength)
	entryCount := 0
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !utf8.ValidString(line) {
			return fmt.Errorf("字典文件 %s 第 %d 行不是合法的 UTF-8 文本", path, lineNumber)
		}
		if strings.ContainsAny(line, "\t\r\n") {
			return fmt.Errorf("字典文件 %s 第 %d 行包含制表符或换行等非法字符", path, lineNumber)
		}
		entryCount++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取字典文件 %s 失败: %w", path, err)
	}
	if entryCount == 0 {
		return fmt.Errorf("字典文件 %s 为空或没有有效条目", path)
	}
	return nil
}

// ResolveWordlistPath 决定最终使用的字典路径。
// manualPath 非空时优先使用并校验；为空则交互询问，回车或 EOF 视为不指定；
// 仍未得到路径时回退 defaultPath。选定的路径必须通过 ValidateWordlist。
func ResolveWordlistPath(manualPath string, defaultPath string, reader io.Reader, writer io.Writer) (string, error) {
	chosen := strings.TrimSpace(manualPath)
	if chosen == "" {
		chosen = promptForWordlist(reader, writer)
	}
	if chosen == "" {
		chosen = defaultPath
	}
	if err := ValidateWordlist(chosen); err != nil {
		return "", err
	}
	return chosen, nil
}

// promptForWordlist 从 reader 读取一行作为用户自定义字典路径，空输入返回空串。
func promptForWordlist(reader io.Reader, writer io.Writer) string {
	fmt.Fprintln(writer, "未通过命令行指定自定义字典。")
	fmt.Fprint(writer, "请输入自定义字典文件路径（直接回车使用默认字典）: ")
	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return ""
	}
	return strings.TrimSpace(line)
}

// ConfirmDefaultFallback 询问用户是否在自定义字典不可用时改用默认字典。
// 回车、y、yes 视为同意，其余视为拒绝。
func ConfirmDefaultFallback(reader io.Reader, writer io.Writer) bool {
	fmt.Fprint(writer, "自定义字典不可用，是否改用默认字典继续？[Y/n] ")
	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "" || answer == "y" || answer == "yes"
}
