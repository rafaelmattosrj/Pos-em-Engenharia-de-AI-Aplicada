package com.opspilot.team;

import com.opspilot.domain.TraceEvent;

import java.util.List;

public record RoleRunResult(BlackboardEntry entry, List<TraceEvent> trace, int llmCalls) {
}
