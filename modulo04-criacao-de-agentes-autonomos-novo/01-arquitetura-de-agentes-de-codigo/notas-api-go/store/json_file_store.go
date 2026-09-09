package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"notas-api/domain"
)

// PersistenceError porta TaskStorePersistenceError (store/json-file-task-store.ts).
type PersistenceError struct {
	Message string
	Cause   error
}

func (e *PersistenceError) Error() string { return e.Message }
func (e *PersistenceError) Unwrap() error { return e.Cause }

type fileTaskStoreDocument struct {
	Tasks []domain.Task `json:"tasks"`
}

// JSONFileTaskStore porta JsonFileTaskStore (store/json-file-task-store.ts):
// mesma estrategia de persistencia atomica (escreve em arquivo temporario e
// renomeia por cima do arquivo final).
type JSONFileTaskStore struct {
	mu         sync.Mutex
	filePath   string
	tasks      map[string]domain.Task
	order      []string
	generateID func() string
}

func NewJSONFileTaskStore(filePath string) (*JSONFileTaskStore, error) {
	return NewJSONFileTaskStoreWithID(filePath, func() string { return uuid.NewString() })
}

func NewJSONFileTaskStoreWithID(filePath string, generateID func() string) (*JSONFileTaskStore, error) {
	store := &JSONFileTaskStore{
		filePath:   filePath,
		tasks:      make(map[string]domain.Task),
		generateID: generateID,
	}
	if err := store.loadFromFile(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *JSONFileTaskStore) Create(title string) domain.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := domain.Task{ID: s.generateID(), Title: title, Status: domain.StatusOpen}
	s.tasks[task.ID] = task
	s.order = append(s.order, task.ID)
	s.persist()
	return task
}

func (s *JSONFileTaskStore) List(filter domain.TaskListFilter) []domain.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]domain.Task, 0, len(s.order))
	for _, id := range s.order {
		task, ok := s.tasks[id]
		if ok && matchesFilter(task, filter) {
			result = append(result, task)
		}
	}
	return result
}

func (s *JSONFileTaskStore) GetByID(id string) (domain.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	return task, ok
}

func (s *JSONFileTaskStore) Complete(id string) (domain.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return domain.Task{}, false
	}
	task.Status = domain.StatusDone
	s.tasks[id] = task
	s.persist()
	return task, true
}

func (s *JSONFileTaskStore) Remove(id string) bool {
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
	s.persist()
	return true
}

func (s *JSONFileTaskStore) loadFromFile() error {
	raw, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return &PersistenceError{Message: fmt.Sprintf("Failed to read task store file at %q", s.filePath), Cause: err}
	}
	if len(raw) == 0 {
		return nil
	}

	var doc fileTaskStoreDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return &PersistenceError{Message: fmt.Sprintf("Task store file at %q contains invalid JSON", s.filePath), Cause: err}
	}

	for _, task := range doc.Tasks {
		if _, err := domain.RequireNonBlank(task.ID, "id"); err != nil {
			return &PersistenceError{Message: fmt.Sprintf("Task store file at %q has an invalid structure", s.filePath), Cause: err}
		}
		if _, err := domain.RequireNonBlank(task.Title, "title"); err != nil {
			return &PersistenceError{Message: fmt.Sprintf("Task store file at %q has an invalid structure", s.filePath), Cause: err}
		}
		if _, err := domain.ParseTaskStatus(string(task.Status)); err != nil {
			return &PersistenceError{Message: fmt.Sprintf("Task store file at %q has an invalid structure", s.filePath), Cause: err}
		}
		s.tasks[task.ID] = task
		s.order = append(s.order, task.ID)
	}
	return nil
}

func (s *JSONFileTaskStore) persist() {
	directory := filepath.Dir(s.filePath)
	tempFilePath := fmt.Sprintf("%s.%d.%d.tmp", s.filePath, os.Getpid(), time.Now().UnixNano())

	tasks := make([]domain.Task, 0, len(s.order))
	for _, id := range s.order {
		tasks = append(tasks, s.tasks[id])
	}
	content, err := json.MarshalIndent(fileTaskStoreDocument{Tasks: tasks}, "", "  ")
	if err != nil {
		panic(&PersistenceError{Message: "Failed to persist task store file at " + strconv.Quote(s.filePath), Cause: err})
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		panic(&PersistenceError{Message: "Failed to persist task store file at " + strconv.Quote(s.filePath), Cause: err})
	}
	if err := os.WriteFile(tempFilePath, append(content, '\n'), 0o644); err != nil {
		panic(&PersistenceError{Message: "Failed to persist task store file at " + strconv.Quote(s.filePath), Cause: err})
	}
	if err := os.Rename(tempFilePath, s.filePath); err != nil {
		panic(&PersistenceError{Message: "Failed to persist task store file at " + strconv.Quote(s.filePath), Cause: err})
	}
}
