package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"agent-memory/model"
)

// LongTermMemory persiste fatos em disco entre execuções do agente —
// equivalente a LongTermMemory.java. Não duplica fatos com o mesmo
// conteúdo (deduplicação case-insensitive por Content).
type LongTermMemory struct {
	mu          sync.Mutex
	facts       []model.MemoryFact
	storageFile string
}

// NewLongTermMemory cria a memória de longo prazo apontando para
// <memoryPath>/long-term.json, carregando o conteúdo existente se houver.
func NewLongTermMemory(memoryPath string) *LongTermMemory {
	m := &LongTermMemory{storageFile: filepath.Join(memoryPath, "long-term.json")}
	m.load()
	return m
}

func (m *LongTermMemory) load() {
	os.MkdirAll(filepath.Dir(m.storageFile), 0o755)

	data, err := os.ReadFile(m.storageFile)
	if err != nil {
		return
	}
	var facts []model.MemoryFact
	if err := json.Unmarshal(data, &facts); err != nil {
		return
	}
	m.facts = facts
}

// AddFact adiciona um fato à memória, ignorando silenciosamente se o
// conteúdo já existir.
func (m *LongTermMemory) AddFact(fact model.MemoryFact) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, f := range m.facts {
		if strings.EqualFold(f.Content, fact.Content) {
			return
		}
	}
	m.facts = append(m.facts, fact)
	m.persist()
}

// GetFacts retorna todos os fatos não expirados.
func (m *LongTermMemory) GetFacts() []model.MemoryFact {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var result []model.MemoryFact
	for _, f := range m.facts {
		if f.ExpiresAt == nil || f.ExpiresAt.After(now) {
			result = append(result, f)
		}
	}
	return result
}

// GetFactsBySource filtra fatos por fonte (ex.: "usuario", "api_externa").
func (m *LongTermMemory) GetFactsBySource(source string) []model.MemoryFact {
	var result []model.MemoryFact
	for _, f := range m.GetFacts() {
		if strings.EqualFold(f.Source, source) {
			result = append(result, f)
		}
	}
	return result
}

// RemoveFact remove um fato pelo ID.
func (m *LongTermMemory) RemoveFact(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered := m.facts[:0]
	for _, f := range m.facts {
		if f.ID != id {
			filtered = append(filtered, f)
		}
	}
	m.facts = filtered
	m.persist()
}

func (m *LongTermMemory) persist() {
	data, err := json.Marshal(m.facts)
	if err != nil {
		panic("falha ao persistir memoria de longo prazo: " + err.Error())
	}
	if err := os.WriteFile(m.storageFile, data, 0o644); err != nil {
		panic("falha ao persistir memoria de longo prazo: " + err.Error())
	}
}
