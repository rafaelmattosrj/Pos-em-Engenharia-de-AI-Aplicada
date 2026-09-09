package com.opspilot.llm;

import java.util.List;

public record ModelResponse(String content, List<ToolCall> toolCalls) {
    public static ModelResponse text(String content) {
        return new ModelResponse(content, List.of());
    }
}
