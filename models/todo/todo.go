package todo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var (
	ListTodos = map[int]Todo{
		1: {
			Id:          1,
			Summary:     "Todo 1 Hello world",
			Description: "Todo 1 Hello world from planet earth",
			IsDone:      false,
			ListId:      "1",
		},
		2: {
			Id:          2,
			Summary:     "Todo 2 Hello world",
			Description: "Todo 2 Hello world from planet earth",
			IsDone:      false,
			ListId:      "1",
		},
		3: {
			Id:          3,
			Summary:     "Todo 3 Hello world",
			Description: "Todo 3 Hello world from planet earth",
			IsDone:      false,
			ListId:      "2",
		},
		4: {
			Id:          4,
			Summary:     "Todo 4 Hello world",
			Description: "Todo 4 Hello world from planet earth",
			IsDone:      false,
			ListId:      "2",
		},
	}
)

func todoToJson(t any) []byte {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

type Todo struct {
	Id          int    `json:"id" db:"id"`
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
	ListId      string `json:"list_id"`
}

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/{todoID}", GetTodo)
	router.Delete("/{todoID}", DeleteTodo)
	router.Post("/", CreateTodo)
	router.Get("/", GetAllTodos)
	return router
}

func GetTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	todoID, err := strconv.Atoi(chi.URLParam(r, "todoID"))
	if err != nil {
		w.Write([]byte(`url parameter todoID is not number`))
		return
	}
	t := ListTodos[todoID]
	w.Write(todoToJson(t))
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	response["message"] = "Deleted TODO successfully"

	w.Write(todoToJson(response)) // Return some demo response
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)

	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
	}
	data := Todo{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp["message"] = "can't parse TODO object"
		w.Write(todoToJson(resp))
		return
	}
	w.WriteHeader(http.StatusCreated)
	ListTodos[data.Id] = data
	// response := make(map[string]string)
	// response["message"] = "Created TODO successfully"
	w.Write(todoToJson(data))
}

func GetAllTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	todos := []Todo{
		ListTodos[1],
		ListTodos[2],
		ListTodos[3],
		ListTodos[4],
	}
	w.Write(todoToJson(todos))
}
