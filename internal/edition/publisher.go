package edition

import "task257-scorecollate/internal/model"

// ConfirmedVariantIDs 提取已确认异文 ID 列表（用于冻结进校勘版本）。
func ConfirmedVariantIDs(variants []*model.Variant) []string {
	ids := make([]string, 0, len(variants))
	for _, v := range variants {
		if v.State == model.VarConfirmed {
			ids = append(ids, v.ID)
		}
	}
	return ids
}
