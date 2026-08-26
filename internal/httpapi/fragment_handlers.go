package httpapi

import (
	"net/http"

	"task257-scorecollate/internal/model"
)

type createFragmentReq struct {
	SourceID string `json:"source_id"`
	Label    string `json:"label"`
	Raw      string `json:"raw"`
}

func (a *API) handleCreateFragment(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req createFragmentReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	f, err := a.svc.CreateFragment(p["id"], req.SourceID, req.Label, req.Raw)
	_ = req
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (a *API) handleListFragments(w http.ResponseWriter, r *http.Request, p map[string]string) {
	fragments, err := a.svc.ListFragments(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, fragments)
}

func (a *API) handleGetFragment(w http.ResponseWriter, r *http.Request, p map[string]string) {
	f, err := a.svc.GetFragment(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

type fragmentStateReq struct {
	State string `json:"state"`
}

func (a *API) handleSetFragmentState(w http.ResponseWriter, r *http.Request, p map[string]string) {
	var req fragmentStateReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	to, ok := model.ParseFragmentState(req.State)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid fragment state")
		return
	}
	f, err := a.svc.SetFragmentState(p["id"], to)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (a *API) handleParseFragment(w http.ResponseWriter, r *http.Request, p map[string]string) {
	f, measures, err := a.svc.ParseFragment(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"fragment": f,
		"measures": measures,
	})
}

func (a *API) handleListMeasures(w http.ResponseWriter, r *http.Request, p map[string]string) {
	measures, err := a.svc.ListMeasures(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, measures)
}
