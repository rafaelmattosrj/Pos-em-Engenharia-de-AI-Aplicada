package com.iadeva.neo4j;

import com.iadeva.neo4j.graph.RagOrchestrator;
import com.iadeva.neo4j.service.Neo4jService;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Size;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
public class ChatController {

    private final RagOrchestrator orchestrator;
    private final Neo4jService neo4jService;

    public ChatController(RagOrchestrator orchestrator, Neo4jService neo4jService) {
        this.orchestrator = orchestrator;
        this.neo4jService = neo4jService;
    }

    @PostMapping("/chat")
    public ResponseEntity<ChatResponse> chat(@Valid @RequestBody ChatRequest request) {
        String answer = orchestrator.query(request.question());
        return ResponseEntity.ok(new ChatResponse(answer));
    }

    // Endpoint para popular o banco com dados de exemplo
    @PostMapping("/seed")
    public ResponseEntity<String> seed() {
        neo4jService.seedData();
        return ResponseEntity.ok("Dados inseridos com sucesso");
    }

    public record ChatRequest(@Size(min = 5) String question) {}
    public record ChatResponse(String answer) {}
}
