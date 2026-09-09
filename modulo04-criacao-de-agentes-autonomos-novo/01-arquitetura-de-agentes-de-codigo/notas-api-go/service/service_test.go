package service_test

import (
	"errors"
	"testing"

	"notas-api/domain"
	"notas-api/service"
	"notas-api/store"
)

func newService(t *testing.T) *service.TaskService {
	t.Helper()
	counter := 0
	return service.New(store.NewInMemoryTaskStoreWithID(func() string {
		counter++
		return "id-" + string(rune('0'+counter))
	}))
}

func TestCreateTaskIsOpen(t *testing.T) {
	svc := newService(t)
	task, err := svc.CreateTask("Comprar leite")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != domain.StatusOpen {
		t.Errorf("expected status open, got %s", task.Status)
	}
	if task.Title != "Comprar leite" {
		t.Errorf("expected title 'Comprar leite', got %s", task.Title)
	}
}

func TestCreateTaskRejectsBlankTitle(t *testing.T) {
	svc := newService(t)
	_, err := svc.CreateTask("   ")
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestListTasksFiltersByStatus(t *testing.T) {
	svc := newService(t)
	svc.CreateTask("A")
	b, _ := svc.CreateTask("B")
	if _, err := svc.CompleteTask(b.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	open := svc.ListTasks(domain.FilterOpen)
	done := svc.ListTasks(domain.FilterDone)
	all := svc.ListTasks(domain.FilterAll)

	if len(open) != 1 || open[0].Title != "A" {
		t.Errorf("expected only A open, got %+v", open)
	}
	if len(done) != 1 || done[0].Title != "B" {
		t.Errorf("expected only B done, got %+v", done)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 tasks total, got %d", len(all))
	}
}

func TestCompleteUnknownTaskReturnsNotFound(t *testing.T) {
	svc := newService(t)
	_, err := svc.CompleteTask("does-not-exist")
	var notFoundErr *domain.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestRemoveUnknownTaskReturnsNotFound(t *testing.T) {
	svc := newService(t)
	err := svc.RemoveTask("does-not-exist")
	var notFoundErr *domain.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestRemoveDeletesTask(t *testing.T) {
	svc := newService(t)
	task, _ := svc.CreateTask("Descartavel")
	if err := svc.RemoveTask(task.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(svc.ListTasks(domain.FilterAll)) != 0 {
		t.Errorf("expected no tasks left")
	}
}

func TestCompleteIsIdempotent(t *testing.T) {
	svc := newService(t)
	task, _ := svc.CreateTask("Idempotente")
	first, err := svc.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := svc.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.Status != domain.StatusDone || second.Status != domain.StatusDone {
		t.Errorf("expected both completions to be done")
	}
}
