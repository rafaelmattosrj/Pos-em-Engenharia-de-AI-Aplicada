package com.opspilot.team;

import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.IntStream;

/** Porte de renderBlackboard (team/blackboard.ts). */
public final class Blackboard {

    private Blackboard() {
    }

    public static String render(List<BlackboardEntry> entries) {
        if (entries.isEmpty()) {
            return "(blackboard vazio -- nenhuma contribuicao ainda)";
        }
        return IntStream.range(0, entries.size())
                .mapToObj(i -> {
                    BlackboardEntry entry = entries.get(i);
                    return "[%d] %s (%s) -- brief: %s\n%s".formatted(
                            i + 1, entry.role(), entry.kind(), entry.brief(), entry.content());
                })
                .collect(Collectors.joining("\n\n"));
    }
}
