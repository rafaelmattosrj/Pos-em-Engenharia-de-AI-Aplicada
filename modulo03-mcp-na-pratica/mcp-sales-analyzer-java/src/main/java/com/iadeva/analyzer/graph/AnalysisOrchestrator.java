package com.iadeva.analyzer.graph;

import com.iadeva.analyzer.model.AnalysisRequest;
import com.iadeva.analyzer.model.AnalysisResult;
import com.iadeva.analyzer.model.IntentResult;
import com.iadeva.analyzer.node.ExecutorNode;
import com.iadeva.analyzer.node.IntentNode;
import org.springframework.stereotype.Component;

// Equivalente ao grafo LangGraph do 09-using-mcp-with-langchain — orquestra os nós de processamento
@Component
public class AnalysisOrchestrator {

    private final IntentNode intentNode;
    private final ExecutorNode executorNode;

    public AnalysisOrchestrator(IntentNode intentNode, ExecutorNode executorNode) {
        this.intentNode = intentNode;
        this.executorNode = executorNode;
    }

    /**
     * Orquestra o pipeline de análise: IntentNode → ExecutorNode.
     *
     * Fluxo:
     * 1. IntentNode: extrai intenção estruturada (tipo de dado, tools necessárias)
     * 2. ExecutorNode: processa com LLM + tools e gera relatório
     *
     * @param request Requisição com pergunta e dados brutos
     * @return AnalysisResult com relatório final e metadados de execução
     */
    public AnalysisResult analyze(AnalysisRequest request) {
        // Nó 1: extração de intenção
        IntentResult intentResult = intentNode.extract(request);

        // Nó 2: execução com tools
        return executorNode.execute(intentResult);
    }
}
