package com.opspilot.strategies;

import com.opspilot.domain.ExecutionMetrics;
import com.opspilot.domain.ReasoningStrategy;
import com.opspilot.domain.StrategyRunInput;
import com.opspilot.domain.StrategyResult;
import com.opspilot.domain.TraceEvent;
import com.opspilot.llm.ChatMessage;
import com.opspilot.llm.ChatModel;
import com.opspilot.llm.ModelResponse;
import com.opspilot.llm.ToolCall;
import com.opspilot.tools.Tool;

import java.util.ArrayList;
import java.util.List;

/**
 * Porte simplificado de agents/react.ts / strategies/react.ts: loop
 * observacao->acao unico agente, usado como estrategia de referencia/baseline
 * ao lado do TeamStrategy multiagente (o foco desta unidade).
 */
public class ReactStrategy implements ReasoningStrategy {

    public static final String SYSTEM_PROMPT =
            "Voce e o OpsPilot, copiloto de plantao de incidentes. Use as ferramentas disponiveis "
                    + "para diagnosticar e agir; responda com a resposta final quando tiver o suficiente.";

    private final ChatModel model;
    private final List<Tool> tools;
    private final int maxIterations;

    public ReactStrategy(ChatModel model, List<Tool> tools) {
        this(model, tools, 10);
    }

    public ReactStrategy(ChatModel model, List<Tool> tools, int maxIterations) {
        this.model = model;
        this.tools = tools;
        this.maxIterations = maxIterations;
    }

    @Override
    public String name() {
        return "react";
    }

    @Override
    public StrategyResult run(StrategyRunInput input) {
        long startedAt = System.currentTimeMillis();
        List<ChatMessage> messages = new ArrayList<>(List.of(
                ChatMessage.system(SYSTEM_PROMPT),
                ChatMessage.user(input.message())));
        List<TraceEvent> trace = new ArrayList<>();
        int llmCalls = 0;
        String answer = "";

        for (int iteration = 0; iteration < maxIterations; iteration++) {
            ModelResponse response = model.invoke(messages, tools);
            llmCalls++;

            if (response.toolCalls().isEmpty()) {
                answer = response.content();
                trace.add(TraceEvent.answer("react", answer));
                break;
            }

            for (ToolCall call : response.toolCalls()) {
                trace.add(TraceEvent.action("react", call.toolName() + " " + call.args()));
                String observation = executeTool(call);
                trace.add(TraceEvent.observation("react", observation));
                messages.add(new ChatMessage(ChatMessage.Role.TOOL, observation));
            }
        }

        return new StrategyResult(
                answer, trace, new ExecutionMetrics(llmCalls, System.currentTimeMillis() - startedAt));
    }

    private String executeTool(ToolCall call) {
        for (Tool tool : tools) {
            if (tool.name().equals(call.toolName())) {
                return tool.execute(call.args());
            }
        }
        return "Erro: ferramenta \"" + call.toolName() + "\" nao encontrada.";
    }
}
