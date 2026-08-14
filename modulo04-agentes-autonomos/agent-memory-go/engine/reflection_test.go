package engine

import (
	"context"
	"testing"

	"agent-memory/model"
)

type fakeChatClient struct {
	response string
}

func (f *fakeChatClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	return f.response, nil
}

func TestReflect_SkipsEpisodesWithoutNotableOutcome(t *testing.T) {
	client := &fakeChatClient{response: "nao deveria ser chamado"}
	engine := NewReflectionEngine(client, t.TempDir())

	lessons := engine.Reflect(context.Background(), model.Episode{
		Input: "listar clientes", Outcome: "5 clientes retornados com sucesso",
	})
	if lessons != nil {
		t.Errorf("esperava nil para outcome sem indicacao de erro, obteve %v", lessons)
	}
}

func TestReflect_ExtractsLessonFromErrorOutcome(t *testing.T) {
	client := &fakeChatClient{response: `SITUAÇÃO: timeout ao consultar API externa
AÇÃO: retry automatico com backoff
RESULTADO: sucesso na segunda tentativa
APRENDIZADO: sempre usar retry com backoff exponencial em chamadas externas
GENERALIZABILIDADE: aplica-se a qualquer integracao HTTP externa`}
	engine := NewReflectionEngine(client, t.TempDir())

	lessons := engine.Reflect(context.Background(), model.Episode{
		Input: "consultar metricas", Outcome: "erro de timeout na primeira tentativa",
	})

	if len(lessons) != 1 {
		t.Fatalf("esperava 1 licao extraida, obteve %d", len(lessons))
	}
	if lessons[0].Learning != "sempre usar retry com backoff exponencial em chamadas externas" {
		t.Errorf("aprendizado inesperado: %q", lessons[0].Learning)
	}

	persisted := engine.Lessons()
	if len(persisted) != 1 {
		t.Errorf("esperava 1 licao persistida, obteve %d", len(persisted))
	}
}

func TestReflect_SemLicaoResponseYieldsNoLessons(t *testing.T) {
	client := &fakeChatClient{response: "SEM_LICAO"}
	engine := NewReflectionEngine(client, t.TempDir())

	lessons := engine.Reflect(context.Background(), model.Episode{
		Input: "x", Outcome: "erro trivial ja conhecido",
	})
	if lessons != nil {
		t.Errorf("esperava nil para resposta SEM_LICAO, obteve %v", lessons)
	}
}

func TestReflect_MissingGeneralizabilityIsNotPersisted(t *testing.T) {
	client := &fakeChatClient{response: `SITUAÇÃO: caso especifico
AÇÃO: solucao pontual
RESULTADO: resolvido
APRENDIZADO: essa correcao especifica funcionou
GENERALIZABILIDADE: `}
	engine := NewReflectionEngine(client, t.TempDir())

	lessons := engine.Reflect(context.Background(), model.Episode{Input: "x", Outcome: "erro pontual"})
	if lessons != nil {
		t.Errorf("esperava nil quando generalizabilidade esta vazia, obteve %v", lessons)
	}
}
