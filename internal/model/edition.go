package model

import "time"

// EditionState 是校勘版本的发布状态机。
type EditionState string

const (
	EditionDraft      EditionState = "draft"      // 草稿
	EditionShared     EditionState = "shared"     // 共享
	EditionFrozen     EditionState = "frozen"     // 冻结（不可变）
	EditionSuperseded EditionState = "superseded" // 替代
)

var editionTransitions = map[EditionState][]EditionState{
	EditionDraft:      {EditionShared, EditionFrozen, EditionSuperseded},
	EditionShared:     {EditionFrozen, EditionSuperseded},
	EditionFrozen:     {},
	EditionSuperseded: {},
}

// ParseEditionState 校验并转换文本状态。
func ParseEditionState(s string) (EditionState, bool) {
	switch EditionState(s) {
	case EditionDraft, EditionShared, EditionFrozen, EditionSuperseded:
		return EditionState(s), true
	}
	return "", false
}

// CanTransitionTo 判断版本状态机迁移是否合法。
func (s EditionState) CanTransitionTo(t EditionState) bool {
	for _, n := range editionTransitions[s] {
		if n == t {
			return true
		}
	}
	return false
}

// IsImmutable 判断版本是否已冻结/替代（不可修改）。
func (s EditionState) IsImmutable() bool { return s == EditionFrozen || s == EditionSuperseded }

// Edition 是不可变校勘版本的载体。
type Edition struct {
	ID        string
	ProjectID string
	Title     string
	State     EditionState
	Version   int64
	CreatedAt time.Time
}

// EditionVariantLink 记录某版本收录了哪些异文及是否纳入定本。
type EditionVariantLink struct {
	EditionID string
	VariantID string
	Included  bool
}

// AuditEntry 是项目级审计日志。
type AuditEntry struct {
	ID        string
	ProjectID string
	Action    string
	Detail    string
	CreatedAt time.Time
}
