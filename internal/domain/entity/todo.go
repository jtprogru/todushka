package entity

type Todo struct {
	ID          int    `json:"id" db:"id"`
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
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
