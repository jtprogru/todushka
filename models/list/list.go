package list

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var (
	ListLists = map[string]List{
		"1": {
			Id:          "1",
			Summary:     "List 1 Hello world",
			Description: "List 1 Hell world from planet earth",
			TodoIds:     []string{"1", "2"},
			ProjectId:   "",
		},
		"2": {
			Id:          "2",
			Summary:     "List 2 Hello world",
			Description: "List 2 Hell world from planet earth",
			TodoIds:     []string{"3", "4"},
			ProjectId:   "",
		},
	}
)

type List struct {
	Id          string   `json:"id"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	TodoIds     []string `json:"todo_ids"` // список задач todo, которые входят в список;
	ProjectId   string   `json:"project_id"`
}

func listToJson(t any) []byte {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/{listID}", GetList)
	router.Delete("/{listID}", DeleteList)
	router.Post("/", CreateList)
	router.Get("/", GetAllLists)
	return router
}

func GetList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	listID := chi.URLParam(r, "listID")
	l := ListLists[listID]
	w.Write(listToJson(l))
}

func DeleteList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	response["message"] = "Deleted LIST successfully"

	w.Write(listToJson(response)) // Return some demo response
}

func CreateList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	response["message"] = "Created LIST successfully"
	w.Write(listToJson(response))
}

func GetAllLists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	lists := []List{
		ListLists["1"],
		ListLists["2"],
	}
	w.Write(listToJson(lists))
}
