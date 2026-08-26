package httpapi

import (
	"net/http"
)

func (a *API) handleAlignProject(w http.ResponseWriter, r *http.Request, p map[string]string) {
	n, err := a.svc.AlignProject(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"variants_created": n})
}

func (a *API) handleListVariants(w http.ResponseWriter, r *http.Request, p map[string]string) {
	variants, err := a.svc.ListVariants(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, variants)
}

func (a *API) handleGetVariant(w http.ResponseWriter, r *http.Request, p map[string]string) {
	v, err := a.svc.GetVariant(p["id"])
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
