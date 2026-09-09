// Package store porta store/task-store.ts e suas duas implementações.
package store

import "notas-api/domain"

type TaskStore interface {
	Create(title string) domain.Task
	List(filter domain.TaskListFilter) []domain.Task
	GetByID(id string) (domain.Task, bool)
	Complete(id string) (domain.Task, bool)
	Remove(id string) bool
}
