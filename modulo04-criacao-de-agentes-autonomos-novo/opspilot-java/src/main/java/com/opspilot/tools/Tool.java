package com.opspilot.tools;

import java.util.Map;

/** Porte do conceito de DynamicStructuredTool (agents/tools.ts), simplificado. */
public interface Tool {
    String name();

    String description();

    String execute(Map<String, Object> args);
}
