package project

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var (
	ListProjects = map[string]Project{
		"1": {
			Id:          "1",
			Summary:     "Project 1 Hello world",
			Description: "Project 1 Hell world from planet earth",
			TodoIds:     []string{"1", "2"},
			IsDone:      false,
			ListId:      "1",
		},
		"2": {
			Id:          "2",
			Summary:     "Project 2 Hello world",
			Description: "Project 2 Hell world from planet earth",
			TodoIds:     []string{"3", "4"},
			IsDone:      false,
			ListId:      "2",
		},
	}
)

type Project struct {
	Id          string   `json:"id"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	TodoIds     []string `json:"todo_ids"` // список задач todo, которые входят в проект;
	IsDone      bool     `json:"is_done"`
	ListId      string   `json:"list_id"`
}

func projectToJson(t any) []byte {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/{projectID}", GetProject)
	router.Delete("/{projectID}", DeleteProject)
	router.Post("/", CreateProject)
	router.Get("/", GetAllProjects)
	return router
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	projectID := chi.URLParam(r, "projectID")
	l := ListProjects[projectID]
	w.Write(projectToJson(l))
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	response["message"] = "Deleted LIST successfully"

	w.Write(projectToJson(response)) // Return some demo response
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	response["message"] = "Created LIST successfully"
	w.Write(projectToJson(response))
}

func GetAllProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	projects := []Project{
		ListProjects["1"],
		ListProjects["2"],
	}
	w.Write(projectToJson(projects))
}
