package todo

import (
	"net/http"

	"github.com/jtprogru/todushka/internal/pkg"
)

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	response["message"] = "Deleted TODO successfully"

	w.Write(pkg.AnyToJson(response)) // Return some demo response
}
