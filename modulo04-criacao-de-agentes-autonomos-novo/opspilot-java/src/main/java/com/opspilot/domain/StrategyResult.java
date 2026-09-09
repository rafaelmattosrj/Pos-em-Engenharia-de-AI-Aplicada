package com.opspilot.domain;

import java.util.List;

public record StrategyResult(String answer, List<TraceEvent> trace, ExecutionMetrics metrics) {
}
