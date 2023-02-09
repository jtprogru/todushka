package todo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jtprogru/todushka/internal/domain/entity"
	"github.com/jtprogru/todushka/internal/pkg"
)

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)

	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
	}
	data := entity.Todo{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp["message"] = "can't parse TODO object"
		w.Write(pkg.AnyToJson(resp))
		return
	}
	w.WriteHeader(http.StatusCreated)
	// ListTodos[data.Id] = data
	// response := make(map[string]string)
	// response["message"] = "Created TODO successfully"
	w.Write(pkg.AnyToJson(data))
}
