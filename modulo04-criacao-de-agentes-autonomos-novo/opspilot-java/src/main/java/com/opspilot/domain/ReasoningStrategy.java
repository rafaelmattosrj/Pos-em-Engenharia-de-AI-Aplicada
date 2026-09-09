package com.opspilot.domain;

/** Porte da interface ReasoningStrategy (domain/types.ts). */
public interface ReasoningStrategy {
    String name();

    StrategyResult run(StrategyRunInput input);
}
