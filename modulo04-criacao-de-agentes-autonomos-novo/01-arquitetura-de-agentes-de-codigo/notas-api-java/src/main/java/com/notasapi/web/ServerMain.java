package com.notasapi.web;

import com.notasapi.service.TaskService;
import com.notasapi.service.TaskServiceImpl;
import com.notasapi.store.InMemoryTaskStore;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.net.InetSocketAddress;

/** Porte de src/index.ts -- sobe o servidor HTTP na porta 3000. */
public final class ServerMain {

    private ServerMain() {
    }

    public static void main(String[] args) throws IOException {
        int port = 3000;
        TaskService taskService = new TaskServiceImpl(new InMemoryTaskStore());

        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/tasks", new TaskHttpHandler(taskService));
        server.setExecutor(null);
        server.start();

        System.out.println("HTTP server listening on http://localhost:" + port);
    }
}
