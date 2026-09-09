package automation

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMontaComandoUploadCorreto(t *testing.T) {
	c, err := MontarComandoUpload("dataset.jsonl", "gs://bucket/dataset.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if c != `gsutil cp "dataset.jsonl" "gs://bucket/dataset.jsonl"` {
		t.Fatalf("comando inesperado: %s", c)
	}
}

func TestRejeitaDatasetQueNaoEhJsonl(t *testing.T) {
	if _, err := MontarComandoUpload("dataset.json", "gs://bucket/x.jsonl"); err == nil || !strings.Contains(err.Error(), "jsonl") {
		t.Fatalf("esperava erro sobre jsonl, obtido: %v", err)
	}
}

func TestRejeitaDestinoQueNaoEhGs(t *testing.T) {
	if _, err := MontarComandoUpload("dataset.jsonl", "/local/path"); err == nil || !strings.Contains(err.Error(), "gs://") {
		t.Fatalf("esperava erro sobre gs://, obtido: %v", err)
	}
}

func TestBloqueiaCriacaoDeJobSemConfirmar(t *testing.T) {
	if err := ExigirConfirmacao(false); err == nil || !strings.Contains(err.Error(), "confirmar") {
		t.Fatalf("esperava bloqueio, obtido: %v", err)
	}
}

func TestPermiteQuandoConfirmarEhTrue(t *testing.T) {
	if err := ExigirConfirmacao(true); err != nil {
		t.Fatal(err)
	}
}

func TestPrimeiroBackoffAplicaOFator(t *testing.T) {
	if got := CalcularProximoIntervalo(5000, 1.5, 60000); got != 7500 {
		t.Fatalf("esperado 7500, obtido %d", got)
	}
}

func TestBackoffNuncaUltrapassaOTeto(t *testing.T) {
	if got := CalcularProximoIntervalo(50000, 1.5, 60000); got != 60000 {
		t.Fatalf("esperado 60000, obtido %d", got)
	}
}

func TestReconheceEstadoTerminalNaPrimeiraConsulta(t *testing.T) {
	chamadasConsulta := 0
	chamadasEspera := 0
	job, err := AcompanharAteFinalizar("job-falso",
		func(nomeJob string) (map[string]any, error) {
			chamadasConsulta++
			return map[string]any{"state": "JOB_STATE_SUCCEEDED"}, nil
		},
		func(d time.Duration) { chamadasEspera++ },
		OpcoesPadrao())
	if err != nil {
		t.Fatal(err)
	}
	if job["state"] != "JOB_STATE_SUCCEEDED" || chamadasConsulta != 1 || chamadasEspera != 0 {
		t.Fatalf("resultado inesperado: job=%v consultas=%d esperas=%d", job, chamadasConsulta, chamadasEspera)
	}
}

func TestEsperaEConsultaEnquantoRunning(t *testing.T) {
	sequencia := []string{"JOB_STATE_PENDING", "JOB_STATE_RUNNING", "JOB_STATE_RUNNING", "JOB_STATE_SUCCEEDED"}
	indice := 0
	chamadasEspera := 0
	job, err := AcompanharAteFinalizar("job-falso",
		func(nomeJob string) (map[string]any, error) {
			s := sequencia[indice]
			indice++
			return map[string]any{"state": s}, nil
		},
		func(d time.Duration) { chamadasEspera++ },
		OpcoesPadrao())
	if err != nil {
		t.Fatal(err)
	}
	if job["state"] != "JOB_STATE_SUCCEEDED" || indice != len(sequencia) || chamadasEspera != len(sequencia)-1 {
		t.Fatalf("resultado inesperado: indice=%d esperas=%d", indice, chamadasEspera)
	}
}

func TestConsultaComRetryAbsorveFalhaTransiente(t *testing.T) {
	tentativas := 0
	job, err := ConsultarComRetry(func(nomeJob string) (map[string]any, error) {
		tentativas++
		if tentativas < 3 {
			return nil, errors.New("falha transiente")
		}
		return map[string]any{"state": "JOB_STATE_SUCCEEDED"}, nil
	}, "job-falso", 3, time.Millisecond, func(time.Duration) {})
	if err != nil {
		t.Fatal(err)
	}
	if job["state"] != "JOB_STATE_SUCCEEDED" || tentativas != 3 {
		t.Fatalf("tentativas=%d job=%v", tentativas, job)
	}
}
