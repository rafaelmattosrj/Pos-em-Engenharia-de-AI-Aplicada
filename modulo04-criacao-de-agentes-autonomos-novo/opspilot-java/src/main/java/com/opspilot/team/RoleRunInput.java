package com.opspilot.team;

import java.util.List;

public record RoleRunInput(String message, String brief, List<BlackboardEntry> blackboard) {
}
