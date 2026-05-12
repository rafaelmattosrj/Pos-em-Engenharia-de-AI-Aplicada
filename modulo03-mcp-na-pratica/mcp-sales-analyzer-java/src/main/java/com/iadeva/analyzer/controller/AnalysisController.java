package com.iadeva.analyzer.controller;

import com.iadeva.analyzer.graph.AnalysisOrchestrator;
import com.iadeva.analyzer.model.AnalysisRequest;
import com.iadeva.analyzer.model.AnalysisResult;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

// Endpoint REST que expõe o pipeline de análise de vendas
@RestController
@RequestMapping("/analyze")
public class AnalysisController {

    private final AnalysisOrchestrator orchestrator;

    public AnalysisController(AnalysisOrchestrator orchestrator) {
        this.orchestrator = orchestrator;
    }

    /**
     * Analisa dados de vendas (CSV ou JSON) respondendo à pergunta do usuário.
     *
     * Exemplo de requisição:
     * POST /analyze
     * {
     *   "question": "Qual produto teve maior receita?",
     *   "data": "produto,quantidade,preco\nNotebook,10,2500.00\nMouse,50,45.00"
     * }
     *
     * @param request Corpo da requisição com question e data
     * @return AnalysisResult com relatório e metadados de execução
     */
    @PostMapping
    public ResponseEntity<AnalysisResult> analyze(@RequestBody AnalysisRequest request) {
        AnalysisResult result = orchestrator.analyze(request);
        return ResponseEntity.ok(result);
    }
}
