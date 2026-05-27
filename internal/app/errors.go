package app

import "errors"

var (
	ErrEmptyTitle          = errors.New("app: empty title")
	ErrDeadlineBeforeStart = errors.New("app: deadline before start date")
	ErrAreaNotEmpty        = errors.New("app: area has children, pass confirm=true")
	ErrAmbiguousProject    = errors.New("app: ambiguous project name")
	ErrUnknownProject      = errors.New("app: unknown project name")
	ErrTagAlreadyExists    = errors.New("app: tag already exists")
	ErrTaskNotFound        = errors.New("app: task not found")
	ErrEmptyInput          = errors.New("app: empty input")
	ErrSchemaTooNew        = errors.New("app: import schema is newer than this binary")
	ErrInvalidImport       = errors.New("app: invalid import payload")
	ErrProjectNotEmpty     = errors.New("app: project has active tasks; confirm required")
	ErrProjectNotFound     = errors.New("app: project not found")
)
