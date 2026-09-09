package com.notasapi.cli;

import com.notasapi.service.TaskServiceImpl;
import com.notasapi.store.JsonFileTaskStore;
import com.notasapi.store.TaskStorePersistenceException;

import java.nio.file.Path;

/** Porte de src/cli.ts -- CLI sobre um JsonFileTaskStore. */
public final class CliMain {

    private CliMain() {
    }

    public static void main(String[] args) {
        String storePathEnv = System.getenv("TASK_CLI_STORE_PATH");
        Path storePath = storePathEnv != null
                ? Path.of(storePathEnv)
                : Path.of(System.getProperty("user.dir"), ".tasks-cli-store.json");

        try {
            var taskService = new TaskServiceImpl(new JsonFileTaskStore(storePath));
            int exitCode = TaskCli.run(args, taskService, System.out, System.err);
            System.exit(exitCode);
        } catch (TaskStorePersistenceException e) {
            System.err.println(e.getMessage());
            System.exit(1);
        }
    }
}
