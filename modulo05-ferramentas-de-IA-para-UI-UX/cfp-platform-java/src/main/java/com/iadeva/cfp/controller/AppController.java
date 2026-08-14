package com.iadeva.cfp.controller;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/**
 * Equivalente a app.controller.ts — health check simples na raiz da API.
 */
@RestController
public class AppController {

    @GetMapping
    public Map<String, String> getData() {
        return Map.of("message", "Hello API");
    }
}
