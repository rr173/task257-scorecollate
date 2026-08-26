package model

import "time"

// VariantState 是乐谱异文（音符差异）的裁决状态机。
type VariantState string

const (
	VarCandidate    VariantState = "candidate"    // 候选
	VarError        VariantState = "error"        // 讹写
	VarValidVariant VariantState = "variant"      // 有效变体
	VarInsufficient VariantState = "insufficient" // 证据不足
	VarConfirmed    VariantState = "confirmed"    // 确认
)

var variantTransitions = map[VariantState][]VariantState{
	VarCandidate:    {VarError, VarValidVariant, VarInsufficient},
	VarError:        {VarConfirmed},
	VarValidVariant: {VarConfirmed},
	VarInsufficient: {VarConfirmed},
	VarConfirmed:    {},
}

// ParseVariantState 校验并转换文本状态。
func ParseVariantState(s string) (VariantState, bool) {
	switch VariantState(s) {
	case VarCandidate, VarError, VarValidVariant, VarInsufficient, VarConfirmed:
		return VariantState(s), true
	}
	return "", false
}

// CanTransitionTo 判断异文状态机迁移是否合法。
func (s VariantState) CanTransitionTo(t VariantState) bool {
	for _, n := range variantTransitions[s] {
		if n == t {
			return true
		}
	}
	return false
}

// Variant 表示在某小节/声部上，两份片段之间的音符差异。
type Variant struct {
	ID            string
	ProjectID     string
	MeasureNumber int
	Voice         int
	FragmentAID   string
	FragmentBID   string
	ContentA      string
	ContentB      string
	DetectedKind  VariantState // 自动识别的初判类型
	State         VariantState
	SupportCount  int // 独立来源支持数：除 A/B 所在祖本谱系外，来自不同祖本谱系且持相同读法的来源数
	Version       int64
	CreatedAt     time.Time
}

// Adjudication 是研究者对某异文的裁决记录。
type Adjudication struct {
	ID        string
	VariantID string
	Decision  VariantState
	Reason    string
	Editor    string
	CreatedAt time.Time
}
