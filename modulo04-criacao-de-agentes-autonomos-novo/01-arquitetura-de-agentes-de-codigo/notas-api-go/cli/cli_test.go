package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"notas-api/cli"
	"notas-api/service"
	"notas-api/store"
)

func newTestService() *service.TaskService {
	return service.New(store.NewInMemoryTaskStore())
}

func TestCreatePrintsCreatedTask(t *testing.T) {
	var stdout, stderr bytes.Buffer
	svc := newTestService()

	exitCode := cli.Run([]string{"create", "--title", "Estudar"}, svc, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Created task:") || !strings.Contains(stdout.String(), "Estudar") {
		t.Errorf("unexpected stdout: %s", stdout.String())
	}
}

func TestListWithoutTasksPrintsMessage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	svc := newTestService()

	exitCode := cli.Run([]string{"list"}, svc, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "No tasks found.") {
		t.Errorf("unexpected stdout: %s", stdout.String())
	}
}

func TestMissingCommandIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	svc := newTestService()

	exitCode := cli.Run([]string{}, svc, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Missing command") || !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("unexpected stderr: %s", stderr.String())
	}
}

func TestCompleteUnknownIdIsNotFoundError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	svc := newTestService()

	exitCode := cli.Run([]string{"complete", "--id", "ghost"}, svc, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "was not found") {
		t.Errorf("unexpected stderr: %s", stderr.String())
	}
}

func TestTaskPrefixIsStripped(t *testing.T) {
	var stdout, stderr bytes.Buffer
	svc := newTestService()
	task, _ := svc.CreateTask("Prefixo")

	exitCode := cli.Run([]string{"task", "complete", "--id", task.ID}, svc, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Completed task:") {
		t.Errorf("unexpected stdout: %s", stdout.String())
	}
}

func TestUnknownCommandIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	svc := newTestService()

	exitCode := cli.Run([]string{"unknown"}, svc, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Unknown command") {
		t.Errorf("unexpected stderr: %s", stderr.String())
	}
}
