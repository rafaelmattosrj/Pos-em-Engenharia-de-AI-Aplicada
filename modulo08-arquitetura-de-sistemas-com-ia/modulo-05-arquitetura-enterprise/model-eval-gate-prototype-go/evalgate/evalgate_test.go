package evalgate

import (
	"context"
	"math"
	"strings"
	"testing"
)

// fakeGateway é um dublê determinístico de Gateway: devolve chat e embeddings
// pré-configurados, sem chamar nenhum Ollama de verdade.
type fakeGateway struct {
	chatFn  func(ctx context.Context, model, systemPrompt, userPrompt string) (string, error)
	embedFn func(ctx context.Context, model, texto string) ([]float64, error)
}

func (f *fakeGateway) Chat(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	return f.chatFn(ctx, model, systemPrompt, userPrompt)
}

func (f *fakeGateway) Embed(ctx context.Context, model, texto string) ([]float64, error) {
	return f.embedFn(ctx, model, texto)
}

func closeEnough(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// ---------- SimilaridadeCosseno: lógica pura ----------

func TestSimilaridadeCosseno_VetoresIdenticos(t *testing.T) {
	if got := SimilaridadeCosseno([]float64{1, 0, 0}, []float64{1, 0, 0}); !closeEnough(got, 1.0) {
		t.Errorf("esperava 1.0, obteve %v", got)
	}
}

func TestSimilaridadeCosseno_VetoresOrtogonais(t *testing.T) {
	if got := SimilaridadeCosseno([]float64{1, 0, 0}, []float64{0, 1, 0}); !closeEnough(got, 0.0) {
		t.Errorf("esperava 0.0, obteve %v", got)
	}
}

func TestSimilaridadeCosseno_VetoresOpostos(t *testing.T) {
	if got := SimilaridadeCosseno([]float64{1, 0}, []float64{-1, 0}); !closeEnough(got, -1.0) {
		t.Errorf("esperava -1.0, obteve %v", got)
	}
}

// ---------- DecidirPromocao: lógica pura de decisão do gate ----------

func TestDecidirPromocao_ScoreIgual_Promove(t *testing.T) {
	if !DecidirPromocao(0.85, 0.85, 0.02) {
		t.Error("esperava promover")
	}
}

func TestDecidirPromocao_DentroDaTolerancia_Promove(t *testing.T) {
	if !DecidirPromocao(0.85, 0.84, 0.02) {
		t.Error("esperava promover")
	}
}

func TestDecidirPromocao_ExatamenteNoLimiteDaTolerancia_Promove(t *testing.T) {
	// 0.5/0.25 são exatamente representáveis em ponto flutuante binário.
	if !DecidirPromocao(0.5, 0.25, 0.25) {
		t.Error("esperava promover no limite exato da tolerância")
	}
}

func TestDecidirPromocao_AlemDaTolerancia_Bloqueia(t *testing.T) {
	if DecidirPromocao(0.85, 0.80, 0.02) {
		t.Error("esperava bloquear")
	}
}

func TestDecidirPromocao_CandidatoMelhorQueBaseline_Promove(t *testing.T) {
	if !DecidirPromocao(0.80, 0.95, 0.02) {
		t.Error("esperava promover")
	}
}

// ---------- AvaliarCandidato: fluxo ponta a ponta com dublê determinístico ----------

func TestAvaliarCandidato_CalculaMediaDosScores(t *testing.T) {
	goldenSet := []GoldenItem{
		{Pergunta: "pergunta 1", Clausula: "clausula 1", Fonte: "fonte 1"},
		{Pergunta: "pergunta 2", Clausula: "clausula 2", Fonte: "fonte 2"},
	}

	embeddings := map[string][]float64{
		"resposta 1": {1, 0},
		"clausula 1": {1, 0}, // score 1.0
		"resposta 2": {1, 0},
		"clausula 2": {0, 1}, // score 0.0
	}

	fake := &fakeGateway{
		chatFn: func(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
			if strings.Contains(userPrompt, "pergunta 1") {
				return "resposta 1", nil
			}
			return "resposta 2", nil
		},
		embedFn: func(ctx context.Context, model, texto string) ([]float64, error) {
			return embeddings[texto], nil
		},
	}

	gate := NewEvalGate(fake)
	media, err := gate.AvaliarCandidato(context.Background(), "qualquer-modelo", goldenSet)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !closeEnough(media, 0.5) {
		t.Errorf("esperava media 0.5, obteve %v", media)
	}
}

func TestAvaliarCandidatoSemContexto_NaoEnviaClausulaNoPrompt(t *testing.T) {
	goldenSet := []GoldenItem{
		{Pergunta: "pergunta sem contexto", Clausula: "clausula que não deve aparecer", Fonte: "fonte"},
	}

	fake := &fakeGateway{
		chatFn: func(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
			if strings.Contains(userPrompt, "clausula que não deve aparecer") {
				t.Error("gerarRespostaSemContexto não deveria incluir a cláusula no prompt")
			}
			return "resposta sem contexto", nil
		},
		embedFn: func(ctx context.Context, model, texto string) ([]float64, error) {
			switch texto {
			case "resposta sem contexto":
				return []float64{1, 0}, nil
			case "clausula que não deve aparecer":
				return []float64{0, 1}, nil
			}
			return nil, nil
		},
	}

	gate := NewEvalGate(fake)
	media, err := gate.AvaliarCandidatoSemContexto(context.Background(), "gemma4:e2b", goldenSet)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !closeEnough(media, 0.0) {
		t.Errorf("esperava media 0.0, obteve %v", media)
	}
}

func TestGerarResposta_IncluiClausulaEFonteNoPrompt(t *testing.T) {
	fake := &fakeGateway{
		chatFn: func(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
			for _, esperado := range []string{"minha-clausula", "minha-fonte", "minha-pergunta"} {
				if !strings.Contains(userPrompt, esperado) {
					t.Errorf("esperava %q no prompt, não encontrado: %q", esperado, userPrompt)
				}
			}
			return "ok", nil
		},
	}
	gate := NewEvalGate(fake)

	resposta, err := gate.GerarResposta(context.Background(), "modelo", "minha-pergunta", "minha-clausula", "minha-fonte")
	if err != nil {
		t.Fatal(err)
	}
	if resposta != "ok" {
		t.Errorf("esperava \"ok\", obteve %q", resposta)
	}
}
