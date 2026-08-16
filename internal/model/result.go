package model

import "time"

// ExecutionStatus 定义单条技术的执行状态。
type ExecutionStatus string

const (
	StatusSucceeded ExecutionStatus = "succeeded"
	StatusFailed    ExecutionStatus = "failed"
	StatusSkipped   ExecutionStatus = "skipped"
)

// ExecutionResult 描述单条技术的执行结果。
type ExecutionResult struct {
	TechniqueID string
	Status      ExecutionStatus
	Duration    time.Duration
	Summary     string
}

// Finding 描述一次最终发现。
type Finding struct {
	TechniqueID string
	Severity    string
	Description string
	Evidence    string
}
