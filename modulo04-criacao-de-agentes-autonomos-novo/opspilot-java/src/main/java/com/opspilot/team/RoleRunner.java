package com.opspilot.team;

import com.opspilot.tools.Tool;

import java.util.List;

/** Porte da interface RoleRunner (team/roles.ts). */
public interface RoleRunner {
    TeamRole role();

    List<Tool> tools();

    RoleRunResult run(RoleRunInput input);
}
