package main

// Status 表示订单状态
type Status int

const (
	StatusPending   Status = iota // 待处理
	StatusActive                  // 进行中
	StatusCompleted               // 已完成
	StatusCancelled               // 已取消
)

// Priority 表示优先级
type Priority int

const (
	PriorityLow    Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)
