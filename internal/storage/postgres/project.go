package postgres

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

func (s *storage) GetProjectByID(ctx context.Context, projectID int) (entity.Project, error) {
	query := `select id, summary, description, is_done from project_items where id=$1;`
	var project entity.Project
	err := s.db.Get(&project, query, projectID)
	if err != nil {
		return project, err
	}
	return project, nil
}

func (s *storage) GetAllProjects(ctx context.Context) ([]entity.Project, error) {
	query := `select id, summary, description, is_done from project_items;`
	var projects []entity.Project
	err := s.db.SelectContext(ctx, &projects, query)
	if err != nil {
		return projects, nil
	}
	return projects, nil
}

func (s *storage) DeleteProject(ctx context.Context, projectID int) error {
	query := `delete from project_items where id=$1;`
	_, err := s.db.Exec(query, projectID)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error) {
	query := `insert into project_items (summary, description) values ($1, $2) returning id;`
	var projectID int
	err := s.db.QueryRow(query, project.Summary, project.Description).Scan(&projectID)
	if err != nil {
		return entity.Project{}, err
	}
	resp := entity.Project{
		ID:          projectID,
		Summary:     project.Summary,
		Description: project.Description,
		IsDone:      false,
	}

	return resp, nil
}

func (s *storage) UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error) {
	query := `update project_items
    set summary = $1, description = $2, is_done = $3
    where id = $4 returning id, summary, description, is_done;`
	var resp entity.Project

	err := s.db.QueryRow(
		query,
		project.Summary,
		project.Description,
		project.IsDone,
		projectID).Scan(&resp.ID, &resp.Summary, &resp.Description, &resp.IsDone)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
