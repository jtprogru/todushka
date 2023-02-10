package project

import (
	"fmt"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) GetProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		projectID, err := strconv.Atoi(chi.URLParam(r, "projectID"))
		msg := make(map[string]any)

		if err != nil {
			msg["result"] = "url parameter projectID is not number"
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(pkg.AnyToJSON(msg))
			return
		}
		project, err := h.srv.GetProjectByID(r.Context(), projectID)
		if err != nil {
			msg["result"] = fmt.Sprintf("project with id=%v not found", projectID)
			msg["status"] = http.StatusNotFound
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(pkg.AnyToJSON(msg))
			return
		}
		msg["result"] = project
		msg["status"] = http.StatusOK
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pkg.AnyToJSON(msg))
	}
}

func (h *Handler) GetAllProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := make(map[string]any)

		projects, err := h.srv.GetAllProjects(r.Context())
		if err != nil {
			msg["result"] = fmt.Sprintf("project internal err: %v", err)
			msg["status"] = http.StatusInternalServerError
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(pkg.AnyToJSON(msg))
		}
		msg["result"] = projects
		msg["status"] = http.StatusOK
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pkg.AnyToJSON(msg))
	}
}
