package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class TokenizerTest {

    @Test
    void removeAcentosEPontuacao() {
        assertThat(Tokenizer.tokenizar("Idade mínima: 12 anos!"))
                .containsExactly("idade", "minima", "12", "anos");
    }

    @Test
    void stringVaziaDevolveListaVazia() {
        assertThat(Tokenizer.tokenizar("")).isEmpty();
    }
}
