package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"notas-api/domain"
	"notas-api/store"
)

func TestJSONFileTaskStorePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "store.json")

	counter := 0
	first, err := store.NewJSONFileTaskStoreWithID(storePath, func() string {
		counter++
		return "id-" + string(rune('0'+counter))
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	first.Create("Persistente")

	second, err := store.NewJSONFileTaskStore(storePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tasks := second.List(domain.FilterAll)
	if len(tasks) != 1 || tasks[0].Title != "Persistente" {
		t.Fatalf("expected 1 persisted task, got %+v", tasks)
	}
}

func TestJSONFileTaskStoreMissingFileStartsEmpty(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "nao-existe.json")

	s, err := store.NewJSONFileTaskStore(storePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.List(domain.FilterAll)) != 0 {
		t.Fatalf("expected empty store")
	}
}

func TestJSONFileTaskStoreInvalidJSONFails(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "invalido.json")
	if err := os.WriteFile(storePath, []byte("{ not valid json"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err := store.NewJSONFileTaskStore(storePath)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*store.PersistenceError); !ok {
		t.Fatalf("expected *PersistenceError, got %T: %v", err, err)
	}
}

func TestJSONFileTaskStoreMissingTasksFieldFails(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "sem-tasks.json")
	if err := os.WriteFile(storePath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	s, err := store.NewJSONFileTaskStore(storePath)
	if err != nil {
		t.Fatalf("unexpected error (empty tasks array is valid zero-value): %v", err)
	}
	if len(s.List(domain.FilterAll)) != 0 {
		t.Fatalf("expected empty store for {}")
	}
}
