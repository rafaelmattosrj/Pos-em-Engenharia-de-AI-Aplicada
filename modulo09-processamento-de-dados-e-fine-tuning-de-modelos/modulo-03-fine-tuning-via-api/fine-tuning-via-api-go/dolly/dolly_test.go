package dolly

import (
	"errors"
	"strings"
	"testing"
)

const fixture = "testdata/dolly-sample.jsonl"

func TestCarregarDollyLeTodasAsLinhasNaoVazias(t *testing.T) {
	registros, err := CarregarDolly(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(registros) != 6 {
		t.Fatalf("esperado 6 registros brutos, obtido %d", len(registros))
	}
}

func TestCarregarDollyErraQuandoArquivoNaoExiste(t *testing.T) {
	if _, err := CarregarDolly("testdata/nao-existe.jsonl"); err == nil {
		t.Fatal("esperava erro para arquivo inexistente")
	}
}

func TestFiltrarCompativeisRemoveCategoriaIncompativelEContextoVazio(t *testing.T) {
	registros, err := CarregarDolly(fixture)
	if err != nil {
		t.Fatal(err)
	}
	compativeis := FiltrarCompativeis(registros)
	// exclui: "open_qa" (categoria incompatível) e o "summarization" com context vazio.
	if len(compativeis) != 4 {
		t.Fatalf("esperado 4 registros compatíveis, obtido %d", len(compativeis))
	}
	for _, r := range compativeis {
		if !categoriaCompativel(r.Category) {
			t.Fatalf("categoria incompatível vazou pro filtro: %s", r.Category)
		}
		if strings.TrimSpace(r.Context) == "" || strings.TrimSpace(r.Response) == "" {
			t.Fatalf("registro com contexto/resposta vazio vazou pro filtro: %+v", r)
		}
	}
}

func TestParaSchemaCanonicoGeraIdsPorCategoriaEIndice(t *testing.T) {
	registros, err := CarregarDolly(fixture)
	if err != nil {
		t.Fatal(err)
	}
	mapeados := ParaSchemaCanonico(FiltrarCompativeis(registros))
	if len(mapeados) != 4 {
		t.Fatalf("esperado 4 mapeados, obtido %d", len(mapeados))
	}
	for _, e := range mapeados {
		if e.Caso != Caso {
			t.Fatalf("caso esperado %q, obtido %q", Caso, e.Caso)
		}
		if e.Saida["texto"] == nil {
			t.Fatalf("exemplo sem saida.texto: %+v", e)
		}
	}
}

func TestPrepararDatasetCompletoDedupERemoveDuplicataExata(t *testing.T) {
	resultado, err := PrepararDatasetCompleto(fixture, 3)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Bruto != 6 {
		t.Fatalf("esperado bruto=6, obtido %d", resultado.Bruto)
	}
	if resultado.Compativeis != 4 {
		t.Fatalf("esperado compativeis=4, obtido %d", resultado.Compativeis)
	}
	if resultado.Mapeados != 4 {
		t.Fatalf("esperado mapeados=4, obtido %d", resultado.Mapeados)
	}
	if resultado.ItensRemovidos != 1 {
		t.Fatalf("esperado 1 item removido por dedup (duplicata exata), obtido %d", resultado.ItensRemovidos)
	}
	if resultado.SemDuplicatas != 3 {
		t.Fatalf("esperado semDuplicatas=3, obtido %d", resultado.SemDuplicatas)
	}
	// as 3 fontes restantes (information_extraction, closed_qa, summarization)
	// têm 1 exemplo cada; pedindo alvoTotal=3 cabe tudo, sem sobra.
	if resultado.Balanceado != 3 {
		t.Fatalf("esperado balanceado=3 quando alvoTotal cobre toda a capacidade, obtido %d", resultado.Balanceado)
	}
}

func TestConverterParaGeminiUsaSaidaComoTextoSolto(t *testing.T) {
	registros, err := CarregarDolly(fixture)
	if err != nil {
		t.Fatal(err)
	}
	mapeados := ParaSchemaCanonico(FiltrarCompativeis(registros))
	conv, err := ConverterParaGemini(mapeados[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(conv.Contents) != 2 {
		t.Fatalf("esperado 2 turnos, obtido %d", len(conv.Contents))
	}
	if conv.Contents[1].Role != "model" || conv.Contents[1].Text != "XPTO" {
		t.Fatalf("turno do modelo esperado texto solto \"XPTO\", obtido %+v", conv.Contents[1])
	}
}

func TestRodarPipelineBloqueiaAntesDeQualquerChamadaComHiperparametroInvalido(t *testing.T) {
	uploadChamado := false
	_, err := RodarPipeline(
		ConfigPipeline{CaminhoLocal: "dataset.jsonl", URIDataset: "gs://bucket/dataset.jsonl", EpochCount: 0, LearningRateMultiplier: 2.0},
		true,
		func(caminhoLocal, uriGcs string) error { uploadChamado = true; return nil },
		func(config map[string]any, confirmar bool) (map[string]any, error) {
			return map[string]any{"name": "job-1"}, nil
		},
		func(nomeJob string, aoAtualizar func(map[string]any)) (map[string]any, error) {
			return map[string]any{"state": "JOB_STATE_SUCCEEDED"}, nil
		},
		func(map[string]any) {},
	)
	if err == nil {
		t.Fatal("esperava erro de hiperparâmetro inválido")
	}
	if uploadChamado {
		t.Fatal("upload não deveria ter sido chamado com hiperparâmetro inválido")
	}
}

func TestRodarPipelineEncadeiaUploadCriacaoEAcompanhamento(t *testing.T) {
	var chamadas []string
	resultado, err := RodarPipeline(
		ConfigPipeline{CaminhoLocal: "dataset.jsonl", URIDataset: "gs://bucket/dataset.jsonl", EpochCount: 3, LearningRateMultiplier: 1.0},
		true,
		func(caminhoLocal, uriGcs string) error { chamadas = append(chamadas, "upload"); return nil },
		func(config map[string]any, confirmar bool) (map[string]any, error) {
			chamadas = append(chamadas, "criar")
			if !confirmar {
				return nil, errors.New("deveria estar confirmado")
			}
			return map[string]any{"name": "job-dolly-1"}, nil
		},
		func(nomeJob string, aoAtualizar func(map[string]any)) (map[string]any, error) {
			chamadas = append(chamadas, "acompanhar:"+nomeJob)
			return map[string]any{"state": "JOB_STATE_SUCCEEDED"}, nil
		},
		func(map[string]any) {},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.JobFinal["state"] != "JOB_STATE_SUCCEEDED" {
		t.Fatalf("estado final inesperado: %v", resultado.JobFinal)
	}
	esperado := []string{"upload", "criar", "acompanhar:job-dolly-1"}
	if len(chamadas) != len(esperado) {
		t.Fatalf("ordem de chamadas inesperada: %v", chamadas)
	}
	for i := range esperado {
		if chamadas[i] != esperado[i] {
			t.Fatalf("ordem de chamadas inesperada: %v", chamadas)
		}
	}
}
