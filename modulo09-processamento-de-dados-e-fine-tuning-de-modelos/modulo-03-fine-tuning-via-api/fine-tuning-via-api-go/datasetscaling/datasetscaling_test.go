package datasetscaling

import (
	"math"
	"testing"
)

func TestDatasetBrutoTem305Exemplos183Auto122Saude(t *testing.T) {
	bruto := GerarDatasetBruto()
	if len(bruto) != 305 {
		t.Fatalf("esperado 305 exemplos, obtido %d", len(bruto))
	}
	auto, saude := 0, 0
	for _, e := range bruto {
		switch e.Caso {
		case "amplitude-auto":
			auto++
		case "amplitude-saude-empresarial":
			saude++
		}
	}
	if auto != 183 {
		t.Fatalf("esperado 183 exemplos amplitude-auto, obtido %d", auto)
	}
	if saude != 122 {
		t.Fatalf("esperado 122 exemplos amplitude-saude-empresarial, obtido %d", saude)
	}
}

func TestNenhumIdSeRepeteNoBruto(t *testing.T) {
	bruto := GerarDatasetBruto()
	ids := make(map[string]bool, len(bruto))
	for _, e := range bruto {
		ids[e.ID] = true
	}
	if len(ids) != len(bruto) {
		t.Fatalf("esperado %d ids unicos, obtido %d", len(bruto), len(ids))
	}
}

func closeTo(got, want, tol float64) bool {
	return math.Abs(got-want) <= tol
}

func TestPipelineReduz305Para300Dedup300Para200Balanceado(t *testing.T) {
	bruto := GerarDatasetBruto()
	alvos := map[string]int{"amplitude-auto": 120, "amplitude-saude-empresarial": 80}
	resultado := LimparEBalancear(bruto, alvos)

	if resultado.Original != 305 {
		t.Fatalf("esperado original=305, obtido %d", resultado.Original)
	}
	if resultado.AposDedup != 300 {
		t.Fatalf("esperado aposDedup=300, obtido %d", resultado.AposDedup)
	}
	if resultado.Total != 200 {
		t.Fatalf("esperado total=200, obtido %d", resultado.Total)
	}

	auto, saude := 0, 0
	for _, e := range resultado.ExemplosFinal {
		switch e.Caso {
		case "amplitude-auto":
			auto++
		case "amplitude-saude-empresarial":
			saude++
		}
	}
	if auto != 120 {
		t.Fatalf("esperado 120 exemplos finais amplitude-auto, obtido %d", auto)
	}
	if saude != 80 {
		t.Fatalf("esperado 80 exemplos finais amplitude-saude-empresarial, obtido %d", saude)
	}

	rAuto := resultado.RelatorioPorCaso["amplitude-auto"]
	rSaude := resultado.RelatorioPorCaso["amplitude-saude-empresarial"]

	if !(rAuto.NEfetivoDepois > rAuto.NEfetivoAntes) {
		t.Fatalf("esperado nEfetivoDepois > nEfetivoAntes (auto): antes=%.3f depois=%.3f", rAuto.NEfetivoAntes, rAuto.NEfetivoDepois)
	}
	if !(rSaude.NEfetivoDepois > rSaude.NEfetivoAntes) {
		t.Fatalf("esperado nEfetivoDepois > nEfetivoAntes (saude): antes=%.3f depois=%.3f", rSaude.NEfetivoAntes, rSaude.NEfetivoDepois)
	}

	if !closeTo(rAuto.NEfetivoAntes, 5.160, 0.01) {
		t.Fatalf("nEfetivoAntes (auto) esperado ~5.160, obtido %.3f", rAuto.NEfetivoAntes)
	}
	if !closeTo(rAuto.NEfetivoDepois, 5.723, 0.01) {
		t.Fatalf("nEfetivoDepois (auto) esperado ~5.723, obtido %.3f", rAuto.NEfetivoDepois)
	}
	if !closeTo(rSaude.NEfetivoAntes, 4.140, 0.01) {
		t.Fatalf("nEfetivoAntes (saude) esperado ~4.140, obtido %.3f", rSaude.NEfetivoAntes)
	}
	if !closeTo(rSaude.NEfetivoDepois, 4.706, 0.01) {
		t.Fatalf("nEfetivoDepois (saude) esperado ~4.706, obtido %.3f", rSaude.NEfetivoDepois)
	}
}

func TestNenhumaFonteAlemDaCapacidadeReal(t *testing.T) {
	bruto := GerarDatasetBruto()
	alvos := map[string]int{"amplitude-auto": 120, "amplitude-saude-empresarial": 80}
	resultado := LimparEBalancear(bruto, alvos)
	rAuto := resultado.RelatorioPorCaso["amplitude-auto"]
	for fonte, n := range rAuto.ContagensDepois {
		if n > rAuto.ContagensAntes[fonte] {
			t.Fatalf("fonte %s alocou %d, mais que a capacidade real %d", fonte, n, rAuto.ContagensAntes[fonte])
		}
	}
}
