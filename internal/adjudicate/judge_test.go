package adjudicate

import (
	"testing"

	"task257-scorecollate/internal/model"
)

func TestValidateAdjudication(t *testing.T) {
	cases := []struct {
		name     string
		current  model.VariantState
		decision model.VariantState
		wantOK   bool
		wantEnd  model.VariantState
	}{
		// 候选态只能先裁决为讹写/有效变体/证据不足之一。
		{"candidate->error", model.VarCandidate, model.VarError, true, model.VarError},
		{"candidate->variant", model.VarCandidate, model.VarValidVariant, true, model.VarValidVariant},
		{"candidate->insufficient", model.VarCandidate, model.VarInsufficient, true, model.VarInsufficient},

		// 已判定为讹写/有效变体/证据不足的条目仍可确认。
		{"error->confirmed", model.VarError, model.VarConfirmed, true, model.VarConfirmed},
		{"variant->confirmed", model.VarValidVariant, model.VarConfirmed, true, model.VarConfirmed},
		{"insufficient->confirmed", model.VarInsufficient, model.VarConfirmed, true, model.VarConfirmed},

		// 未判定类型的候选不得直接确认，以免未裁决差异进入可发布集合。
		{"candidate->confirmed rejected", model.VarCandidate, model.VarConfirmed, false, model.VarCandidate},

		// 非法跳转一律拒绝。
		{"candidate->candidate rejected", model.VarCandidate, model.VarCandidate, false, model.VarCandidate},
		{"error->variant rejected", model.VarError, model.VarValidVariant, false, model.VarError},
		{"confirmed->error rejected", model.VarConfirmed, model.VarError, false, model.VarConfirmed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			end, ok := ValidateAdjudication(c.current, c.decision)
			if ok != c.wantOK {
				t.Fatalf("ok = %v, want %v", ok, c.wantOK)
			}
			if ok && end != c.wantEnd {
				t.Fatalf("end = %s, want %s", end, c.wantEnd)
			}
			if !ok && end != c.current {
				t.Fatalf("on rejection end = %s, want unchanged %s", end, c.current)
			}
		})
	}
}
