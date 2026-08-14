// Package engine implementa o motor de reflexão evolutiva — equivalente a
// ReflectionEngine.java (aula 14 do curso).
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"agent-memory/model"

	"github.com/google/uuid"
)

// ChatClient é a única operação de que ReflectionEngine depende.
type ChatClient interface {
	Chat(ctx context.Context, userPrompt string) (string, error)
}

const reflectionPromptTemplate = `Analise esta execução de agente e extraia lições aprendidas:

INPUT: %s

PASSOS EXECUTADOS:
- %s

RESULTADO: %s

Extraia UMA lição no seguinte formato exato (cada campo em uma linha):
SITUAÇÃO: [descrição da situação que gerou o aprendizado]
AÇÃO: [o que foi feito]
RESULTADO: [o que aconteceu]
APRENDIZADO: [lição concisa e acionável]
GENERALIZABILIDADE: [em quais outras situações essa lição se aplica — deixe em branco se não for generalizável]

Se não houver lição relevante, responda apenas: SEM_LICAO
`

// ReflectionEngine analisa episódios de execução e extrai lições
// generalizáveis — equivalente a ReflectionEngine.java. Só processa
// episódios que contêm erros ou comportamentos notáveis.
type ReflectionEngine struct {
	Client ChatClient

	mu          sync.Mutex
	lessons     []model.Lesson
	storageFile string
}

// NewReflectionEngine cria o motor de reflexão apontando para
// <memoryPath>/lessons.json, carregando o conteúdo existente se houver.
func NewReflectionEngine(client ChatClient, memoryPath string) *ReflectionEngine {
	e := &ReflectionEngine{Client: client, storageFile: filepath.Join(memoryPath, "lessons.json")}
	e.load()
	return e
}

func (e *ReflectionEngine) load() {
	os.MkdirAll(filepath.Dir(e.storageFile), 0o755)

	data, err := os.ReadFile(e.storageFile)
	if err != nil {
		return
	}
	var lessons []model.Lesson
	if err := json.Unmarshal(data, &lessons); err != nil {
		return
	}
	e.lessons = lessons
}

// Reflect analisa um episódio e extrai lições, se houver erro ou
// comportamento notável. Retorna slice vazio se o episódio não contém
// aprendizados relevantes.
func (e *ReflectionEngine) Reflect(ctx context.Context, episode model.Episode) []model.Lesson {
	if !meritaReflexao(episode) {
		return nil
	}

	prompt := buildReflectionPrompt(episode)

	response, err := e.Client.Chat(ctx, prompt)
	if err != nil {
		return nil
	}

	extracted := parseLessons(response)
	if extracted == nil {
		return nil
	}

	e.mu.Lock()
	e.lessons = append(e.lessons, extracted...)
	e.persist()
	e.mu.Unlock()

	return extracted
}

// Lessons retorna todas as lições acumuladas.
func (e *ReflectionEngine) Lessons() []model.Lesson {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make([]model.Lesson, len(e.lessons))
	copy(result, e.lessons)
	return result
}

func meritaReflexao(episode model.Episode) bool {
	if episode.Outcome == "" {
		return false
	}
	lower := strings.ToLower(episode.Outcome)
	for _, keyword := range []string{"erro", "error", "falha", "notável", "aprendizado"} {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func buildReflectionPrompt(episode model.Episode) string {
	steps := "nenhum passo registrado"
	if len(episode.Steps) > 0 {
		steps = strings.Join(episode.Steps, "\n- ")
	}
	return fmt.Sprintf(reflectionPromptTemplate, episode.Input, steps, episode.Outcome)
}

// parseLessons faz o parse da resposta do LLM no formato estruturado
// definido no prompt, retornando apenas lições generalizáveis e acionáveis
// (mesmo filtro de LessoesFiltradas em Java: aprendizado e generalizability
// não vazios).
func parseLessons(response string) []model.Lesson {
	if strings.TrimSpace(response) == "" || strings.Contains(response, "SEM_LICAO") {
		return nil
	}

	situacao := extractField(response, "SITUAÇÃO:")
	acao := extractField(response, "AÇÃO:")
	resultado := extractField(response, "RESULTADO:")
	aprendizado := extractField(response, "APRENDIZADO:")
	generalizabilidade := extractField(response, "GENERALIZABILIDADE:")

	if aprendizado == "" || generalizabilidade == "" {
		return nil
	}

	return []model.Lesson{{
		ID:               uuid.NewString(),
		Situation:        situacao,
		Action:           acao,
		Result:           resultado,
		Learning:         aprendizado,
		Generalizability: generalizabilidade,
		ExtractedAt:      time.Now(),
	}}
}

func extractField(text, prefix string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func (e *ReflectionEngine) persist() {
	data, err := json.Marshal(e.lessons)
	if err != nil {
		panic("falha ao persistir licoes: " + err.Error())
	}
	if err := os.WriteFile(e.storageFile, data, 0o644); err != nil {
		panic("falha ao persistir licoes: " + err.Error())
	}
}
