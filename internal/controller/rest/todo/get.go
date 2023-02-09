package todo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) GetTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		todoId, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		msg := make(map[string]any)

		if err != nil {
			msg["result"] = "url parameter todoID is not number"
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			w.Write(pkg.AnyToJson(msg))
			return
		}
		todo, err := h.srv.GetById(r.Context(), todoId)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo with id=%v not found", todoId)
			msg["status"] = http.StatusNotFound
			w.WriteHeader(http.StatusNotFound)
			w.Write(pkg.AnyToJson(msg))
			return
		}
		msg["result"] = todo
		msg["status"] = http.StatusOK
		w.WriteHeader(http.StatusOK)
		w.Write(pkg.AnyToJson(msg))
	}
}

func (h *Handler) GetAllTodos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := make(map[string]any)

		todos, err := h.srv.GetAllTodos(r.Context())
		if err != nil {
			msg["result"] = fmt.Sprintf("todo internal err: %v", err)
			msg["status"] = http.StatusInternalServerError
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(pkg.AnyToJson(msg))
		}
		msg["result"] = todos
		msg["status"] = http.StatusOK
		w.WriteHeader(http.StatusOK)
		w.Write(pkg.AnyToJson(msg))
	}
}
