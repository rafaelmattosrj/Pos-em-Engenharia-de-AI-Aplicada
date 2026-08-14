package com.iadeva.bragbot.controller;

import com.iadeva.bragbot.model.BragDocument;
import com.iadeva.bragbot.model.BragRequest;
import com.iadeva.bragbot.service.BragService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/**
 * Equivalente à rota POST /api/brag de server.ts.
 */
@RestController
public class BragController {

    private final BragService bragService;

    public BragController(BragService bragService) {
        this.bragService = bragService;
    }

    @PostMapping("/api/brag")
    public ResponseEntity<?> generate(@RequestBody BragRequest request) {
        if (request.definition() == null || request.definition().isBlank()) {
            return ResponseEntity.badRequest().body(Map.of("error", "Definition is required"));
        }

        try {
            BragDocument document = bragService.generate(request.definition());
            return ResponseEntity.ok(document);
        } catch (Exception e) {
            return ResponseEntity.internalServerError().body(Map.of("error", "Failed to generate brag"));
        }
    }
}
