package com.iadeva.roteamento.graph;

import java.util.List;

// Estado imutável que flui pelo grafo — equivalente ao GraphState do LangGraph
public record GraphState(
        List<String> messages,
        String output,
        Command command
) {
    public enum Command { UPPERCASE, LOWERCASE, UNKNOWN }

    public GraphState withOutput(String newOutput) {
        return new GraphState(messages, newOutput, command);
    }

    public GraphState withCommand(Command newCommand) {
        return new GraphState(messages, output, newCommand);
    }

    public static GraphState initial(String input) {
        return new GraphState(List.of(input), input, Command.UNKNOWN);
    }
}
