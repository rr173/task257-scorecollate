package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"task257-scorecollate/internal/model"
)

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// httpStatus 把领域错误映射为 HTTP 状态码。
func httpStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, model.ErrInvalidInput),
		errors.Is(err, model.ErrInvalidState),
		errors.Is(err, model.ErrSelfCycle),
		errors.Is(err, model.ErrUnreadable):
		return http.StatusBadRequest
	case errors.Is(err, model.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, model.ErrSealed), errors.Is(err, model.ErrFrozen):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func fail(w http.ResponseWriter, err error) {
	writeError(w, httpStatus(err), err.Error())
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

// handleHealth 健康检查。
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "task257-scorecollate"})
}

// handleSelfCheck 项目自检快照。
func (a *API) handleSelfCheck(w http.ResponseWriter, r *http.Request, p map[string]string) {
	snapshot, err := a.svc.SelfCheck(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}
