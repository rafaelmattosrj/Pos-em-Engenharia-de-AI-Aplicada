package agente

import "testing"

// Cobre o mesmo cenario de testarGate() em agent-components-demo.js / .py:
// acao sem gate executa direto; acao com gate nunca chega a executar.

func TestExecutarOuGatear_AcaoSemGate_ExecutaDireto(t *testing.T) {
	resultado := ExecutarOuGatear(AcaoProposta{
		Tipo: "x", RequerAprovacao: false, Executar: func() string { return "feito" },
	})

	if resultado.Status != "executada" {
		t.Errorf("esperava status=executada, obteve %q", resultado.Status)
	}
	if resultado.Resultado != "feito" {
		t.Errorf("esperava resultado=feito, obteve %q", resultado.Resultado)
	}
	if resultado.Mensagem != "" {
		t.Errorf("não esperava mensagem, obteve %q", resultado.Mensagem)
	}
}

func TestExecutarOuGatear_AcaoComGate_NuncaChegaAExecutar(t *testing.T) {
	executou := false
	resultado := ExecutarOuGatear(AcaoProposta{
		Tipo:            "y",
		RequerAprovacao: true,
		Executar: func() string {
			executou = true
			return "nunca deveria rodar"
		},
	})

	if resultado.Status != "aguardando_aprovacao" {
		t.Errorf("esperava status=aguardando_aprovacao, obteve %q", resultado.Status)
	}
	if resultado.Resultado != "" {
		t.Errorf("não esperava resultado, obteve %q", resultado.Resultado)
	}
	if executou {
		t.Error("Executar não deveria ter rodado quando RequerAprovacao=true")
	}
}
