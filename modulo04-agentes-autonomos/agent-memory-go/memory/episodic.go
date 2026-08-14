package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"agent-memory/model"
)

// EpisodicMemory registra cada execução completa do agente como um episódio
// — equivalente a EpisodicMemory.java.
type EpisodicMemory struct {
	mu          sync.Mutex
	episodes    []model.Episode
	storageFile string
}

// NewEpisodicMemory cria a memória episódica apontando para
// <memoryPath>/episodes.json, carregando o conteúdo existente se houver.
func NewEpisodicMemory(memoryPath string) *EpisodicMemory {
	m := &EpisodicMemory{storageFile: filepath.Join(memoryPath, "episodes.json")}
	m.load()
	return m
}

func (m *EpisodicMemory) load() {
	os.MkdirAll(filepath.Dir(m.storageFile), 0o755)

	data, err := os.ReadFile(m.storageFile)
	if err != nil {
		return
	}
	var episodes []model.Episode
	if err := json.Unmarshal(data, &episodes); err != nil {
		return
	}
	m.episodes = episodes
}

// AddEpisode registra um novo episódio de execução.
func (m *EpisodicMemory) AddEpisode(episode model.Episode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.episodes = append(m.episodes, episode)
	m.persist()
}

// GetEpisodes retorna todos os episódios ordenados do mais recente para o
// mais antigo.
func (m *EpisodicMemory) GetEpisodes() []model.Episode {
	m.mu.Lock()
	defer m.mu.Unlock()

	sorted := make([]model.Episode, len(m.episodes))
	copy(sorted, m.episodes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})
	return sorted
}

// GetRecentEpisodes retorna os N episódios mais recentes.
func (m *EpisodicMemory) GetRecentEpisodes(limit int) []model.Episode {
	all := m.GetEpisodes()
	if limit > len(all) {
		limit = len(all)
	}
	return all[:limit]
}

func (m *EpisodicMemory) persist() {
	data, err := json.Marshal(m.episodes)
	if err != nil {
		panic("falha ao persistir memoria episodica: " + err.Error())
	}
	if err := os.WriteFile(m.storageFile, data, 0o644); err != nil {
		panic("falha ao persistir memoria episodica: " + err.Error())
	}
}
