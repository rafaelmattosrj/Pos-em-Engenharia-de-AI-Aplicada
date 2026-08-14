package dataset

import "testing"

// Cenário 1 de EvalsTest.java: EvalDataset deve carregar 5 cenários do
// eval-dataset.json.
func TestLoad_ReturnsFiveScenarios(t *testing.T) {
	ds := Load()
	scenarios := ds.Scenarios()

	if len(scenarios) != 5 {
		t.Fatalf("esperava 5 cenarios, obteve %d", len(scenarios))
	}

	expectedIDs := []string{
		"deploy-latency-001", "memory-leak-002", "database-slow-003",
		"service-down-004", "cpu-spike-005",
	}
	for i, id := range expectedIDs {
		if scenarios[i].ID != id {
			t.Errorf("scenario[%d].ID = %q, esperava %q", i, scenarios[i].ID, id)
		}
	}

	validDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
	for _, s := range scenarios {
		if len(s.ExpectedTools) == 0 {
			t.Errorf("scenario %q sem expectedTools", s.ID)
		}
		if !validDifficulties[s.Difficulty] {
			t.Errorf("scenario %q com difficulty invalida: %q", s.ID, s.Difficulty)
		}
	}
}
