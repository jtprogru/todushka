package todo

import (
	"fmt"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/pkg/utils"
)

func (h *Handler) GetTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		todoID, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		msg := make(map[string]any)

		if err != nil {
			msg["result"] = "url parameter todoID is not number"
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(utils.AnyToJSON(msg))
			return
		}
		todo, err := h.srv.GetTodoByID(r.Context(), todoID)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo with id=%v not found", todoID)
			msg["status"] = http.StatusNotFound
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(utils.AnyToJSON(msg))
			return
		}
		msg["result"] = todo
		msg["status"] = http.StatusOK
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(utils.AnyToJSON(msg))
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
			_, _ = w.Write(utils.AnyToJSON(msg))
		}
		msg["result"] = todos
		msg["status"] = http.StatusOK
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(utils.AnyToJSON(msg))
	}
}
