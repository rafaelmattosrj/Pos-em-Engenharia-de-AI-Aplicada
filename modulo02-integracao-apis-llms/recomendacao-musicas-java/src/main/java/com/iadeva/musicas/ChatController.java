package com.iadeva.musicas;

import com.iadeva.musicas.graph.MusicChatOrchestrator;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Size;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
public class ChatController {

    private final MusicChatOrchestrator orchestrator;

    public ChatController(MusicChatOrchestrator orchestrator) {
        this.orchestrator = orchestrator;
    }

    @PostMapping("/chat")
    public ResponseEntity<ChatResponse> chat(@Valid @RequestBody ChatRequest request) {
        String reply = orchestrator.chat(
                request.userId() != null ? request.userId() : "default-user",
                request.sessionId() != null ? request.sessionId() : "default-session",
                request.message()
        );
        return ResponseEntity.ok(new ChatResponse(reply));
    }

    public record ChatRequest(
            String userId,
            String sessionId,
            @Size(min = 3) String message
    ) {}

    public record ChatResponse(String reply) {}
}
