package decision

import "testing"

// Cobre os mesmos cenarios dos testes automatizados originais
// (decision-framework-tool.js#rodarTestes / decision_framework_tool.py
// #TestClassificarTarefa): as 4 combinacoes da arvore pura do Bloco 1.

func TestClassificarTarefa_P1VerdadeiroComQualquerP2P3EhRegraDeterministica(t *testing.T) {
	casos := []struct {
		p2, p3 bool
	}{
		{true, true},
		{true, false},
		{false, true},
		{false, false},
	}
	for _, c := range casos {
		got := ClassificarTarefa(true, c.p2, c.p3)
		if got != RegraDeterministica {
			t.Errorf("ClassificarTarefa(true, %v, %v) = %q, want %q", c.p2, c.p3, got, RegraDeterministica)
		}
	}
}

func TestClassificarTarefa_P1FalsoP2VerdadeiroComQualquerP3EhApprovalGate(t *testing.T) {
	for _, p3 := range []bool{true, false} {
		got := ClassificarTarefa(false, true, p3)
		if got != ApprovalGateObrigatorio {
			t.Errorf("ClassificarTarefa(false, true, %v) = %q, want %q", p3, got, ApprovalGateObrigatorio)
		}
	}
}

func TestClassificarTarefa_P1FalsoP2FalsoP3VerdadeiroEhAgenteAutonomo(t *testing.T) {
	got := ClassificarTarefa(false, false, true)
	if got != AgenteAutonomo {
		t.Errorf("ClassificarTarefa(false, false, true) = %q, want %q", got, AgenteAutonomo)
	}
}

func TestClassificarTarefa_P1FalsoP2FalsoP3FalsoEhRegraDeterministicaEnumeravel(t *testing.T) {
	got := ClassificarTarefa(false, false, false)
	if got != RegraEnumeravel {
		t.Errorf("ClassificarTarefa(false, false, false) = %q, want %q", got, RegraEnumeravel)
	}
}

// Bloco 2: decomposicao de tarefa hibrida (referencia TrialForge - Emenda de
// Protocolo). Mesma nota honesta do original: das quatro linhas da tabela de
// referencia do checklist, so estas duas subtarefas mapeiam de forma limpa
// para uma unica resposta p1/p2/p3. "Rotear pela criticidade" e "Regenerar
// documentos afetados" misturam regra e gate condicional de um jeito mais
// sutil que a arvore pura de tres perguntas nao representa sozinha, por isso
// ficam de fora do teste automatizado. Limitacao real da arvore, nao bug
// deste porte.
func subtarefasReferencia() []Subtarefa {
	return []Subtarefa{
		{
			Nome: "Extrair o que mudou entre versões do protocolo",
			Tipo: "Extração/Interpretação",
			// Não é regra enumerável: é extração/interpretação de linguagem
			// natural sobre o texto do protocolo.
			P1: false,
			// A extração em si não causa, diretamente, um erro caro e
			// irreversível.
			P2: false,
			// O comportamento muda bastante conforme o que de fato mudou
			// entre as duas versões do protocolo.
			P3: true,
		},
		{
			Nome: "Classificar o tipo de emenda (administrativa/substancial)",
			Tipo: "Decisão de Negócio",
			// Segue um critério fixo definido pela ANVISA: regra finita
			// cobre os casos reais.
			P1: true,
			P2: false,
			P3: false,
		},
	}
}

func TestDecomporTarefaHibrida_ExtrairOQueMudouEhAgenteAutonomo(t *testing.T) {
	resultado := DecomporTarefaHibrida(subtarefasReferencia())
	if resultado[0].Classificacao != AgenteAutonomo {
		t.Errorf("resultado[0].Classificacao = %q, want %q", resultado[0].Classificacao, AgenteAutonomo)
	}
}

func TestDecomporTarefaHibrida_ClassificarTipoDeEmendaEhRegraDeterministica(t *testing.T) {
	resultado := DecomporTarefaHibrida(subtarefasReferencia())
	if resultado[1].Classificacao != RegraDeterministica {
		t.Errorf("resultado[1].Classificacao = %q, want %q", resultado[1].Classificacao, RegraDeterministica)
	}
}

func TestDecomporTarefaHibrida_PreservaCamposOriginaisDeCadaSubtarefa(t *testing.T) {
	entrada := subtarefasReferencia()
	resultado := DecomporTarefaHibrida(entrada)

	if len(resultado) != len(entrada) {
		t.Fatalf("len(resultado) = %d, want %d", len(resultado), len(entrada))
	}

	if resultado[0].Nome != entrada[0].Nome {
		t.Errorf("resultado[0].Nome = %q, want %q", resultado[0].Nome, entrada[0].Nome)
	}
	if resultado[0].Tipo != entrada[0].Tipo {
		t.Errorf("resultado[0].Tipo = %q, want %q", resultado[0].Tipo, entrada[0].Tipo)
	}
	if resultado[0].P1 != entrada[0].P1 || resultado[0].P2 != entrada[0].P2 || resultado[0].P3 != entrada[0].P3 {
		t.Errorf("resultado[0] p1/p2/p3 = %v/%v/%v, want %v/%v/%v",
			resultado[0].P1, resultado[0].P2, resultado[0].P3, entrada[0].P1, entrada[0].P2, entrada[0].P3)
	}

	if resultado[1].Nome != entrada[1].Nome {
		t.Errorf("resultado[1].Nome = %q, want %q", resultado[1].Nome, entrada[1].Nome)
	}
	if resultado[1].Tipo != entrada[1].Tipo {
		t.Errorf("resultado[1].Tipo = %q, want %q", resultado[1].Tipo, entrada[1].Tipo)
	}
}
