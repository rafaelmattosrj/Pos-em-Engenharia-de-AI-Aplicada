package com.opspilot.domain;

/** Porte simplificado de TraceEvent (domain/types.ts) -- so os campos usados pelo nucleo portado. */
public record TraceEvent(Type type, String node, String content, String to) {
    public enum Type { THOUGHT, ACTION, OBSERVATION, PLAN, ANSWER, HANDOFF }

    public static TraceEvent handoff(String node, String to, String content) {
        return new TraceEvent(Type.HANDOFF, node, content, to);
    }

    public static TraceEvent answer(String node, String content) {
        return new TraceEvent(Type.ANSWER, node, content, null);
    }

    public static TraceEvent plan(String node, String content) {
        return new TraceEvent(Type.PLAN, node, content, null);
    }

    public static TraceEvent observation(String node, String content) {
        return new TraceEvent(Type.OBSERVATION, node, content, null);
    }

    public static TraceEvent action(String node, String content) {
        return new TraceEvent(Type.ACTION, node, content, null);
    }

    public static TraceEvent thought(String node, String content) {
        return new TraceEvent(Type.THOUGHT, node, content, null);
    }
}
