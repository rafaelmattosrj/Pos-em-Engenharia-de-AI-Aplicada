package trialforge

import "testing"

// Porte de criarEstadoProtocolo/revisarProtocolo (JS/Python) — o Protocolo e
// um recurso VERSIONADO, nao um valor congelado: a versao anterior fica
// preservada no historico, nunca apagada.

func TestNovoEstadoProtocoloComecaNaVersao1(t *testing.T) {
	estado := NovoEstadoProtocolo(map[string]int{"idadeMinima": 13})

	if estado.Versao() != 1 {
		t.Errorf("Versao() = %d, esperado 1", estado.Versao())
	}
	if estado.Criterios()["idadeMinima"] != 13 {
		t.Errorf("Criterios()[idadeMinima] = %d, esperado 13", estado.Criterios()["idadeMinima"])
	}
	if len(estado.Historico()) != 0 {
		t.Errorf("Historico() deveria comecar vazio, tem %d entradas", len(estado.Historico()))
	}
}

func TestRevisarIncrementaVersaoEPreservaHistoricoDaAnterior(t *testing.T) {
	estado := NovoEstadoProtocolo(map[string]int{"idadeMinima": 13})

	estado.Revisar(map[string]int{"idadeMinima": 12}, "Comitê de ética corrigiu a idade mínima.")

	if estado.Versao() != 2 {
		t.Errorf("Versao() = %d, esperado 2", estado.Versao())
	}
	if estado.Criterios()["idadeMinima"] != 12 {
		t.Errorf("Criterios()[idadeMinima] = %d, esperado 12", estado.Criterios()["idadeMinima"])
	}

	historico := estado.Historico()
	if len(historico) != 1 {
		t.Fatalf("Historico() tem %d entradas, esperado 1", len(historico))
	}
	if historico[0].VersaoAnterior != 1 {
		t.Errorf("VersaoAnterior = %d, esperado 1", historico[0].VersaoAnterior)
	}
	if historico[0].CriteriosAnteriores["idadeMinima"] != 13 {
		t.Errorf("CriteriosAnteriores[idadeMinima] = %d, esperado 13", historico[0].CriteriosAnteriores["idadeMinima"])
	}
}
