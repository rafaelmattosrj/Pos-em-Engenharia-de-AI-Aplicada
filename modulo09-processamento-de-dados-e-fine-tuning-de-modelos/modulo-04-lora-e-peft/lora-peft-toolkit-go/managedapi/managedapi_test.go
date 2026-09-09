package managedapi

import "testing"

func TestUsaEndpointRealDaTogetherAi(t *testing.T) {
	r, err := MontarRequisicaoPadrao("", "file-abc123")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if r.URL != TogetherEndpoint || r.Method != "POST" {
		t.Fatalf("url=%v method=%v", r.URL, r.Method)
	}
}

func TestExigeTrainingFileId(t *testing.T) {
	_, err := MontarRequisicaoPadrao("", "")
	if err == nil {
		t.Fatal("esperava erro por trainingFileId vazio")
	}
}

func TestReaproveitaConfigTreinadaLocal(t *testing.T) {
	r, _ := MontarRequisicaoPadrao("", "file-abc123")
	if r.Body.TrainingType.LoraR != 8 || r.Body.TrainingType.LoraAlpha != 20.0 || r.Body.TrainingType.LoraDropout != 0.0 {
		t.Fatalf("trainingType=%+v", r.Body.TrainingType)
	}
}

func TestCorpoUsaFormatoDocumentado(t *testing.T) {
	r, _ := MontarRequisicaoPadrao("", "file-abc123")
	if r.Body.TrainingType.Type != "Lora" {
		t.Fatalf("type=%v, esperado Lora", r.Body.TrainingType.Type)
	}
	if r.Body.TrainingFile == "" || r.Body.Model == "" {
		t.Fatal("training_file e model não deveriam ficar vazios")
	}
}

func TestSemApiKeyUsaPlaceholder(t *testing.T) {
	r, _ := MontarRequisicaoPadrao("", "file-abc123")
	if r.Headers["Authorization"] != "Bearer <TOGETHER_API_KEY>" {
		t.Fatalf("Authorization=%v", r.Headers["Authorization"])
	}
}

func TestComApiKeyCarregaAChaveDeVerdade(t *testing.T) {
	r, _ := MontarRequisicaoPadrao("sk-real-123", "file-abc123")
	if r.Headers["Authorization"] != "Bearer sk-real-123" {
		t.Fatalf("Authorization=%v", r.Headers["Authorization"])
	}
}
