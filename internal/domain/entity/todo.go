package entity

type Todo struct {
	Id          int    `json:"id" db:"id"`
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
	ListId      string `json:"list_id"`
}
