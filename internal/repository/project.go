package repository

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type StorageProject interface {
	GetProjectByID(ctx context.Context, projectID int) (entity.Project, error)
	GetAllProjects(ctx context.Context) ([]entity.Project, error)
	DeleteProject(ctx context.Context, projectID int) error
	CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error)
	UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error)
}

func (r *repo) GetProjectByID(ctx context.Context, projectID int) (entity.Project, error) {
	project, err := r.sp.GetProjectByID(ctx, projectID)
	if err != nil {
		return entity.Project{}, err
	}
	return project, nil
}

func (r *repo) GetAllProjects(ctx context.Context) ([]entity.Project, error) {
	projects, err := r.sp.GetAllProjects(ctx)
	if err != nil {
		return []entity.Project{}, err
	}
	return projects, nil
}

func (r *repo) DeleteProject(ctx context.Context, projectID int) error {
	err := r.sp.DeleteProject(ctx, projectID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error) {
	t, err := r.sp.CreateProject(ctx, project)
	if err != nil {
		return entity.Project{}, err
	}
	return t, nil
}

func (r *repo) UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error) {
	t, err := r.sp.UpdateProject(ctx, projectID, project)
	if err != nil {
		return entity.Project{}, err
	}
	return t, nil
}
