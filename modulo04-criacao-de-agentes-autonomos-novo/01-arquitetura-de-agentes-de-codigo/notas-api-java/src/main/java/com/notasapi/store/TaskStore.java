package com.notasapi.store;

import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;

import java.util.List;
import java.util.Optional;

/** Porte 1:1 da interface TaskStore (store/task-store.ts). */
public interface TaskStore {

    Task create(String title);

    List<Task> list(TaskListFilter filter);

    Optional<Task> getById(String id);

    Optional<Task> complete(String id);

    boolean remove(String id);
}
