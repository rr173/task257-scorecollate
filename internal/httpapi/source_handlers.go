package httpapi

import (
	"net/http"
)

type createSourceReq struct {
	Siglum      string `json:"siglum"`
	Title       string `json:"title"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description"`
}

func (a *API) handleCreateSource(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req createSourceReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	src, err := a.svc.CreateSource(p["id"], req.Siglum, req.Title, req.ParentID, req.Description)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, src)
}

func (a *API) handleListSources(w http.ResponseWriter, r *http.Request, p map[string]string) {
	sources, err := a.svc.ListSources(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sources)
}

type reparentReq struct {
	ParentID string `json:"parent_id"`
}

func (a *API) handleReparentSource(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req reparentReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	src, err := a.svc.ReparentSource(p["id"], req.ParentID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, src)
}
