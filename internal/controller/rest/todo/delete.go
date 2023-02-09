package todo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) DeleteTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := make(map[string]any)
		todoId, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		if err != nil {
			msg["status"] = http.StatusBadRequest
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write(pkg.AnyToJson(msg))
			return
		}

		err = h.srv.DeleteTodo(r.Context(), todoId)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(pkg.AnyToJson(err.Error()))
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
