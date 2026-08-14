package com.example;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class PromptBuilderTest {

    @Test
    void build_incluiPerguntaEContextoNoPrompt() {
        String prompt = PromptBuilder.build("O que é um tensor?", "Um tensor é uma estrutura de dados.");

        assertThat(prompt)
                .contains("O que é um tensor?")
                .contains("Um tensor é uma estrutura de dados.")
                .contains("TensorFlow.js")
                .contains("**Pergunta do usuário:**")
                .contains("**Contexto recuperado do documento:**");
    }

    @Test
    void build_mantemInstrucoesFixasDoTemplate() {
        String prompt = PromptBuilder.build("pergunta", "contexto");

        assertThat(prompt).contains(
                "1. Use APENAS as informações do contexto fornecido para responder",
                "8. Use analogias e exemplos práticos para facilitar o entendimento"
        );
    }
}
