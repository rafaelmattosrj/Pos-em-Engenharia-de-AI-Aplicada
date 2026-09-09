package com.notasapi.cli;

import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;
import com.notasapi.service.TaskNotFoundException;
import com.notasapi.service.TaskService;
import com.notasapi.service.TaskValidationException;
import com.notasapi.store.TaskStorePersistenceException;

import java.io.PrintStream;
import java.util.Arrays;
import java.util.List;

/** Porte 1:1 de cli/commands.ts. */
public final class TaskCli {

    private static final String USAGE = String.join("\n",
            "Usage:",
            "  task create --title <title>",
            "  task list [--status all|open|done]",
            "  task complete --id <task-id>",
            "  task remove --id <task-id>");

    private TaskCli() {
    }

    public static int run(String[] argv, TaskService taskService, PrintStream stdout, PrintStream stderr) {
        try {
            List<String> args = argv.length > 0 && "task".equals(argv[0])
                    ? Arrays.asList(argv).subList(1, argv.length)
                    : Arrays.asList(argv);

            if (args.isEmpty()) {
                throw new CliUsageException("Missing command");
            }

            String commandName = args.get(0);
            List<String> commandArgs = args.subList(1, args.size());

            switch (commandName) {
                case "create" -> {
                    String title = parseFlagValue(commandArgs, "--title", "create");
                    Task task = taskService.createTask(title);
                    stdout.println("Created task: " + formatTask(task));
                    return 0;
                }
                case "list" -> {
                    String status = parseOptionalFlagValue(commandArgs, "--status", "list");
                    TaskListFilter filter = status == null ? TaskListFilter.ALL : TaskListFilter.fromWireValue(status);
                    List<Task> tasks = taskService.listTasks(filter);
                    if (tasks.isEmpty()) {
                        stdout.println("No tasks found.");
                        return 0;
                    }
                    tasks.forEach(task -> stdout.println(formatTask(task)));
                    return 0;
                }
                case "complete" -> {
                    String id = parseFlagValue(commandArgs, "--id", "complete");
                    Task task = taskService.completeTask(id);
                    stdout.println("Completed task: " + formatTask(task));
                    return 0;
                }
                case "remove" -> {
                    String id = parseFlagValue(commandArgs, "--id", "remove");
                    taskService.removeTask(id);
                    stdout.println("Removed task: " + id);
                    return 0;
                }
                default -> throw new CliUsageException("Unknown command \"" + commandName + "\"");
            }
        } catch (CliUsageException | TaskValidationException | TaskNotFoundException e) {
            stderr.println(e.getMessage());
            stderr.println(USAGE);
            return 1;
        } catch (TaskStorePersistenceException e) {
            stderr.println(e.getMessage());
            return 1;
        }
    }

    private static String formatTask(Task task) {
        return task.id() + "\t" + task.status().wireValue() + "\t" + task.title();
    }

    private static String parseFlagValue(List<String> args, String flagName, String commandName) {
        int flagIndex = args.indexOf(flagName);
        if (flagIndex == -1) {
            throw new CliUsageException("Missing required flag " + flagName + " for command \"" + commandName + "\"");
        }
        String value = flagIndex + 1 < args.size() ? args.get(flagIndex + 1) : null;
        if (value == null || value.startsWith("--")) {
            throw new CliUsageException("Missing value for flag " + flagName + " in command \"" + commandName + "\"");
        }
        if (args.size() != flagIndex + 2) {
            throw new CliUsageException("Unexpected extra arguments for command \"" + commandName + "\"");
        }
        return value;
    }

    private static String parseOptionalFlagValue(List<String> args, String flagName, String commandName) {
        if (args.isEmpty()) {
            return null;
        }
        int flagIndex = args.indexOf(flagName);
        if (flagIndex == -1 || flagIndex != 0 || args.size() != 2) {
            throw new CliUsageException("Invalid arguments for command \"" + commandName + "\"");
        }
        String value = args.get(1);
        if (value == null || value.startsWith("--")) {
            throw new CliUsageException("Missing value for flag " + flagName + " in command \"" + commandName + "\"");
        }
        return value;
    }
}
