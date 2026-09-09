package com.opspilot.llm;

import java.util.Map;

public record ToolCall(String toolName, Map<String, Object> args) {
}
