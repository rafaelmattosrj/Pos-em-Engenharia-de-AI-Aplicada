package tool

import (
	"encoding/json"
	"testing"
)

func TestGetMetricsTool_ReturnsServiceFromArgs(t *testing.T) {
	out := GetMetricsTool{}.Execute(map[string]any{"service": "checkout"})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("saida nao e JSON valido: %v", err)
	}
	if parsed["service"] != "checkout" {
		t.Errorf("esperava service=checkout, obteve %v", parsed["service"])
	}
}

func TestGetLogsTool_LimitsLines(t *testing.T) {
	out := GetLogsTool{}.Execute(map[string]any{"lines": float64(2)})

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	logs, _ := parsed["logs"].([]any)
	if len(logs) != 2 {
		t.Errorf("esperava 2 logs, obteve %d", len(logs))
	}
}

func TestGetDeployHistoryTool_ReturnsThreeDeploys(t *testing.T) {
	out := GetDeployHistoryTool{}.Execute(map[string]any{})

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["total_deploys"] != float64(3) {
		t.Errorf("esperava total_deploys=3, obteve %v", parsed["total_deploys"])
	}
}

func TestSaveIncidentTool_PersistsIncident(t *testing.T) {
	tool := &SaveIncidentTool{}

	out := tool.Execute(map[string]any{"title": "API fora do ar", "severity": "high", "description": "500s em massa"})

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["success"] != true {
		t.Fatalf("esperava success=true, obteve %+v", parsed)
	}

	incidents := tool.AllIncidents()
	if len(incidents) != 1 {
		t.Fatalf("esperava 1 incidente persistido, obteve %d", len(incidents))
	}
	if incidents[0]["title"] != "API fora do ar" {
		t.Errorf("titulo inesperado: %v", incidents[0]["title"])
	}
}
