package com.notasapi;

import com.notasapi.service.TaskServiceImpl;
import com.notasapi.store.InMemoryTaskStore;
import com.notasapi.web.TaskHttpHandler;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

import static org.assertj.core.api.Assertions.assertThat;

class TaskHttpHandlerTest {

    private HttpServer server;
    private HttpClient client;
    private String baseUrl;

    @BeforeEach
    void setUp() throws IOException {
        server = HttpServer.create(new InetSocketAddress(0), 0);
        server.createContext("/tasks", new TaskHttpHandler(new TaskServiceImpl(new InMemoryTaskStore())));
        server.setExecutor(null);
        server.start();
        baseUrl = "http://localhost:" + server.getAddress().getPort();
        client = HttpClient.newHttpClient();
    }

    @AfterEach
    void tearDown() {
        server.stop(0);
    }

    @Test
    void createThenListThenComplete() throws Exception {
        HttpResponse<String> created = client.send(
                HttpRequest.newBuilder(URI.create(baseUrl + "/tasks"))
                        .POST(HttpRequest.BodyPublishers.ofString("{\"title\":\"Ler livro\"}"))
                        .header("content-type", "application/json")
                        .build(),
                HttpResponse.BodyHandlers.ofString());
        assertThat(created.statusCode()).isEqualTo(201);
        assertThat(created.body()).contains("\"status\":\"open\"").contains("Ler livro");

        String id = created.body().split("\"id\":\"")[1].split("\"")[0];

        HttpResponse<String> listed = client.send(
                HttpRequest.newBuilder(URI.create(baseUrl + "/tasks?status=open")).GET().build(),
                HttpResponse.BodyHandlers.ofString());
        assertThat(listed.statusCode()).isEqualTo(200);
        assertThat(listed.body()).contains("Ler livro");

        HttpResponse<String> completed = client.send(
                HttpRequest.newBuilder(URI.create(baseUrl + "/tasks/" + id + "/complete"))
                        .method("PATCH", HttpRequest.BodyPublishers.noBody())
                        .build(),
                HttpResponse.BodyHandlers.ofString());
        assertThat(completed.statusCode()).isEqualTo(200);
        assertThat(completed.body()).contains("\"status\":\"done\"");

        HttpResponse<Void> removed = client.send(
                HttpRequest.newBuilder(URI.create(baseUrl + "/tasks/" + id))
                        .DELETE()
                        .build(),
                HttpResponse.BodyHandlers.discarding());
        assertThat(removed.statusCode()).isEqualTo(204);
    }

    @Test
    void blankTitleReturns400() throws Exception {
        HttpResponse<String> response = client.send(
                HttpRequest.newBuilder(URI.create(baseUrl + "/tasks"))
                        .POST(HttpRequest.BodyPublishers.ofString("{\"title\":\"  \"}"))
                        .header("content-type", "application/json")
                        .build(),
                HttpResponse.BodyHandlers.ofString());

        assertThat(response.statusCode()).isEqualTo(400);
        assertThat(response.body()).contains("Validation failed");
    }

    @Test
    void completingUnknownIdReturns404() throws Exception {
        HttpResponse<String> response = client.send(
                HttpRequest.newBuilder(URI.create(baseUrl + "/tasks/does-not-exist/complete"))
                        .method("PATCH", HttpRequest.BodyPublishers.noBody())
                        .build(),
                HttpResponse.BodyHandlers.ofString());

        assertThat(response.statusCode()).isEqualTo(404);
    }
}
