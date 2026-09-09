package com.opspilot.domain;

import java.util.List;

public record StrategyRunInput(String message, List<ConversationMessage> history) {
    public static StrategyRunInput of(String message) {
        return new StrategyRunInput(message, List.of());
    }
}
