package repository

type repo struct {
	st StorageTodo
	sp StorageProject
}

func New(st StorageTodo, sp StorageProject) *repo {
	return &repo{
		st: st,
		sp: sp,
	}
}
