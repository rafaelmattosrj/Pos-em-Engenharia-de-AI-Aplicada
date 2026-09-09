// Package service porta service/task-service.ts.
package service

import (
	"notas-api/domain"
	"notas-api/store"
)

type TaskService struct {
	store store.TaskStore
}

func New(store store.TaskStore) *TaskService {
	return &TaskService{store: store}
}

func (s *TaskService) CreateTask(title string) (domain.Task, error) {
	parsedTitle, err := domain.RequireNonBlank(title, "Task title is required")
	if err != nil {
		return domain.Task{}, err
	}
	return s.store.Create(parsedTitle), nil
}

func (s *TaskService) ListTasks(filter domain.TaskListFilter) []domain.Task {
	if filter == "" {
		filter = domain.FilterAll
	}
	return s.store.List(filter)
}

func (s *TaskService) CompleteTask(id string) (domain.Task, error) {
	parsedID, err := domain.RequireNonBlank(id, "Task id is required")
	if err != nil {
		return domain.Task{}, err
	}
	task, ok := s.store.Complete(parsedID)
	if !ok {
		return domain.Task{}, &domain.NotFoundError{TaskID: parsedID}
	}
	return task, nil
}

func (s *TaskService) RemoveTask(id string) error {
	parsedID, err := domain.RequireNonBlank(id, "Task id is required")
	if err != nil {
		return err
	}
	if !s.store.Remove(parsedID) {
		return &domain.NotFoundError{TaskID: parsedID}
	}
	return nil
}
