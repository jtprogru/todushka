package project

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/domain/entity"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) UpdateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := make(map[string]any)
		var data entity.ProjectUpdate
		todoID, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		if err != nil {
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(pkg.AnyToJSON(msg))
			return
		}
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(pkg.AnyToJSON(msg))
			return
		}
		todo, err := h.srv.UpdateProject(r.Context(), todoID, data)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo can't create with err: %v", err.Error())
			msg["status"] = http.StatusInternalServerError
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(pkg.AnyToJSON(msg))
			return
		}
		msg["result"] = todo
		msg["status"] = http.StatusAccepted
		_, _ = w.Write(pkg.AnyToJSON(msg))
	}
}
