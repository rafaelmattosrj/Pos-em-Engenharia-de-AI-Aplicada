// Package domain é o porte 1:1 de src/domain/task.ts.
package domain

import (
	"errors"
	"strings"
)

type TaskStatus string

const (
	StatusOpen TaskStatus = "open"
	StatusDone TaskStatus = "done"
)

func ParseTaskStatus(value string) (TaskStatus, error) {
	switch TaskStatus(value) {
	case StatusOpen:
		return StatusOpen, nil
	case StatusDone:
		return StatusDone, nil
	default:
		return "", errors.New("invalid task status: " + value)
	}
}

type TaskListFilter string

const (
	FilterAll  TaskListFilter = "all"
	FilterOpen TaskListFilter = "open"
	FilterDone TaskListFilter = "done"
)

func ParseTaskListFilter(value string) (TaskListFilter, error) {
	switch TaskListFilter(value) {
	case FilterAll:
		return FilterAll, nil
	case FilterOpen:
		return FilterOpen, nil
	case FilterDone:
		return FilterDone, nil
	default:
		return "", errors.New("invalid task list filter: " + value)
	}
}

type Task struct {
	ID     string     `json:"id"`
	Title  string     `json:"title"`
	Status TaskStatus `json:"status"`
}

// RequireNonBlank porta a validação `.trim().min(1)` usada em taskIdSchema/taskTitleSchema.
func RequireNonBlank(value string, message string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", &ValidationError{Issues: []string{message}}
	}
	return trimmed, nil
}

// ValidationError porta TaskValidationError (service/task-errors.ts).
type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Issues, "; ")
}

// NotFoundError porta TaskNotFoundError (service/task-errors.ts).
type NotFoundError struct {
	TaskID string
}

func (e *NotFoundError) Error() string {
	return "Task \"" + e.TaskID + "\" was not found"
}
