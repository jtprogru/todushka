package project

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Repository interface {
	GetProjectByID(ctx context.Context, projectID int) (entity.Project, error)
	GetAllProjects(ctx context.Context) ([]entity.Project, error)
	DeleteProject(ctx context.Context, projectID int) error
	CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error)
	UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error)
}

type srv struct {
	repo Repository
}

func New(repo Repository) *srv {
	return &srv{
		repo: repo,
	}
}

func (s *srv) GetProjectByID(ctx context.Context, projectID int) (entity.Project, error) {
	return s.repo.GetProjectByID(ctx, projectID)
}

func (s *srv) GetAllProjects(ctx context.Context) ([]entity.Project, error) {
	return s.repo.GetAllProjects(ctx)
}

func (s *srv) DeleteProject(ctx context.Context, projectID int) error {
	return s.repo.DeleteProject(ctx, projectID)
}

func (s *srv) CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error) {
	return s.repo.CreateProject(ctx, project)
}

func (s *srv) UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error) {
	return s.repo.UpdateProject(ctx, projectID, project)
}
