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
		todoId, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`url parameter todoID is not number`))
			return
		}
		todo, err := h.srv.GetById(r.Context(), todoId)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(pkg.AnyToJson(err.Error()))
		}
		w.Write(pkg.AnyToJson(todo))
	}
}

func (h *Handler) GetAllTodos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		todos, err := h.srv.GetAllTodos(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(pkg.AnyToJson(err.Error()))
		}
		w.Write(pkg.AnyToJson(todos))
	}
}
