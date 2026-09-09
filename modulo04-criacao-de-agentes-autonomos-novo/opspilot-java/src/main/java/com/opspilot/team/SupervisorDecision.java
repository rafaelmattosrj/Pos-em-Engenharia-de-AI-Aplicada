package com.opspilot.team;

/** Porte de SupervisorDecision (team/supervisor.ts). next == null significa "done". */
public record SupervisorDecision(TeamRole next, String brief) {
    public boolean isDone() {
        return next == null;
    }

    public static SupervisorDecision done(String brief) {
        return new SupervisorDecision(null, brief);
    }
}
