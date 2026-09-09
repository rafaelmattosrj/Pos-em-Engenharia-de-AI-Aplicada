package store

import (
	"sync"

	"github.com/google/uuid"

	"notas-api/domain"
)

// InMemoryTaskStore porta store/in-memory-task-store.ts.
type InMemoryTaskStore struct {
	mu         sync.Mutex
	tasks      map[string]domain.Task
	order      []string
	generateID func() string
}

func NewInMemoryTaskStore() *InMemoryTaskStore {
	return NewInMemoryTaskStoreWithID(func() string { return uuid.NewString() })
}

func NewInMemoryTaskStoreWithID(generateID func() string) *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks:      make(map[string]domain.Task),
		generateID: generateID,
	}
}

func (s *InMemoryTaskStore) Create(title string) domain.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := domain.Task{ID: s.generateID(), Title: title, Status: domain.StatusOpen}
	s.tasks[task.ID] = task
	s.order = append(s.order, task.ID)
	return task
}

func (s *InMemoryTaskStore) List(filter domain.TaskListFilter) []domain.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]domain.Task, 0, len(s.order))
	for _, id := range s.order {
		task, ok := s.tasks[id]
		if !ok {
			continue
		}
		if matchesFilter(task, filter) {
			result = append(result, task)
		}
	}
	return result
}

func (s *InMemoryTaskStore) GetByID(id string) (domain.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	return task, ok
}

func (s *InMemoryTaskStore) Complete(id string) (domain.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return domain.Task{}, false
	}
	task.Status = domain.StatusDone
	s.tasks[id] = task
	return task, true
}

func (s *InMemoryTaskStore) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	for i, orderedID := range s.order {
		if orderedID == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

func matchesFilter(task domain.Task, filter domain.TaskListFilter) bool {
	switch filter {
	case domain.FilterAll, "":
		return true
	case domain.FilterOpen:
		return task.Status == domain.StatusOpen
	case domain.FilterDone:
		return task.Status == domain.StatusDone
	default:
		return false
	}
}
