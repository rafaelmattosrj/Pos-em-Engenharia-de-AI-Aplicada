package com.iadeva.guardrails.model;

import java.util.List;

public record User(String username, String role, List<String> permissions, String displayName) {}
