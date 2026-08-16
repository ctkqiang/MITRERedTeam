package utilities_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"mitre_red_team/internal/utilities"
)

// newTestLogger 构造写入内存缓冲区的日志器，便于断言输出内容。
func newTestLogger() (*utilities.Logger, *bytes.Buffer) {
	var buffer bytes.Buffer
	return utilities.New(&buffer), &buffer
}

// 验证输出为单行，且包含时间戳、操作名、内存统计与描述字段。
func TestLogSingleLineWithAllFields(t *testing.T) {
	logger, buffer := newTestLogger()
	logger.Info("DirectoryEnumeration",
		[]utilities.MemoryStat{{Name: "x", Bytes: 1024}, {Name: "y", Bytes: 4096}},
		"完成目录扫描")

	line := buffer.String()
	if strings.Count(line, "\n") != 1 {
		t.Fatalf("期望单行输出，实际换行数=%d，输出=%q", strings.Count(line, "\n"), line)
	}
	for _, field := range []string{"op=DirectoryEnumeration", "mem[x]=1.0KiB", "mem[y]=4.0KiB", "desc=完成目录扫描"} {
		if !strings.Contains(line, field) {
			t.Errorf("输出缺少字段 %q，实际=%s", field, line)
		}
	}
}

// 通过公开 API 的输出间接验证内存字节数的单位换算。
func TestLogFormatsMemoryBytes(t *testing.T) {
	testCases := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0B"},
		{1023, "1023B"},
		{1024, "1.0KiB"},
		{1536, "1.5KiB"},
		{1048576, "1.0MiB"},
		{1073741824, "1.0GiB"},
	}
	for _, testCase := range testCases {
		logger, buffer := newTestLogger()
		logger.Info("FormatBytesProbe",
			[]utilities.MemoryStat{{Name: "payload", Bytes: testCase.bytes}},
			"字节格式验证")

		expectedField := "mem[payload]=" + testCase.expected
		if !strings.Contains(buffer.String(), expectedField) {
			t.Errorf("字节 %d 期望字段 %q，实际=%s", testCase.bytes, expectedField, buffer.String())
		}
	}
}

// 验证描述中的换行与回车被替换为空格，保持单行。
func TestLogSanitizesDescription(t *testing.T) {
	logger, buffer := newTestLogger()
	logger.Warn("PortDiscovery", nil, "第一行\n第二行\r第三行")

	line := buffer.String()
	body := strings.TrimSuffix(line, "\n")
	if strings.ContainsAny(body, "\n\r") {
		t.Errorf("描述中的换行未被替换，输出=%q", line)
	}
	if !strings.Contains(line, "desc=第一行 第二行 第三行") {
		t.Errorf("描述未按预期合并，输出=%q", line)
	}
}

// 验证并发写入时每行保持完整，不出现字段交错。
func TestLogConcurrentWrites(t *testing.T) {
	logger, buffer := newTestLogger()
	var waitGroup sync.WaitGroup
	const concurrentWriters = 20
	for index := 0; index < concurrentWriters; index++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			logger.Info("ConcurrentOperation",
				[]utilities.MemoryStat{{Name: "payload", Bytes: uint64(index)}},
				"并发写入验证")
		}(index)
	}
	waitGroup.Wait()

	lines := strings.Split(strings.TrimSuffix(buffer.String(), "\n"), "\n")
	if len(lines) != concurrentWriters {
		t.Fatalf("期望 %d 行输出，实际 %d 行", concurrentWriters, len(lines))
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, "20") {
			t.Errorf("行缺少时间戳前缀，行=%q", line)
		}
		if strings.Count(line, "op=ConcurrentOperation") != 1 ||
			strings.Count(line, "desc=并发写入验证") != 1 {
			t.Errorf("并发写入导致字段交错或缺失，行=%q", line)
		}
	}
}
