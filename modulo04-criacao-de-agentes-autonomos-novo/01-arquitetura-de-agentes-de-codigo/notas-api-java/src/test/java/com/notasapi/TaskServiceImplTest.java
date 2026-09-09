package com.notasapi;

import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;
import com.notasapi.domain.TaskStatus;
import com.notasapi.service.TaskNotFoundException;
import com.notasapi.service.TaskService;
import com.notasapi.service.TaskServiceImpl;
import com.notasapi.service.TaskValidationException;
import com.notasapi.store.InMemoryTaskStore;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class TaskServiceImplTest {

    private TaskService taskService;

    @BeforeEach
    void setUp() {
        AtomicInteger counter = new AtomicInteger();
        taskService = new TaskServiceImpl(new InMemoryTaskStore(() -> "id-" + counter.incrementAndGet()));
    }

    @Test
    void createsTaskAsOpen() {
        Task task = taskService.createTask("Comprar leite");

        assertThat(task.title()).isEqualTo("Comprar leite");
        assertThat(task.status()).isEqualTo(TaskStatus.OPEN);
        assertThat(task.id()).isEqualTo("id-1");
    }

    @Test
    void rejectsBlankTitle() {
        assertThatThrownBy(() -> taskService.createTask("   "))
                .isInstanceOf(TaskValidationException.class);
    }

    @Test
    void listsOnlyMatchingFilter() {
        taskService.createTask("A");
        Task b = taskService.createTask("B");
        taskService.completeTask(b.id());

        assertThat(taskService.listTasks(TaskListFilter.ALL)).hasSize(2);
        assertThat(taskService.listTasks(TaskListFilter.OPEN)).extracting(Task::title).containsExactly("A");
        assertThat(taskService.listTasks(TaskListFilter.DONE)).extracting(Task::title).containsExactly("B");
    }

    @Test
    void completingUnknownTaskThrowsNotFound() {
        assertThatThrownBy(() -> taskService.completeTask("does-not-exist"))
                .isInstanceOf(TaskNotFoundException.class);
    }

    @Test
    void removingUnknownTaskThrowsNotFound() {
        assertThatThrownBy(() -> taskService.removeTask("does-not-exist"))
                .isInstanceOf(TaskNotFoundException.class);
    }

    @Test
    void removeDeletesTask() {
        Task task = taskService.createTask("Descartavel");
        taskService.removeTask(task.id());

        assertThat(taskService.listTasks(TaskListFilter.ALL)).isEmpty();
    }

    @Test
    void completingIsIdempotent() {
        Task task = taskService.createTask("Idempotente");
        Task first = taskService.completeTask(task.id());
        Task second = taskService.completeTask(task.id());

        assertThat(first.status()).isEqualTo(TaskStatus.DONE);
        assertThat(second.status()).isEqualTo(TaskStatus.DONE);
    }

    @Test
    void allListMeansNoFilter() {
        assertThat(taskService.listTasks(null)).isEqualTo(List.of());
    }
}
