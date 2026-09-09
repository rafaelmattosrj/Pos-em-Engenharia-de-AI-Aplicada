package com.opspilot;

import com.opspilot.llm.FakeChatModel;
import com.opspilot.llm.ModelResponse;
import com.opspilot.llm.ToolCall;
import com.opspilot.team.BlackboardEntry;
import com.opspilot.team.RoleRunInput;
import com.opspilot.team.RoleRunResult;
import com.opspilot.team.RoleRunner;
import com.opspilot.team.RoleRunners;
import com.opspilot.tools.Tool;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class RoleRunnersTest {

    private static Tool echoTool(String name) {
        return new Tool() {
            public String name() {
                return name;
            }

            public String description() {
                return "tool de teste";
            }

            public String execute(Map<String, Object> args) {
                return "resultado de " + name + " " + args;
            }
        };
    }

    @Test
    void planejadorMakesSingleDirectCall() {
        FakeChatModel model = new FakeChatModel(List.of(ModelResponse.text("1. Fazer X\n2. Fazer Y")));
        RoleRunner planejador = RoleRunners.planejador(model);

        RoleRunResult result = planejador.run(new RoleRunInput("mensagem", "planeje", List.of()));

        assertThat(result.llmCalls()).isEqualTo(1);
        assertThat(result.entry().kind()).isEqualTo(BlackboardEntry.Kind.PLAN);
        assertThat(result.entry().content()).contains("Fazer X");
        assertThat(model.callCount()).isEqualTo(1);
    }

    @Test
    void analistaUsesToolThenAnswers() {
        Tool listIncidents = echoTool("list_incidents");
        FakeChatModel model = new FakeChatModel(List.of(
                new ModelResponse("", List.of(new ToolCall("list_incidents", Map.of()))),
                ModelResponse.text("2 incidentes abertos, nenhum critico.")));

        RoleRunner analista = RoleRunners.analista(model, List.of(listIncidents));
        RoleRunResult result = analista.run(new RoleRunInput("o que esta pegando?", "diagnostique", List.of()));

        assertThat(result.llmCalls()).isEqualTo(2);
        assertThat(result.entry().kind()).isEqualTo(BlackboardEntry.Kind.FACTS);
        assertThat(result.entry().content()).contains("nenhum critico");
    }

    @Test
    void executorFallsBackToErrorMessageForUnknownTool() {
        FakeChatModel model = new FakeChatModel(List.of(
                new ModelResponse("", List.of(new ToolCall("tool_inexistente", Map.of()))),
                ModelResponse.text("Nao consegui executar a acao.")));

        RoleRunner executor = RoleRunners.executor(model, List.of());
        RoleRunResult result = executor.run(new RoleRunInput("resolva o incidente", "execute", List.of()));

        assertThat(result.trace().stream().anyMatch(e -> e.content().contains("nao disponivel"))).isTrue();
    }
}
