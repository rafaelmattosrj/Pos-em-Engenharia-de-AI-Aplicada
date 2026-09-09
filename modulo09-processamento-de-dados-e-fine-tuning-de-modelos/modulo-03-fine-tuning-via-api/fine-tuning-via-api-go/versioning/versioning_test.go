package versioning

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func jobFalso() map[string]any {
	return map[string]any{
		"name": "projects/x/locations/y/tuningJobs/123",
		"baseModel": "gemini-2.5-flash",
		"tunedModelDisplayName": "teste",
		"state": "JOB_STATE_SUCCEEDED",
		"createTime": "2026-08-08T00:00:00Z",
		"endTime": "2026-08-08T01:00:00Z",
		"supervisedTuningSpec": map[string]any{
			"trainingDatasetUri": "gs://bucket/dataset.jsonl",
			"hyperParameters": map[string]any{
				"epochCount": float64(3), "learningRateMultiplier": float64(5), "adapterSize": "ADAPTER_SIZE_FOUR",
			},
		},
		"tuningDataStats": map[string]any{
			"supervisedTuningDataStats": map[string]any{
				"tuningDatasetExampleCount": float64(200), "totalBillableTokenCount": float64(27353),
			},
		},
		"tunedModel": map[string]any{
			"model": "projects/x/locations/y/models/999", "endpoint": "projects/x/locations/y/endpoints/888",
		},
	}
}

func TestGeraFichaCompletaAPartirDeJobBemFormado(t *testing.T) {
	f, err := GerarFichaVersionamento(jobFalso(), "hash-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	if f.JobID != "projects/x/locations/y/tuningJobs/123" || *f.EpochCount != 3 || f.DatasetHashSha256 != "hash-de-teste" {
		t.Fatalf("ficha inesperada: %+v", f)
	}
}

func TestRejeitaJobSemName(t *testing.T) {
	if _, err := GerarFichaVersionamento(map[string]any{}, "hash"); err == nil || !strings.Contains(err.Error(), "job inválido") {
		t.Fatalf("esperava erro, obtido: %v", err)
	}
}

func TestValidacaoAceitaFichaCompleta(t *testing.T) {
	f, _ := GerarFichaVersionamento(jobFalso(), "hash-de-teste")
	if err := ValidarFichaCompleta(f); err != nil {
		t.Fatal(err)
	}
}

func TestCustoEstimadoUsaDuracaoRealDoJob(t *testing.T) {
	f, _ := GerarFichaVersionamento(jobFalso(), "hash-de-teste")
	if f.CustoEstimado.DuracaoFormatada != "60min 0s" {
		t.Fatalf("duracao inesperada: %s", f.CustoEstimado.DuracaoFormatada)
	}
	if f.CustoEstimado.ConsumerMinUsd != 0.40 || f.CustoEstimado.ConsumerMaxUsd != 0.80 {
		t.Fatalf("custo consumer inesperado: %+v", f.CustoEstimado)
	}
	if f.CustoEstimado.H100MinUsd != 2.50 || f.CustoEstimado.H100MaxUsd != 4.00 {
		t.Fatalf("custo H100 inesperado: %+v", f.CustoEstimado)
	}
}

func TestFormataDuracaoRealDoJobDeProducao(t *testing.T) {
	segundos, err := DuracaoSegundos("2026-08-08T02:38:12.307201Z", "2026-08-08T03:23:54.310390Z")
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatarDuracaoComEspaco(segundos); got != "45min 42s" {
		t.Fatalf("esperado 45min 42s, obtido %s", got)
	}
}

func TestModelCardIncluiHashEEndpoint(t *testing.T) {
	f, _ := GerarFichaVersionamento(jobFalso(), "hash-de-teste-abc123")
	md := GerarModelCardMarkdown(f)
	if !strings.Contains(md, "hash-de-teste-abc123") || !strings.Contains(md, "projects/x/locations/y/endpoints/888") {
		t.Fatalf("model card incompleto: %s", md)
	}
}

func TestModelCardIncluiCustoRealEFaixaDeGpu(t *testing.T) {
	f, _ := GerarFichaVersionamento(jobFalso(), "hash-de-teste")
	md := GerarModelCardMarkdown(f)
	for _, esperado := range []string{"## Custo real", "R$2,39", "82.059 unidades cobradas", "US$ 0,40-0,80", "US$ 2,50-4,00"} {
		if !strings.Contains(md, esperado) {
			t.Fatalf("model card nao contem %q:\n%s", esperado, md)
		}
	}
}

func TestCustoRealUsaTokensXEpocasXTaxaReal(t *testing.T) {
	f, _ := GerarFichaVersionamento(jobFalso(), "hash-de-teste")
	if f.CustoReal.Unidades != 82059 || f.CustoReal.CustoReais != 2.39 {
		t.Fatalf("custo real inesperado: %+v", f.CustoReal)
	}
}

func TestSecaoDpoCitaJobRealDePreferenceTuning(t *testing.T) {
	secao := GerarSecaoDpo()
	for _, esperado := range []string{"## Continuação: preference tuning (DPO)", "tuningJobs/3733013646142341120", "17min 32s", "40 pares de preferência", "preference-dataset-amplitude.jsonl"} {
		if !strings.Contains(secao, esperado) {
			t.Fatalf("secao DPO nao contem %q:\n%s", esperado, secao)
		}
	}
}

func TestHashDoMesmoArquivoEIdentico(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "_teste_hash.jsonl")
	if err := os.WriteFile(caminho, []byte(`{"a":1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h1, err := CalcularHashDataset(caminho)
	if err != nil {
		t.Fatal(err)
	}
	h2, _ := CalcularHashDataset(caminho)
	if h1 != h2 || len(h1) != 64 {
		t.Fatalf("hash inconsistente: %s vs %s", h1, h2)
	}
}
