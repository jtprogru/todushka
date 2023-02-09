package todo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jtprogru/todushka/internal/domain/entity"
	"github.com/jtprogru/todushka/internal/pkg"
)

func (h *Handler) CreateTodo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := make(map[string]any)
		var data entity.TodoCreate
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo bad request: %v", err.Error())
			msg["status"] = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			w.Write(pkg.AnyToJson(msg))
			return
		}

		todo, err := h.srv.CreateTodo(r.Context(), data)
		if err != nil {
			msg["result"] = fmt.Sprintf("todo can't create with err: %v", err.Error())
			msg["status"] = http.StatusInternalServerError
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(pkg.AnyToJson(msg))
			return
		}
		msg["result"] = todo
		msg["status"] = http.StatusCreated
		w.Write(pkg.AnyToJson(msg))
	}
}
