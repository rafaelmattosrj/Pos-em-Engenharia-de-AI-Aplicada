package com.trialforge.agentcomponents;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre o mesmo cenario de testarMemoria() em agent-components-demo.js / .py:
 * duas chamadas com o mesmo usuarioId acumulam estado (memoria de longo prazo).
 */
class MemoriaTest {

    @Test
    void memoriaCurtoPrazo_construiDuasMensagens_semEstadoCompartilhado() {
        List<Memoria.Mensagem> contexto = Memoria.memoriaCurtoPrazo("Estudo fase II, público-alvo 12-17 anos.");

        assertThat(contexto).hasSize(2);
        assertThat(contexto.get(0).role()).isEqualTo("system");
        assertThat(contexto.get(1).role()).isEqualTo("user");
        assertThat(contexto.get(1).content()).contains("Estudo fase II, público-alvo 12-17 anos.");
    }

    @Test
    void memoriaLongoPrazo_duasChamadasComMesmoUsuario_acumulamEstado() {
        Memoria.MemoriaLongoPrazo banco = new Memoria.MemoriaLongoPrazo();

        Memoria.Historico primeira = banco.registrar("usuario-teste", "prefere respostas curtas");
        Memoria.Historico segunda = banco.registrar("usuario-teste", "prefere tom formal");

        assertThat(primeira.interacoes()).isEqualTo(1);
        assertThat(segunda.interacoes()).isEqualTo(2);
        assertThat(segunda.preferencias()).containsExactly("prefere respostas curtas", "prefere tom formal");
    }

    @Test
    void memoriaLongoPrazo_usuariosDiferentes_naoCompartilhamEstado() {
        Memoria.MemoriaLongoPrazo banco = new Memoria.MemoriaLongoPrazo();

        banco.registrar("usuario-a", "prefere respostas curtas");
        Memoria.Historico usuarioB = banco.registrar("usuario-b", null);

        assertThat(usuarioB.interacoes()).isEqualTo(1);
        assertThat(usuarioB.preferencias()).isEmpty();
    }

    @Test
    void memoriaLongoPrazo_instanciasDiferentes_naoCompartilhamEstado() {
        Memoria.MemoriaLongoPrazo primeiroBanco = new Memoria.MemoriaLongoPrazo();
        primeiroBanco.registrar("usuario-teste", "prefere respostas curtas");

        Memoria.MemoriaLongoPrazo segundoBanco = new Memoria.MemoriaLongoPrazo();
        Memoria.Historico resultado = segundoBanco.registrar("usuario-teste", null);

        assertThat(resultado.interacoes()).isEqualTo(1);
    }
}
