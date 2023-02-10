package entity

type Project struct {
	ID          int    `json:"id" db:"id"`
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
}

type ProjectCreate struct {
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
}

type ProjectUpdate struct {
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description,omitempty" db:"description"`
	IsDone      bool   `json:"is_done" db:"is_done"`
}
