package entity

type Todo struct {
	Id          int    `json:"id" db:"id"`
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
	// ListId      int    `json:"list_id,omitempty" db:"list_id"`
}

type TodoCreate struct {
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
}

type TodoUpdate struct {
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
}
