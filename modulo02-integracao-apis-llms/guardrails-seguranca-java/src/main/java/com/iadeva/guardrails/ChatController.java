package com.iadeva.guardrails;

import com.iadeva.guardrails.graph.SafeguardOrchestrator;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Size;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
public class ChatController {

    private final SafeguardOrchestrator orchestrator;

    public ChatController(SafeguardOrchestrator orchestrator) {
        this.orchestrator = orchestrator;
    }

    @PostMapping("/chat")
    public ResponseEntity<ChatResponse> chat(@Valid @RequestBody ChatRequest request) {
        var result = orchestrator.process(
                request.username() != null ? request.username() : "member",
                request.message()
        );
        return ResponseEntity.ok(new ChatResponse(result.allowed(), result.message()));
    }

    public record ChatRequest(
            String username,
            @Size(min = 3) String message
    ) {}

    public record ChatResponse(boolean allowed, String message) {}
}
