package model

import "time"

// ProjectState 是校勘项目的生命周期状态。
type ProjectState string

const (
	ProjectArranging         ProjectState = "arranging"          // 整理中
	ProjectAwaitingAlignment ProjectState = "awaiting_alignment" // 待对齐
	ProjectAwaitingAdjud     ProjectState = "awaiting_adjudication" // 待裁决
	ProjectPublished         ProjectState = "published"          // 已发布
	ProjectSealed            ProjectState = "sealed"             // 封存
)

// projectTransitions 描述项目状态机的合法迁移。
var projectTransitions = map[ProjectState][]ProjectState{
	ProjectArranging:         {ProjectAwaitingAlignment, ProjectSealed},
	ProjectAwaitingAlignment: {ProjectAwaitingAdjud, ProjectArranging, ProjectSealed},
	ProjectAwaitingAdjud:     {ProjectPublished, ProjectAwaitingAlignment, ProjectSealed},
	ProjectPublished:         {ProjectSealed},
	ProjectSealed:            {},
}

// ParseProjectState 校验并转换文本状态。
func ParseProjectState(s string) (ProjectState, bool) {
	switch ProjectState(s) {
	case ProjectArranging, ProjectAwaitingAlignment, ProjectAwaitingAdjud, ProjectPublished, ProjectSealed:
		return ProjectState(s), true
	}
	return "", false
}

// IsValid 判断状态是否合法。
func (s ProjectState) IsValid() bool {
	_, ok := projectTransitions[s]
	return ok
}

// CanTransitionTo 判断能否迁移到目标状态。
func (s ProjectState) CanTransitionTo(t ProjectState) bool {
	for _, n := range projectTransitions[s] {
		if n == t {
			return true
		}
	}
	return false
}

// IsTerminal 判断是否为终态（封存后不可再迁移）。
func (s ProjectState) IsTerminal() bool { return len(projectTransitions[s]) == 0 }

// Project 是校勘工程的聚合根。
type Project struct {
	ID          string
	Title       string
	Description string
	State       ProjectState
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
