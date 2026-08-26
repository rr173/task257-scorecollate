package httpapi

import (
	"net/http"

	"task257-scorecollate/internal/model"
)

type adjudicateReq struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
	Editor   string `json:"editor"`
}

func (a *API) handleAdjudicateVariant(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req adjudicateReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	decision, ok := model.ParseVariantState(req.Decision)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid variant decision")
		return
	}
	v, err := a.svc.AdjudicateVariant(p["id"], decision, req.Reason, req.Editor)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
