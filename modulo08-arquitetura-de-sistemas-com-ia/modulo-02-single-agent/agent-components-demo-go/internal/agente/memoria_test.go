package agente

import "testing"

// Cobre o mesmo cenario de testarMemoria() em agent-components-demo.js / .py:
// duas chamadas com o mesmo usuarioId acumulam estado (memoria de longo prazo).

func TestMemoriaCurtoPrazo_ConstroiDuasMensagens(t *testing.T) {
	contexto := MemoriaCurtoPrazo("Estudo fase II, público-alvo 12-17 anos.")

	if len(contexto) != 2 {
		t.Fatalf("esperava 2 mensagens, obteve %d", len(contexto))
	}
	if contexto[0].Role != "system" {
		t.Errorf("esperava role=system na primeira mensagem, obteve %q", contexto[0].Role)
	}
	if contexto[1].Role != "user" {
		t.Errorf("esperava role=user na segunda mensagem, obteve %q", contexto[1].Role)
	}
}

func TestMemoriaLongoPrazo_DuasChamadasComMesmoUsuario_AcumulamEstado(t *testing.T) {
	banco := NovaMemoriaLongoPrazo()

	primeira := banco.Registrar("usuario-teste", "prefere respostas curtas")
	segunda := banco.Registrar("usuario-teste", "prefere tom formal")

	if primeira.Interacoes != 1 {
		t.Errorf("esperava 1 interação na primeira chamada, obteve %d", primeira.Interacoes)
	}
	if segunda.Interacoes != 2 {
		t.Errorf("esperava 2 interações na segunda chamada, obteve %d", segunda.Interacoes)
	}
	if len(segunda.Preferencias) != 2 {
		t.Errorf("esperava 2 preferências acumuladas, obteve %d", len(segunda.Preferencias))
	}
}

func TestMemoriaLongoPrazo_UsuariosDiferentes_NaoCompartilhamEstado(t *testing.T) {
	banco := NovaMemoriaLongoPrazo()

	banco.Registrar("usuario-a", "prefere respostas curtas")
	usuarioB := banco.Registrar("usuario-b", "")

	if usuarioB.Interacoes != 1 {
		t.Errorf("esperava 1 interação para usuario-b, obteve %d", usuarioB.Interacoes)
	}
	if len(usuarioB.Preferencias) != 0 {
		t.Errorf("esperava 0 preferências para usuario-b, obteve %d", len(usuarioB.Preferencias))
	}
}

func TestMemoriaLongoPrazo_InstanciasDiferentes_NaoCompartilhamEstado(t *testing.T) {
	primeiroBanco := NovaMemoriaLongoPrazo()
	primeiroBanco.Registrar("usuario-teste", "prefere respostas curtas")

	segundoBanco := NovaMemoriaLongoPrazo()
	resultado := segundoBanco.Registrar("usuario-teste", "")

	if resultado.Interacoes != 1 {
		t.Errorf("esperava banco novo isolado com 1 interação, obteve %d", resultado.Interacoes)
	}
}
