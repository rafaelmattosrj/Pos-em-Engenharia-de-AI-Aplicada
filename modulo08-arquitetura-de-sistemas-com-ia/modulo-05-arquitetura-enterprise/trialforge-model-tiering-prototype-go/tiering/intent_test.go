package tiering

import "testing"

func TestClassificarIntencao_ReconheceSinteseCsrPorPalavraCsr(t *testing.T) {
	if got := ClassificarIntencao("Preciso da síntese do CSR final desse estudo."); got != SinteseCSR {
		t.Errorf("esperava %q, obteve %q", SinteseCSR, got)
	}
}

func TestClassificarIntencao_ReconheceSinteseCsrPorRelatorioFinal(t *testing.T) {
	if got := ClassificarIntencao("Quero o relatório final do estudo."); got != SinteseCSR {
		t.Errorf("esperava %q, obteve %q", SinteseCSR, got)
	}
}

func TestClassificarIntencao_ReconheceConsultaClausulaComoPadrao(t *testing.T) {
	if got := ClassificarIntencao("Quais são as regras de assentimento pra menores?"); got != ConsultaClausula {
		t.Errorf("esperava %q, obteve %q", ConsultaClausula, got)
	}
}
