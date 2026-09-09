package cleaning

import (
	"math"
	"testing"
)

func acharExemplo(dataset []Exemplo, id string) Exemplo {
	for _, e := range dataset {
		if e.Metadata.ID == id {
			return e
		}
	}
	panic("nao encontrado: " + id)
}

func TestAssinaturaMinHashDeTextoIdenticoProduzSimilaridade1(t *testing.T) {
	dataset := GerarDatasetSimulado()
	coeficientes := GerarCoeficientesHash(MinhashK, MinhashSemente)
	e := acharExemplo(dataset, "amplitude-auto-Oficina Estrela-0")
	sig := AssinaturaMinHash(Shingles(e.Entrada, 5), coeficientes)
	if SimilaridadeMinHashEstimada(sig, sig) != 1.0 {
		t.Fatal("esperado similaridade 1.0 contra si mesmo")
	}
}

func TestMinHashAproximaJaccardExatoParaDuplicataExata(t *testing.T) {
	dataset := GerarDatasetSimulado()
	coeficientes := GerarCoeficientesHash(MinhashK, MinhashSemente)
	a := acharExemplo(dataset, "amplitude-auto-Oficina Estrela-0")
	b := acharExemplo(dataset, "amplitude-auto-Oficina Estrela-0-reenviado")
	sigA := AssinaturaMinHash(Shingles(a.Entrada, 5), coeficientes)
	sigB := AssinaturaMinHash(Shingles(b.Entrada, 5), coeficientes)
	estimado := SimilaridadeMinHashEstimada(sigA, sigB)
	exato := SimilaridadeJaccardExata(a.Entrada, b.Entrada, 5)
	if math.Abs(estimado-exato) >= 0.1 {
		t.Fatalf("erro muito alto: exato=%v estimado=%v", exato, estimado)
	}
}

func TestMinHashAproximaJaccardExatoParaDuplicataComRuidoDeOcr(t *testing.T) {
	dataset := GerarDatasetSimulado()
	coeficientes := GerarCoeficientesHash(MinhashK, MinhashSemente)
	a := acharExemplo(dataset, "amplitude-auto-Oficina Estrela-2")
	b := acharExemplo(dataset, "amplitude-auto-Oficina Estrela-2-ruido-ocr")
	sigA := AssinaturaMinHash(Shingles(a.Entrada, 5), coeficientes)
	sigB := AssinaturaMinHash(Shingles(b.Entrada, 5), coeficientes)
	estimado := SimilaridadeMinHashEstimada(sigA, sigB)
	exato := SimilaridadeJaccardExata(a.Entrada, b.Entrada, 5)
	if math.Abs(estimado-exato) >= 0.15 {
		t.Fatalf("erro muito alto: exato=%v estimado=%v", exato, estimado)
	}
}

func TestLshEncontraExatamenteOs3ParesDeQuaseDuplicataPlantados(t *testing.T) {
	dataset := GerarDatasetSimulado()
	r := EncontrarQuaseDuplicatasMinHashLSH(dataset)
	if len(r.ParesDuplicata) != 3 {
		t.Fatalf("esperado 3 pares, obtive %d", len(r.ParesDuplicata))
	}
}

func TestLshReduzComparacoesEmPeloMenos80Porcento(t *testing.T) {
	dataset := GerarDatasetSimulado()
	r := EncontrarQuaseDuplicatasMinHashLSH(dataset)
	reducao := 1 - float64(r.TotalCandidatosLSH)/float64(r.TotalParesForcaBruta)
	if reducao < 0.8 {
		t.Fatalf("reducao de so %.1f%%", reducao*100)
	}
}

func TestNenhumParNaoDuplicataEConfirmado(t *testing.T) {
	dataset := GerarDatasetSimulado()
	r := EncontrarQuaseDuplicatasMinHashLSH(dataset)
	idsDuplicata := map[string]bool{}
	for _, p := range r.ParesDuplicata {
		idsDuplicata[dataset[p.I].Metadata.ID] = true
		idsDuplicata[dataset[p.J].Metadata.ID] = true
	}
	esperados := []string{
		"amplitude-auto-Oficina Estrela-0", "amplitude-auto-Oficina Estrela-0-reenviado",
		"amplitude-auto-Oficina Estrela-2", "amplitude-auto-Oficina Estrela-2-ruido-ocr",
		"amplitude-saude-empresarial-Clinica Vitalis-0", "amplitude-saude-empresarial-Clinica Vitalis-0-reenviado",
	}
	if len(idsDuplicata) != len(esperados) {
		t.Fatalf("esperado %d ids, obtive %d: %v", len(esperados), len(idsDuplicata), idsDuplicata)
	}
	for _, id := range esperados {
		if !idsDuplicata[id] {
			t.Errorf("id esperado ausente: %s", id)
		}
	}
}

func TestRemocaoDeQuaseDuplicataTiraExatamente3(t *testing.T) {
	dataset := GerarDatasetSimulado()
	r := RemoverQuaseDuplicatas(dataset)
	if r.Removidos != 3 {
		t.Fatalf("esperado 3 removidos, obtive %d", r.Removidos)
	}
}
