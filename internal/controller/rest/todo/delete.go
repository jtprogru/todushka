package todo

import (
	"fmt"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) DeleteTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := make(map[string]any)
		todoID, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		if err != nil {
			msg["status"] = http.StatusBadRequest
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(pkg.AnyToJSON(msg))
			return
		}

		err = h.srv.DeleteTodo(r.Context(), todoID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(pkg.AnyToJSON(msg))
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
