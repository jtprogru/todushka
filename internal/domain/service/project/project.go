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
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return entity.Project{}, err
	}
	return project, nil
}

func (s *srv) GetAllProjects(ctx context.Context) ([]entity.Project, error) {
	projects, err := s.repo.GetAllProjects(ctx)
	if err != nil {
		return []entity.Project{}, err
	}
	return projects, nil
}

func (s *srv) DeleteProject(ctx context.Context, projectID int) error {
	err := s.repo.DeleteProject(ctx, projectID)
	if err != nil {
		return err
	}
	return nil
}

func (s *srv) CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error) {
	t, err := s.repo.CreateProject(ctx, project)
	if err != nil {
		return entity.Project{}, err
	}
	return t, nil
}

func (s *srv) UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error) {
	t, err := s.repo.UpdateProject(ctx, projectID, project)
	if err != nil {
		return entity.Project{}, err
	}
	return t, nil
}
