package com.opspilot.team;

import java.util.List;

/** Porte do tipo DecideNextFn (team/supervisor.ts). */
public interface DecideNextFn {
    SupervisorDecision decideNext(String message, List<BlackboardEntry> blackboard, int handoffCount);
}
