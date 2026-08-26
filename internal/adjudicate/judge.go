package adjudicate

import "task257-scorecollate/internal/model"

// ValidateAdjudication 校验裁决迁移是否合法，并返回最终状态与是否通过。
// 候选态只能先迁移到 讹写/有效变体/证据不足 之一，再各自迁移到 确认；
// 未判定类型的候选不得直接确认，以免未裁决的差异进入可发布集合。
func ValidateAdjudication(current model.VariantState, decision model.VariantState) (model.VariantState, bool) {
	if !current.CanTransitionTo(decision) {
		return current, false
	}
	return decision, true
}
