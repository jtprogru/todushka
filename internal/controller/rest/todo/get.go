package todo

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) GetTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		if err != nil {
			w.Write([]byte(`url parameter todoID is not number`))
			return
		}
		todo, _ := h.srv.GetById(r.Context(), 1)
		w.Write(pkg.AnyToJson(todo))
	}
}
