package datarelevance

import (
	"reflect"
	"sort"
	"testing"
)

func TestCandidatoComOs4CriteriosVerdadeirosEAceito(t *testing.T) {
	r := AvaliarCandidato(map[string]bool{
		"contemGroundTruth": true, "producaoReal": true, "cobreVariacao": true, "passaCompliance": true,
	})
	if !r.Aceito || len(r.CriteriosFalhos) != 0 {
		t.Fatalf("esperado aceito sem falhas, obtive %+v", r)
	}
}

func TestCandidatoSemGroundTruthERejeitado(t *testing.T) {
	r := AvaliarCandidato(map[string]bool{
		"contemGroundTruth": false, "producaoReal": true, "cobreVariacao": true, "passaCompliance": true,
	})
	if r.Aceito || !reflect.DeepEqual(r.CriteriosFalhos, []string{"contemGroundTruth"}) {
		t.Fatalf("esperado rejeitado por contemGroundTruth, obtive %+v", r)
	}
}

func TestCandidatoQueFalhaComplianceERejeitado(t *testing.T) {
	r := AvaliarCandidato(map[string]bool{
		"contemGroundTruth": true, "producaoReal": true, "cobreVariacao": true, "passaCompliance": false,
	})
	if r.Aceito || !reflect.DeepEqual(r.CriteriosFalhos, []string{"passaCompliance"}) {
		t.Fatalf("esperado rejeitado por passaCompliance, obtive %+v", r)
	}
}

func TestAplicacaoAos7CandidatosReais(t *testing.T) {
	esperadoAceito := map[string]bool{
		"orcamento-oficina": true, "boletim-ocorrencia": false, "foto-veiculo-danificado": false,
		"transcricao-ligacao": false, "recibo-medico": true, "prontuario-medico-completo": false,
		"cadastro-beneficiarios": false,
	}
	for _, c := range Candidatos {
		r := AvaliarCandidato(c.Criterios)
		if r.Aceito != esperadoAceito[c.ID] {
			t.Errorf("%s: esperado aceito=%v, obtive %v", c.ID, esperadoAceito[c.ID], r.Aceito)
		}
	}
}

func TestExatamente2Dos7CandidatosSaoAceitos(t *testing.T) {
	var aceitos []string
	for _, c := range Candidatos {
		if AvaliarCandidato(c.Criterios).Aceito {
			aceitos = append(aceitos, c.ID)
		}
	}
	sort.Strings(aceitos)
	if !reflect.DeepEqual(aceitos, []string{"orcamento-oficina", "recibo-medico"}) {
		t.Fatalf("esperado [orcamento-oficina recibo-medico], obtive %v", aceitos)
	}
}
