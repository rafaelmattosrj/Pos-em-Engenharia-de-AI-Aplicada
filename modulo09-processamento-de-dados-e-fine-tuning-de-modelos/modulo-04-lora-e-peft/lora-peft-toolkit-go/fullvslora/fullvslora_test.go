package fullvslora

import (
	"math"
	"testing"
)

func acha(tipo string) Configuracao {
	return acharConfiguracao(ConfiguracoesReais, tipo)
}

func TestFullFineTuningMelhoraValLossFrenteALoraRank8(t *testing.T) {
	ganho := CalcularGanhoValLoss(acha("LoRA rank 8"), acha("Full fine-tuning"))
	if math.Abs(ganho-31.62) >= 0.5 {
		t.Fatalf("ganho=%v, esperado ~31.62", ganho)
	}
}

func TestFullFineTuningUsa153xMaisParametros(t *testing.T) {
	r := CalcularRazaoCusto(acha("LoRA rank 8"), acha("Full fine-tuning"))
	if math.Abs(r.Parametro-153.5) >= 1 {
		t.Fatalf("razao=%v, esperado ~153.5", r.Parametro)
	}
}

func TestCheckpointFullE73Ponto8xMaior(t *testing.T) {
	r := CalcularRazaoCusto(acha("LoRA rank 8"), acha("Full fine-tuning"))
	if r.Checkpoint != 73.8 {
		t.Fatalf("checkpoint=%v, esperado 73.8", r.Checkpoint)
	}
}

func TestFullUsaMaisMemoriaQueQualquerLora(t *testing.T) {
	full := acha("Full fine-tuning")
	for _, c := range ConfiguracoesReais {
		if len(c.Tipo) >= 4 && c.Tipo[:4] == "LoRA" {
			if full.PicoMemGB <= c.PicoMemGB {
				t.Fatalf("%s: full=%v, lora=%v", c.Tipo, full.PicoMemGB, c.PicoMemGB)
			}
		}
	}
}

func TestValeAPenaComLimiar20PorCento(t *testing.T) {
	r := ValeAPenaFullFineTuning(ConfiguracoesReais, 20)
	if r.MelhorLora != "LoRA rank 16" {
		t.Fatalf("melhorLora=%v, esperado LoRA rank 16", r.MelhorLora)
	}
	if r.ValeAPena {
		t.Fatalf("valeAPena deveria ser false, ganho=%v", r.GanhoPercentual)
	}
}

func TestValeAPenaComLimiar10PorCento(t *testing.T) {
	r := ValeAPenaFullFineTuning(ConfiguracoesReais, 10)
	if !r.ValeAPena {
		t.Fatalf("valeAPena deveria ser true")
	}
}

func TestResultadoInferenciaExemplos(t *testing.T) {
	if !ExemploFacil.AcertouTodosCampos {
		t.Fatal("exemploFacil deveria acertar todos os campos")
	}
	if !ExemploComDistratores.AcertouTodosCampos {
		t.Fatal("exemploComDistratores deveria acertar todos os campos")
	}
}
