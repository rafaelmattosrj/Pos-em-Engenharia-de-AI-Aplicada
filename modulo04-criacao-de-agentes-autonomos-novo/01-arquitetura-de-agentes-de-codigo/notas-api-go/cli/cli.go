// Package cli porta cli/commands.ts.
package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"notas-api/domain"
	"notas-api/service"
)

const usage = "Usage:\n" +
	"  task create --title <title>\n" +
	"  task list [--status all|open|done]\n" +
	"  task complete --id <task-id>\n" +
	"  task remove --id <task-id>"

type usageError struct{ message string }

func (e *usageError) Error() string { return e.message }

// Run porta runTaskCli. Retorna o exit code (0 ou 1), como o original.
func Run(argv []string, taskService *service.TaskService, stdout, stderr io.Writer) int {
	args := argv
	if len(args) > 0 && args[0] == "task" {
		args = args[1:]
	}

	err := run(args, taskService, stdout)
	if err == nil {
		return 0
	}

	var usageErr *usageError
	var validationErr *domain.ValidationError
	var notFoundErr *domain.NotFoundError
	switch {
	case errors.As(err, &usageErr), errors.As(err, &validationErr), errors.As(err, &notFoundErr):
		fmt.Fprintln(stderr, err.Error())
		fmt.Fprintln(stderr, usage)
	default:
		fmt.Fprintln(stderr, err.Error())
	}
	return 1
}

func run(args []string, taskService *service.TaskService, stdout io.Writer) error {
	if len(args) == 0 {
		return &usageError{"Missing command"}
	}
	commandName := args[0]
	commandArgs := args[1:]

	switch commandName {
	case "create":
		title, err := parseFlagValue(commandArgs, "--title", "create")
		if err != nil {
			return err
		}
		task, err := taskService.CreateTask(title)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Created task: %s\n", formatTask(task))
		return nil

	case "list":
		statusValue, err := parseOptionalFlagValue(commandArgs, "--status", "list")
		if err != nil {
			return err
		}
		filter := domain.FilterAll
		if statusValue != "" {
			filter, err = domain.ParseTaskListFilter(statusValue)
			if err != nil {
				return err
			}
		}
		tasks := taskService.ListTasks(filter)
		if len(tasks) == 0 {
			fmt.Fprintln(stdout, "No tasks found.")
			return nil
		}
		for _, task := range tasks {
			fmt.Fprintln(stdout, formatTask(task))
		}
		return nil

	case "complete":
		id, err := parseFlagValue(commandArgs, "--id", "complete")
		if err != nil {
			return err
		}
		task, err := taskService.CompleteTask(id)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Completed task: %s\n", formatTask(task))
		return nil

	case "remove":
		id, err := parseFlagValue(commandArgs, "--id", "remove")
		if err != nil {
			return err
		}
		if err := taskService.RemoveTask(id); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Removed task: %s\n", id)
		return nil

	default:
		return &usageError{fmt.Sprintf("Unknown command %q", commandName)}
	}
}

func formatTask(task domain.Task) string {
	return fmt.Sprintf("%s\t%s\t%s", task.ID, task.Status, task.Title)
}

func indexOf(args []string, value string) int {
	for i, arg := range args {
		if arg == value {
			return i
		}
	}
	return -1
}

func parseFlagValue(args []string, flagName, commandName string) (string, error) {
	flagIndex := indexOf(args, flagName)
	if flagIndex == -1 {
		return "", &usageError{fmt.Sprintf("Missing required flag %s for command %q", flagName, commandName)}
	}
	if flagIndex+1 >= len(args) || strings.HasPrefix(args[flagIndex+1], "--") {
		return "", &usageError{fmt.Sprintf("Missing value for flag %s in command %q", flagName, commandName)}
	}
	if len(args) != flagIndex+2 {
		return "", &usageError{fmt.Sprintf("Unexpected extra arguments for command %q", commandName)}
	}
	return args[flagIndex+1], nil
}

func parseOptionalFlagValue(args []string, flagName, commandName string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	flagIndex := indexOf(args, flagName)
	if flagIndex == -1 || flagIndex != 0 || len(args) != 2 {
		return "", &usageError{fmt.Sprintf("Invalid arguments for command %q", commandName)}
	}
	if strings.HasPrefix(args[1], "--") {
		return "", &usageError{fmt.Sprintf("Missing value for flag %s in command %q", flagName, commandName)}
	}
	return args[1], nil
}
