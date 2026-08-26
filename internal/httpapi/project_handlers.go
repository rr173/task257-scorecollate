package httpapi

import (
	"net/http"

	"task257-scorecollate/internal/model"
)

type createProjectReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (a *API) handleCreateProject(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var req createProjectReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	p, err := a.svc.CreateProject(req.Title, req.Description)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (a *API) handleListProjects(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	projects, err := a.svc.ListProjects()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (a *API) handleGetProject(w http.ResponseWriter, r *http.Request, p map[string]string) {
	proj, err := a.svc.GetProject(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, proj)
}

type transitionReq struct {
	State string `json:"state"`
}

func (a *API) handleTransitionProject(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req transitionReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	to, ok := model.ParseProjectState(req.State)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid project state")
		return
	}
	proj, err := a.svc.TransitionProject(p["id"], to)
	if err != nil && to != model.ProjectSealed {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, proj)
}
