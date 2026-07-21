package com.iadeva.agendamento;

import com.iadeva.agendamento.graph.AppointmentOrchestrator;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Size;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
public class ChatController {

    private final AppointmentOrchestrator orchestrator;

    public ChatController(AppointmentOrchestrator orchestrator) {
        this.orchestrator = orchestrator;
    }

    @PostMapping("/chat")
    public ResponseEntity<ChatResponse> chat(@Valid @RequestBody ChatRequest request) {
        String reply = orchestrator.process(request.question());
        return ResponseEntity.ok(new ChatResponse(reply));
    }

    public record ChatRequest(@Size(min = 5) String question) {}
    public record ChatResponse(String reply) {}
}
