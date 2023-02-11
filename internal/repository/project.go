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
	return r.sp.GetProjectByID(ctx, projectID)
}

func (r *repo) GetAllProjects(ctx context.Context) ([]entity.Project, error) {
	return r.sp.GetAllProjects(ctx)
}

func (r *repo) DeleteProject(ctx context.Context, projectID int) error {
	return r.sp.DeleteProject(ctx, projectID)
}

func (r *repo) CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error) {
	return r.sp.CreateProject(ctx, project)
}

func (r *repo) UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error) {
	return r.sp.UpdateProject(ctx, projectID, project)
}
