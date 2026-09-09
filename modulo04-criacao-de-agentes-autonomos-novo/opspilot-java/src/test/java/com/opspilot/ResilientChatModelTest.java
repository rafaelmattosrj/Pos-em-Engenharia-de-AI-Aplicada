package com.opspilot;

import com.opspilot.domain.Exceptions;
import com.opspilot.llm.ChatMessage;
import com.opspilot.llm.ChatModel;
import com.opspilot.llm.ModelResponse;
import com.opspilot.llm.ResilientChatModel;
import com.opspilot.tools.Tool;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class ResilientChatModelTest {

    private static ChatModel alwaysFails(AtomicInteger counter) {
        return (messages, tools) -> {
            counter.incrementAndGet();
            throw new RuntimeException("provider indisponivel");
        };
    }

    private static ChatModel alwaysSucceeds(String answer) {
        return (messages, tools) -> ModelResponse.text(answer);
    }

    @Test
    void succeedsOnPrimaryWithoutTouchingBackup() {
        AtomicInteger backupCalls = new AtomicInteger();
        ChatModel primary = alwaysSucceeds("ok");
        ChatModel backup = (messages, tools) -> {
            backupCalls.incrementAndGet();
            return ModelResponse.text("nao deveria chegar aqui");
        };

        ResilientChatModel resilient = new ResilientChatModel(primary, backup);
        ModelResponse response = resilient.invoke(List.of(ChatMessage.user("oi")), List.of());

        assertThat(response.content()).isEqualTo("ok");
        assertThat(backupCalls.get()).isZero();
    }

    @Test
    void fallsBackToBackupWhenPrimaryExhaustsRetries() {
        AtomicInteger primaryCalls = new AtomicInteger();
        ChatModel primary = alwaysFails(primaryCalls);
        ChatModel backup = alwaysSucceeds("resposta do backup");

        ResilientChatModel resilient = new ResilientChatModel(primary, backup);
        ModelResponse response = resilient.invoke(List.of(ChatMessage.user("oi")), List.of());

        assertThat(response.content()).isEqualTo("resposta do backup");
        assertThat(primaryCalls.get()).isEqualTo(ResilientChatModel.RETRY_ATTEMPTS);
    }

    @Test
    void throwsModelUnavailableWhenNoBackupConfigured() {
        ResilientChatModel resilient = new ResilientChatModel(alwaysFails(new AtomicInteger()), null);

        assertThatThrownBy(() -> resilient.invoke(List.of(ChatMessage.user("oi")), List.<Tool>of()))
                .isInstanceOf(Exceptions.ModelUnavailableException.class);
    }

    @Test
    void throwsModelUnavailableWhenBothPrimaryAndBackupFail() {
        ResilientChatModel resilient = new ResilientChatModel(
                alwaysFails(new AtomicInteger()), alwaysFails(new AtomicInteger()));

        assertThatThrownBy(() -> resilient.invoke(List.of(ChatMessage.user("oi")), List.<Tool>of()))
                .isInstanceOf(Exceptions.ModelUnavailableException.class);
    }
}
