package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mitre_red_team/internal/model"
	"mitre_red_team/internal/technique/enumeration"
)

// 验证字典缺失时返回失败，而不是伪装成成功。
func TestDirectoryEnumerationMissingWordlist(t *testing.T) {
	executor := enumeration.NewDirectoryEnumeration(fakeToolPath(t, behaviorEcho), filepath.Join(t.TempDir(), "missing.txt"))
	result, err := executor.Execute(context.Background(), model.Target{Host: "example.com", Scheme: "https"})
	if err == nil {
		t.Fatal("期望字典缺失返回错误，实际无错误")
	}
	if result.Status != model.StatusFailed {
		t.Errorf("期望失败状态，实际 %s", result.Status)
	}
}

// 验证外部工具非零退出时返回错误，stderr 空时给出退出码提示。
func TestDirectoryEnumerationToolFailure(t *testing.T) {
	wordlist := filepath.Join(t.TempDir(), "words.txt")
	if err := os.WriteFile(wordlist, []byte("admin\n"), 0600); err != nil {
		t.Fatalf("写入字典失败: %v", err)
	}
	executor := enumeration.NewDirectoryEnumeration(fakeToolPath(t, behaviorFail), wordlist)
	_, err := executor.Execute(context.Background(), model.Target{Host: "example.com", Scheme: "https"})
	if err == nil {
		t.Fatal("期望工具非零退出返回错误，实际无错误")
	}
}

// 验证执行成功且无命中时返回成功状态与明确提示。
func TestDirectoryEnumerationNoHits(t *testing.T) {
	wordlist := filepath.Join(t.TempDir(), "words.txt")
	if err := os.WriteFile(wordlist, []byte("admin\n"), 0600); err != nil {
		t.Fatalf("写入字典失败: %v", err)
	}
	executor := enumeration.NewDirectoryEnumeration(fakeToolPath(t, behaviorTrue), wordlist)
	result, err := executor.Execute(context.Background(), model.Target{Host: "example.com", Scheme: "https"})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if result.Status != model.StatusSucceeded {
		t.Errorf("期望成功状态，实际 %s", result.Status)
	}
	if !strings.Contains(result.Summary, "未发现命中") {
		t.Errorf("期望未发现命中提示，实际 %s", result.Summary)
	}
}
