package agente

import "testing"

// Cobre o mesmo cenario de testarFerramenta() em agent-components-demo.js / .py:
// faixa etaria com menor de idade encontra clausula; faixa so de adultos, nao.

func TestBuscarClausulaAssentimento_FaixaComMenorDeIdade_EncontraClausula(t *testing.T) {
	resultado := BuscarClausulaAssentimento([]int{12, 15, 17})

	if resultado.Texto == "" {
		t.Error("esperava cláusula encontrada, obteve texto vazio")
	}
	if resultado.Fonte != "RDC ANVISA 466/2012, Art. 4º" {
		t.Errorf("fonte inesperada: %q", resultado.Fonte)
	}
	if resultado.Aviso != "" {
		t.Errorf("não esperava aviso, obteve %q", resultado.Aviso)
	}
}

func TestBuscarClausulaAssentimento_FaixaSoDeAdultos_NaoEncontraClausula(t *testing.T) {
	resultado := BuscarClausulaAssentimento([]int{25, 40, 55})

	if resultado.Texto != "" {
		t.Errorf("não esperava cláusula, obteve %q", resultado.Texto)
	}
	if resultado.Aviso == "" {
		t.Error("esperava aviso de população adulta")
	}
}

func TestBuscarClausulaAssentimento_FaixaMista_EncontraClausula(t *testing.T) {
	resultado := BuscarClausulaAssentimento([]int{17, 30, 45})

	if resultado.Texto == "" {
		t.Error("esperava cláusula encontrada para faixa mista com um menor")
	}
}
