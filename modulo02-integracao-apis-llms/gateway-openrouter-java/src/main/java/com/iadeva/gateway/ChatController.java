package com.iadeva.gateway;

import jakarta.validation.Valid;
import jakarta.validation.constraints.Size;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
public class ChatController {

    private final OpenRouterService openRouterService;

    public ChatController(OpenRouterService openRouterService) {
        this.openRouterService = openRouterService;
    }

    @PostMapping("/chat")
    public ResponseEntity<LlmResponse> chat(@Valid @RequestBody ChatRequest request) {
        LlmResponse response = openRouterService.generate(request.question());
        return ResponseEntity.ok(response);
    }

    public record ChatRequest(@Size(min = 5) String question) {}
}
