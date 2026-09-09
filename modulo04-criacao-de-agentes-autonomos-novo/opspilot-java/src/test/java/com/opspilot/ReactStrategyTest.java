package com.opspilot;

import com.opspilot.domain.StrategyRunInput;
import com.opspilot.domain.StrategyResult;
import com.opspilot.llm.FakeChatModel;
import com.opspilot.llm.ModelResponse;
import com.opspilot.llm.ToolCall;
import com.opspilot.strategies.ReactStrategy;
import com.opspilot.tools.Tool;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class ReactStrategyTest {

    @Test
    void answersDirectlyWithoutToolsWhenModelHasNoToolCall() {
        FakeChatModel model = new FakeChatModel(List.of(ModelResponse.text("Tudo operacional.")));
        ReactStrategy strategy = new ReactStrategy(model, List.of());

        StrategyResult result = strategy.run(StrategyRunInput.of("status geral?"));

        assertThat(result.answer()).isEqualTo("Tudo operacional.");
        assertThat(result.metrics().llmCalls()).isEqualTo(1);
    }

    @Test
    void usesToolBeforeFinalAnswer() {
        Tool checkStatus = new Tool() {
            public String name() {
                return "check_provider_status";
            }

            public String description() {
                return "";
            }

            public String execute(Map<String, Object> args) {
                return "operacional";
            }
        };
        FakeChatModel model = new FakeChatModel(List.of(
                new ModelResponse("", List.of(new ToolCall("check_provider_status", Map.of()))),
                ModelResponse.text("Provedores operacionais.")));

        ReactStrategy strategy = new ReactStrategy(model, List.of(checkStatus));
        StrategyResult result = strategy.run(StrategyRunInput.of("provedores estao ok?"));

        assertThat(result.answer()).isEqualTo("Provedores operacionais.");
        assertThat(result.metrics().llmCalls()).isEqualTo(2);
        assertThat(result.trace()).anyMatch(event -> event.content().contains("operacional"));
    }

    @Test
    void stopsAtMaxIterationsIfModelNeverAnswers() {
        Tool loopTool = new Tool() {
            public String name() {
                return "loop";
            }

            public String description() {
                return "";
            }

            public String execute(Map<String, Object> args) {
                return "de novo";
            }
        };
        FakeChatModel model = new FakeChatModel(messages ->
                new ModelResponse("", List.of(new ToolCall("loop", Map.of()))));

        ReactStrategy strategy = new ReactStrategy(model, List.of(loopTool), 3);
        StrategyResult result = strategy.run(StrategyRunInput.of("nunca termina"));

        assertThat(result.answer()).isEmpty();
        assertThat(result.metrics().llmCalls()).isEqualTo(3);
    }
}
