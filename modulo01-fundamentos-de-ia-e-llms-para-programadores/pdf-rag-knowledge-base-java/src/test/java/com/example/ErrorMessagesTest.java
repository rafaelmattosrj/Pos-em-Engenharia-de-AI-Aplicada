package com.example;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class ErrorMessagesTest {

    @Test
    void truncate_mensagemDentroDoLimiteNaoEAlterada() {
        assertThat(ErrorMessages.truncate("erro curto", 200)).isEqualTo("erro curto");
    }

    @Test
    void truncate_mensagemAcimaDoLimiteECortada() {
        String mensagem = "a".repeat(250);

        String resultado = ErrorMessages.truncate(mensagem, 200);

        assertThat(resultado).hasSize(200);
        assertThat(resultado).isEqualTo("a".repeat(200));
    }

    @Test
    void truncate_mensagemNulaRetornaStringVazia() {
        assertThat(ErrorMessages.truncate(null, 200)).isEmpty();
    }

    @Test
    void truncate_mensagemComTamanhoExatoNaoEAlterada() {
        String mensagem = "b".repeat(200);

        assertThat(ErrorMessages.truncate(mensagem, 200)).isEqualTo(mensagem);
    }
}
