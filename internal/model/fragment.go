package model

import "time"

// FragmentState 是乐谱片段的解析/采用状态。
type FragmentState string

const (
	FragPendingParse FragmentState = "pending_parse" // 待解析
	FragAligned      FragmentState = "aligned"       // 已对齐
	FragUnreadable   FragmentState = "unreadable"    // 不可辨识
	FragExcluded     FragmentState = "excluded"      // 排除
)

var fragmentTransitions = map[FragmentState][]FragmentState{
	FragPendingParse: {FragAligned, FragUnreadable, FragExcluded},
	FragAligned:      {FragExcluded, FragUnreadable},
	FragUnreadable:   {FragPendingParse, FragExcluded},
	FragExcluded:     {FragAligned},
}

// ParseFragmentState 校验并转换文本状态。
func ParseFragmentState(s string) (FragmentState, bool) {
	switch FragmentState(s) {
	case FragPendingParse, FragAligned, FragUnreadable, FragExcluded:
		return FragmentState(s), true
	}
	return "", false
}

// CanTransitionTo 判断片段状态机迁移是否合法。
func (s FragmentState) CanTransitionTo(t FragmentState) bool {
	for _, n := range fragmentTransitions[s] {
		if n == t {
			return true
		}
	}
	return false
}

// Fragment 是某份传抄来源的手稿片段。
type Fragment struct {
	ID          string
	ProjectID   string
	SourceID    string
	Label       string
	State       FragmentState
	Fingerprint string
	RawNotation string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
