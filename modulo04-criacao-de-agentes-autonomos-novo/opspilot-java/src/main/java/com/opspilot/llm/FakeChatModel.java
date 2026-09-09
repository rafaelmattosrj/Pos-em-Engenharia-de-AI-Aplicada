package com.opspilot.llm;

import com.opspilot.tools.Tool;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.List;
import java.util.function.Function;

/**
 * Test double sem chamada de rede. O UNIDADE.md da unidade 9 registra que
 * "o codigo passa no typecheck e nos testes com fakes" -- este e o
 * equivalente Java desses fakes: uma fila de respostas programadas, ou uma
 * funcao arbitraria (para simular comportamento dependente do input).
 */
public class FakeChatModel implements ChatModel {

    private final Deque<ModelResponse> scriptedResponses;
    private final Function<List<ChatMessage>, ModelResponse> responder;
    private final List<List<ChatMessage>> callHistory = new ArrayList<>();

    public FakeChatModel(List<ModelResponse> scriptedResponses) {
        this.scriptedResponses = new ArrayDeque<>(scriptedResponses);
        this.responder = null;
    }

    public FakeChatModel(Function<List<ChatMessage>, ModelResponse> responder) {
        this.scriptedResponses = null;
        this.responder = responder;
    }

    @Override
    public ModelResponse invoke(List<ChatMessage> messages, List<Tool> tools) {
        callHistory.add(messages);
        if (responder != null) {
            return responder.apply(messages);
        }
        if (scriptedResponses.isEmpty()) {
            throw new IllegalStateException("FakeChatModel ran out of scripted responses");
        }
        return scriptedResponses.poll();
    }

    public int callCount() {
        return callHistory.size();
    }
}
