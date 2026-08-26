package edition

import "task257-scorecollate/internal/model"

// ConfirmedVariantIDs 提取已确认异文 ID 列表（用于冻结进校勘版本）。
// 尚未裁决（candidate）或仅初判未确认的异文不在此列，故不会进入冻结快照。
func ConfirmedVariantIDs(variants []*model.Variant) []string {
	ids := make([]string, 0, len(variants))
	for _, v := range variants {
		if v.State == model.VarConfirmed {
			ids = append(ids, v.ID)
		}
	}
	return ids
}
