package tests

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mitre_red_team/internal/technique/enumeration"
)

// writeTestWordlist 在临时目录写入指定内容的字典文件并返回路径。
func writeTestWordlist(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "words.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("写入字典失败: %v", err)
	}
	return path
}

// 验证合法字典通过校验，空行与注释行被忽略。
func TestValidateWordlistValid(t *testing.T) {
	path := writeTestWordlist(t, "# 注释行\n\nadmin\napi\nuploads\n\n")
	if err := enumeration.ValidateWordlist(path); err != nil {
		t.Errorf("期望合法字典通过校验，实际错误: %v", err)
	}
}

// 验证不存在的文件报错且错误信息清晰。
func TestValidateWordlistMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")
	err := enumeration.ValidateWordlist(path)
	if err == nil {
		t.Fatal("期望不存在的字典报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "不存在") {
		t.Errorf("错误信息应说明文件不存在，实际: %v", err)
	}
}

// 验证目录路径被拒绝。
func TestValidateWordlistIsDirectory(t *testing.T) {
	err := enumeration.ValidateWordlist(t.TempDir())
	if err == nil {
		t.Fatal("期望目录路径报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "目录") {
		t.Errorf("错误信息应说明是目录，实际: %v", err)
	}
}

// 验证无有效条目的文件报错。
func TestValidateWordlistEmpty(t *testing.T) {
	path := writeTestWordlist(t, "\n# 只有注释\n")
	err := enumeration.ValidateWordlist(path)
	if err == nil {
		t.Fatal("期望空字典报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "没有有效条目") {
		t.Errorf("错误信息应说明没有有效条目，实际: %v", err)
	}
}

// 验证非法 UTF-8 字节序列被拒绝。
func TestValidateWordlistInvalidUTF8(t *testing.T) {
	path := writeTestWordlist(t, "admin\n\xff\xfe\x00bad\n")
	err := enumeration.ValidateWordlist(path)
	if err == nil {
		t.Fatal("期望非法编码报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "UTF-8") {
		t.Errorf("错误信息应说明编码问题，实际: %v", err)
	}
}

// 验证包含制表符等非法字符的条目被拒绝。
func TestValidateWordlistTabCharacter(t *testing.T) {
	path := writeTestWordlist(t, "admin\napi\tv1\n")
	err := enumeration.ValidateWordlist(path)
	if err == nil {
		t.Fatal("期望制表符条目报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "非法字符") {
		t.Errorf("错误信息应说明非法字符，实际: %v", err)
	}
}

// 验证超长行被拒绝，防止异常输入撑爆内存。
func TestValidateWordlistOverlongLine(t *testing.T) {
	path := writeTestWordlist(t, "admin\n"+strings.Repeat("a", 70*1024)+"\n")
	if err := enumeration.ValidateWordlist(path); err == nil {
		t.Fatal("期望超长行报错，实际无错误")
	}
}

// 验证无读取权限的文件被拒绝（受限环境下跳过）。
func TestValidateWordlistUnreadable(t *testing.T) {
	path := writeTestWordlist(t, "admin\n")
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatalf("修改权限失败: %v", err)
	}
	defer os.Chmod(path, 0600)
	if _, err := os.Open(path); err != nil {
		// 当前用户确实无权读取时，才断言校验报错。
		if validationErr := enumeration.ValidateWordlist(path); validationErr == nil {
			t.Fatal("期望无权限字典报错，实际无错误")
		}
	}
}

// 验证手动指定字典优先于默认字典。
func TestResolveWordlistPathManual(t *testing.T) {
	custom := writeTestWordlist(t, "admin\n")
	chosen, err := enumeration.ResolveWordlistPath(custom, "/nonexistent/default.txt", strings.NewReader(""), io.Discard)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if chosen != custom {
		t.Errorf("期望使用手动字典 %s，实际 %s", custom, chosen)
	}
}

// 验证手动路径无效时返回错误，不静默回退。
func TestResolveWordlistPathManualInvalid(t *testing.T) {
	_, err := enumeration.ResolveWordlistPath(
		"/nonexistent/custom.txt",
		"/nonexistent/default.txt",
		strings.NewReader(""),
		io.Discard,
	)
	if err == nil {
		t.Fatal("期望无效手动字典报错，实际无错误")
	}
}

// 验证未手动指定且交互输入有效路径时使用该路径。
func TestResolveWordlistPathPromptInput(t *testing.T) {
	custom := writeTestWordlist(t, "admin\n")
	var promptOutput strings.Builder
	chosen, err := enumeration.ResolveWordlistPath(
		"",
		"/nonexistent/default.txt",
		strings.NewReader(custom+"\n"),
		&promptOutput,
	)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if chosen != custom {
		t.Errorf("期望使用交互输入字典 %s，实际 %s", custom, chosen)
	}
	if !strings.Contains(promptOutput.String(), "自定义字典") {
		t.Error("提示信息应说明自定义字典输入方式")
	}
}

// 验证交互输入为空（直接回车）时使用默认字典。
func TestResolveWordlistPathPromptEmptyUsesDefault(t *testing.T) {
	defaultPath := writeTestWordlist(t, "admin\n")
	chosen, err := enumeration.ResolveWordlistPath("", defaultPath, strings.NewReader("\n"), io.Discard)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if chosen != defaultPath {
		t.Errorf("期望回车后使用默认字典 %s，实际 %s", defaultPath, chosen)
	}
}

// 验证回退确认：回车与 y 视为同意，n 视为拒绝。
func TestConfirmDefaultFallback(t *testing.T) {
	agree := map[string]bool{
		"\n":    true,
		"y\n":   true,
		"yes\n": true,
		"Y\n":   true,
		"n\n":   false,
		"no\n":  false,
	}
	for input, want := range agree {
		if got := enumeration.ConfirmDefaultFallback(strings.NewReader(input), io.Discard); got != want {
			t.Errorf("输入 %q 期望 %v，实际 %v", input, want, got)
		}
	}
}

// 验证非交互环境（stdin 为管道）跳过交互询问，直接使用默认字典，且不阻塞。
func TestResolveWordlistPathNonInteractiveFallback(t *testing.T) {
	defaultPath := writeTestWordlist(t, "admin\n")
	pipeReader, pipeWriter, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("创建管道失败: %v", pipeErr)
	}
	defer pipeReader.Close()
	defer pipeWriter.Close()

	var output strings.Builder
	chosen, err := enumeration.ResolveWordlistPath("", defaultPath, pipeReader, &output)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if chosen != defaultPath {
		t.Errorf("期望非交互环境使用默认字典 %s，实际 %s", defaultPath, chosen)
	}
	if !strings.Contains(output.String(), "默认字典") {
		t.Errorf("提示信息应说明回退到默认字典，实际: %q", output.String())
	}
}

// 验证非交互环境无法确认回退时返回拒绝，避免无声阻塞。
func TestConfirmDefaultFallbackNonInteractive(t *testing.T) {
	pipeReader, pipeWriter, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("创建管道失败: %v", pipeErr)
	}
	defer pipeReader.Close()
	defer pipeWriter.Close()

	var output strings.Builder
	if got := enumeration.ConfirmDefaultFallback(pipeReader, &output); got {
		t.Fatal("期望非交互环境拒绝回退，实际同意")
	}
	if !strings.Contains(output.String(), "无法确认回退") {
		t.Errorf("提示信息应说明无法确认回退，实际: %q", output.String())
	}
}
