package com.opspilot.team;

/** Porte de BlackboardEntry (team/blackboard.ts). */
public record BlackboardEntry(TeamRole role, Kind kind, String brief, String content) {
    public enum Kind { FACTS, PLAN, EXECUTION, ERROR }
}
