package httpapi

import (
	"net/http"
)

type createEditionReq struct {
	Title string `json:"title"`
}

func (a *API) handleCreateEdition(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req createEditionReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	e, err := a.svc.CreateEdition(p["id"], req.Title)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (a *API) handleListEditions(w http.ResponseWriter, r *http.Request, p map[string]string) {
	editions, err := a.svc.ListEditions(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, editions)
}

func (a *API) handleGetEdition(w http.ResponseWriter, r *http.Request, p map[string]string) {
	e, err := a.svc.GetEdition(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (a *API) handlePublishEdition(w http.ResponseWriter, r *http.Request, p map[string]string) {
	e, err := a.svc.PublishEdition(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

type supersedeReq struct {
	Title string `json:"title"`
}

func (a *API) handleSupersedeEdition(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req supersedeReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	e, err := a.svc.SupersedeEdition(p["id"], req.Title)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}
