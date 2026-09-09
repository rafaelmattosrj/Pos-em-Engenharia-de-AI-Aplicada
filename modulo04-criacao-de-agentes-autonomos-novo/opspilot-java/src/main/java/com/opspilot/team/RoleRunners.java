package com.opspilot.team;

import com.opspilot.domain.TraceEvent;
import com.opspilot.llm.ChatMessage;
import com.opspilot.llm.ChatModel;
import com.opspilot.llm.ModelResponse;
import com.opspilot.llm.ToolCall;
import com.opspilot.tools.Tool;

import java.util.ArrayList;
import java.util.List;

/** Porte de agents/roles.ts -- os 3 papeis restritos (analista/planejador/executor). */
public final class RoleRunners {

    public static final String ANALISTA_SYSTEM_PROMPT = String.join("\n",
            "Voce e o ANALISTA do plantao. Sua unica funcao: produzir diagnostico FACTUAL do",
            "estado atual, usando as ferramentas de leitura. NAO proponha solucoes. NAO abra",
            "nem resolva nada.");

    public static final String PLANEJADOR_SYSTEM_PROMPT = String.join("\n",
            "Voce e o PLANEJADOR do plantao. Transforme os fatos do blackboard em um plano",
            "numerado e executavel. Voce nao tem ferramentas: nao invente dados que nao estejam",
            "no blackboard.");

    public static final String EXECUTOR_SYSTEM_PROMPT = String.join("\n",
            "Voce e o EXECUTOR do plantao. Execute acoes de incidente (abrir, resolver, listar)",
            "conforme o brief e o plano do blackboard, usando somente as ferramentas disponiveis.");

    private static final int DEFAULT_MAX_ITERATIONS = 6;

    private RoleRunners() {
    }

    public static RoleRunner analista(ChatModel model, List<Tool> tools) {
        return createToolRoleRunner(TeamRole.ANALISTA, BlackboardEntry.Kind.FACTS, ANALISTA_SYSTEM_PROMPT, model, tools);
    }

    public static RoleRunner executor(ChatModel model, List<Tool> tools) {
        return createToolRoleRunner(TeamRole.EXECUTOR, BlackboardEntry.Kind.EXECUTION, EXECUTOR_SYSTEM_PROMPT, model, tools);
    }

    /** Planejador tem zero ferramentas por assinatura: so uma chamada de modelo sobre o blackboard. */
    public static RoleRunner planejador(ChatModel model) {
        return new RoleRunner() {
            public TeamRole role() {
                return TeamRole.PLANEJADOR;
            }

            public List<Tool> tools() {
                return List.of();
            }

            public RoleRunResult run(RoleRunInput input) {
                ModelResponse response = model.invoke(List.of(
                        ChatMessage.system(PLANEJADOR_SYSTEM_PROMPT),
                        ChatMessage.user(roleUserMessage(input))));
                BlackboardEntry entry = new BlackboardEntry(
                        TeamRole.PLANEJADOR, BlackboardEntry.Kind.PLAN, input.brief(), response.content());
                return new RoleRunResult(entry, List.of(TraceEvent.plan("planejador", response.content())), 1);
            }
        };
    }

    private static RoleRunner createToolRoleRunner(
            TeamRole role, BlackboardEntry.Kind kind, String prompt, ChatModel model, List<Tool> tools) {
        return new RoleRunner() {
            public TeamRole role() {
                return role;
            }

            public List<Tool> tools() {
                return tools;
            }

            public RoleRunResult run(RoleRunInput input) {
                List<ChatMessage> messages = new ArrayList<>(List.of(
                        ChatMessage.system(prompt),
                        ChatMessage.user(roleUserMessage(input))));
                List<TraceEvent> trace = new ArrayList<>();
                int llmCalls = 0;
                String finalAnswer = "";

                for (int iteration = 0; iteration < DEFAULT_MAX_ITERATIONS; iteration++) {
                    ModelResponse response = model.invoke(messages, tools);
                    llmCalls++;

                    if (response.toolCalls().isEmpty()) {
                        finalAnswer = response.content();
                        trace.add(TraceEvent.answer(role.name().toLowerCase(), finalAnswer));
                        break;
                    }

                    for (ToolCall call : response.toolCalls()) {
                        trace.add(TraceEvent.action(role.name().toLowerCase(), call.toolName() + " " + call.args()));
                        String observation = executeTool(tools, call);
                        trace.add(TraceEvent.observation(role.name().toLowerCase(), observation));
                        messages.add(new ChatMessage(ChatMessage.Role.TOOL, observation));
                    }
                }

                BlackboardEntry entry = new BlackboardEntry(role, kind, input.brief(), finalAnswer);
                return new RoleRunResult(entry, trace, llmCalls);
            }
        };
    }

    private static String executeTool(List<Tool> tools, ToolCall call) {
        for (Tool tool : tools) {
            if (tool.name().equals(call.toolName())) {
                return tool.execute(call.args());
            }
        }
        return "Erro: ferramenta \"" + call.toolName() + "\" nao disponivel para este papel.";
    }

    private static String roleUserMessage(RoleRunInput input) {
        return String.join("\n",
                "Pedido original do plantonista: " + input.message(),
                "Sua tarefa (brief do supervisor): " + input.brief(),
                "",
                "Blackboard atual:",
                Blackboard.render(input.blackboard()));
    }
}
