package todo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/domain/entity"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) UpdateTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := make(map[string]any)
		var data entity.TodoUpdate
		todoId, err := strconv.Atoi(chi.URLParam(r, "todoID"))
		if err != nil {
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			w.Write(pkg.AnyToJson(msg))
			return
		}
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			w.Write(pkg.AnyToJson(msg))
			return
		}
		todo, err := h.srv.UpdateTodo(r.Context(), todoId, data)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo can't create with err: %v", err.Error())
			msg["status"] = http.StatusInternalServerError
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(pkg.AnyToJson(msg))
			return
		}
		msg["result"] = todo
		msg["status"] = http.StatusAccepted
		w.Write(pkg.AnyToJson(msg))
	}
}
