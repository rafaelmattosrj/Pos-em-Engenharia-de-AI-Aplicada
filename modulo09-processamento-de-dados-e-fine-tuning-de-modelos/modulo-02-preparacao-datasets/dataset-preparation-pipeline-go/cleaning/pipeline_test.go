package cleaning

import "testing"

func TestPipelineCompletoProduzDatasetFinalMenorComDiversidadeMaior(t *testing.T) {
	dataset := GerarDatasetSimulado()
	resultado := LimparEBalancear(dataset, AlphaTemperatura,
		map[string]int{"amplitude-auto": 20, "amplitude-saude-empresarial": 14})

	if resultado.DuplicatasRemovidas != 3 {
		t.Fatalf("esperado 3 duplicatas removidas, obtive %d", resultado.DuplicatasRemovidas)
	}
	if !(resultado.Final < resultado.Original) {
		t.Fatalf("esperado final < original, obtive final=%d original=%d", resultado.Final, resultado.Original)
	}
	auto := resultado.RelatorioPorCaso["amplitude-auto"]
	if !(auto.NEfetivoDepois > auto.NEfetivoAntes) {
		t.Fatalf("esperado diversidade maior depois pra amplitude-auto: antes=%v depois=%v", auto.NEfetivoAntes, auto.NEfetivoDepois)
	}
	saude := resultado.RelatorioPorCaso["amplitude-saude-empresarial"]
	if !(saude.NEfetivoDepois > saude.NEfetivoAntes) {
		t.Fatalf("esperado diversidade maior depois pra amplitude-saude-empresarial: antes=%v depois=%v", saude.NEfetivoAntes, saude.NEfetivoDepois)
	}

	temAuto, temSaude := false, false
	for _, e := range resultado.ExemplosFinal {
		if e.Metadata.Caso == "amplitude-auto" {
			temAuto = true
		}
		if e.Metadata.Caso == "amplitude-saude-empresarial" {
			temSaude = true
		}
	}
	if !temAuto || !temSaude {
		t.Fatal("esperado ambos os casos presentes no dataset final")
	}
}
