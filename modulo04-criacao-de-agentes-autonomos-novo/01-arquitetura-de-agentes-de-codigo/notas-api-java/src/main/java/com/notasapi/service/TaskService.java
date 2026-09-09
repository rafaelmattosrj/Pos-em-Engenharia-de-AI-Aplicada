package com.notasapi.service;

import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;

import java.util.List;

/** Porte 1:1 da interface TaskService (service/task-service.ts). */
public interface TaskService {

    Task createTask(String title);

    List<Task> listTasks(TaskListFilter filter);

    Task completeTask(String id);

    void removeTask(String id);
}
