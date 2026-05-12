package com.iadeva.analyzer.node;

import com.iadeva.analyzer.model.AnalysisResult;
import com.iadeva.analyzer.model.IntentResult;
import com.iadeva.analyzer.tool.CsvToJsonTool;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

// Equivalente ao ExecutorNode do LangGraph — executa as ações com base na intenção extraída
@Component
public class ExecutorNode {

    private final ChatClient chatClient;
    private final CsvToJsonTool csvToJsonTool;

    /**
     * Constrói o ExecutorNode com ChatClient configurado com todas as tools disponíveis.
     * As MCP tools são injetadas automaticamente pelo Spring AI via ToolCallbackProvider
     * quando um servidor MCP está configurado no application.properties.
     *
     * @param chatClientBuilder Builder do ChatClient (auto-configurado pelo Spring AI)
     * @param csvToJsonTool     Tool local para conversão de CSV
     * @param mcpToolProviders  Providers de tools MCP (injetados se MCP estiver configurado)
     */
    public ExecutorNode(
            ChatClient.Builder chatClientBuilder,
            CsvToJsonTool csvToJsonTool,
            List<ToolCallbackProvider> mcpToolProviders
    ) {
        this.csvToJsonTool = csvToJsonTool;

        // Registra a tool local CsvToJsonTool no ChatClient
        // As MCP tools são registradas automaticamente pelo Spring AI quando disponíveis
        ChatClient.Builder builder = chatClientBuilder
                .defaultTools(csvToJsonTool);

        // Adiciona tools MCP se houver servidores configurados
        if (mcpToolProviders != null && !mcpToolProviders.isEmpty()) {
            for (ToolCallbackProvider provider : mcpToolProviders) {
                builder = builder.defaultToolCallbacks(provider);
            }
        }

        this.chatClient = builder.build();
    }

    /**
     * Executa a análise com base na intenção extraída pelo IntentNode.
     * Usa ChatClient com function calling (tools locais + MCP) para processar
     * os dados e responder à pergunta do usuário.
     *
     * @param intentResult Resultado do IntentNode com intenção e dados normalizados
     * @return AnalysisResult com relatório, tools usadas e passos de processamento
     */
    public AnalysisResult execute(IntentResult intentResult) {
        List<String> processingSteps = new ArrayList<>();
        processingSteps.add("IntentNode concluído: tipo=" + intentResult.dataType()
                + ", tools sugeridas=" + intentResult.suggestedTools());

        // Prompt de execução — instrui o LLM a usar as tools disponíveis
        String prompt = """
                Você é um analista de vendas especializado. Analise os dados abaixo e responda à pergunta.
                
                Use as tools disponíveis quando necessário:
                - csvToJson: para converter CSV em JSON antes de analisar
                - Qualquer tool MCP disponível para análises adicionais
                
                Tipo dos dados: %s
                
                Pergunta: %s
                
                Dados:
                %s
                
                Forneça uma análise detalhada e objetiva. Se os dados estiverem em CSV, converta-os primeiro.
                """.formatted(
                intentResult.dataType(),
                intentResult.question(),
                intentResult.parsedData()
        );

        processingSteps.add("ExecutorNode iniciado: enviando prompt ao LLM com tools habilitadas");

        String report = chatClient.prompt()
                .user(prompt)
                .call()
                .content();

        processingSteps.add("ExecutorNode concluído: relatório gerado com sucesso");

        // Determina tools usadas com base nas tools sugeridas e no tipo de dado
        List<String> toolsUsed = new ArrayList<>(intentResult.suggestedTools());
        if ("csv".equals(intentResult.dataType()) && !toolsUsed.contains("csvToJson")) {
            toolsUsed.add("csvToJson");
        }

        return new AnalysisResult(report, toolsUsed, processingSteps);
    }
}
