package com.iadeva.roteamento;

import com.iadeva.roteamento.graph.WorkflowOrchestrator;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Size;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
public class ChatController {

    private final WorkflowOrchestrator orchestrator;

    public ChatController(WorkflowOrchestrator orchestrator) {
        this.orchestrator = orchestrator;
    }

    @PostMapping("/chat")
    public ResponseEntity<String> chat(@Valid @RequestBody ChatRequest request) {
        String output = orchestrator.invoke(request.question()).output();
        return ResponseEntity.ok(output);
    }

    public record ChatRequest(@Size(min = 5) String question) {}
}
