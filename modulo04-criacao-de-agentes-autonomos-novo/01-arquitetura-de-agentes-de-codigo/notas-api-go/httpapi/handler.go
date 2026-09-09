// Package httpapi porta http/task-routes.ts e http/http-errors.ts.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"

	"notas-api/domain"
	"notas-api/service"
)

var (
	completePathRe = regexp.MustCompile(`^/tasks/([^/]+)/complete$`)
	taskPathRe     = regexp.MustCompile(`^/tasks/([^/]+)$`)
)

type createTaskBody struct {
	Title string `json:"title"`
}

type errorBody struct {
	Error  string   `json:"error"`
	Issues []string `json:"issues,omitempty"`
}

func NewHandler(taskService *service.TaskService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case r.Method == http.MethodPost && path == "/tasks":
			handleCreate(w, r, taskService)
		case r.Method == http.MethodGet && path == "/tasks":
			handleList(w, r, taskService)
		case r.Method == http.MethodPatch && completePathRe.MatchString(path):
			id, _ := url.PathUnescape(completePathRe.FindStringSubmatch(path)[1])
			handleComplete(w, taskService, id)
		case r.Method == http.MethodDelete && taskPathRe.MatchString(path):
			id, _ := url.PathUnescape(taskPathRe.FindStringSubmatch(path)[1])
			handleRemove(w, taskService, id)
		default:
			sendJSON(w, http.StatusNotFound, errorBody{Error: "Route not found"})
		}
	})
}

func handleCreate(w http.ResponseWriter, r *http.Request, taskService *service.TaskService) {
	var body createTaskBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendJSON(w, http.StatusBadRequest, errorBody{Error: "Invalid JSON body"})
		return
	}
	task, err := taskService.CreateTask(body.Title)
	if err != nil {
		writeError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, task)
}

func handleList(w http.ResponseWriter, r *http.Request, taskService *service.TaskService) {
	statusParam := r.URL.Query().Get("status")
	if statusParam == "" {
		statusParam = "all"
	}
	filter, err := domain.ParseTaskListFilter(statusParam)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, errorBody{Error: "Validation failed", Issues: []string{err.Error()}})
		return
	}
	tasks := taskService.ListTasks(filter)
	sendJSON(w, http.StatusOK, tasks)
}

func handleComplete(w http.ResponseWriter, taskService *service.TaskService, id string) {
	task, err := taskService.CompleteTask(id)
	if err != nil {
		writeError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, task)
}

func handleRemove(w http.ResponseWriter, taskService *service.TaskService, id string) {
	if err := taskService.RemoveTask(id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		sendJSON(w, http.StatusBadRequest, errorBody{Error: "Validation failed", Issues: validationErr.Issues})
		return
	}
	var notFoundErr *domain.NotFoundError
	if errors.As(err, &notFoundErr) {
		sendJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusInternalServerError, errorBody{Error: "Internal server error"})
}

func sendJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
