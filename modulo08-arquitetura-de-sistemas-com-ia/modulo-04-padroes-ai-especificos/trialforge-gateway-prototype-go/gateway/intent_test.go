package gateway

import "testing"

// Mesmos 7 casos usados pelos testes puros dos dois originais.
func TestClassificarIntencao(t *testing.T) {
	casos := []struct {
		pergunta string
		esperado string
	}{
		{"Preciso da síntese do CSR final desse estudo.", "sintese_csr"},
		{"Quero o relatório final do estudo.", "sintese_csr"},
		{"Como os eventos adversos aparecem no relatório final?", "sintese_csr"},
		{"Qual é o critério de idade mínima pra participar desse estudo?", "consulta_protocolo"},
		{"Quais são os critérios de exclusão desse protocolo?", "consulta_protocolo"},
		{"Quais são as regras de assentimento pra menores?", "consulta_icf"},
		{"Qual o prazo de armazenamento das amostras biológicas?", "consulta_icf"},
	}

	for _, c := range casos {
		if got := ClassificarIntencao(c.pergunta); got != c.esperado {
			t.Errorf("ClassificarIntencao(%q) = %q, esperado %q", c.pergunta, got, c.esperado)
		}
	}
}

func TestClassificarIntencaoCaseInsensitive(t *testing.T) {
	if got := ClassificarIntencao("SÍNTESE do CSR"); got != "sintese_csr" {
		t.Errorf("esperado sintese_csr, obtido %q", got)
	}
}
