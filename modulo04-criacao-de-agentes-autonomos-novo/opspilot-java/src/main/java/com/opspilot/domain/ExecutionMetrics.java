package com.opspilot.domain;

public record ExecutionMetrics(int llmCalls, long latencyMs) {
}
