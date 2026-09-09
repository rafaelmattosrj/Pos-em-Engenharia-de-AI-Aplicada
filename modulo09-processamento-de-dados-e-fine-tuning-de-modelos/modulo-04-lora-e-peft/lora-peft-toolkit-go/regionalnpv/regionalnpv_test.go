package regionalnpv

import "testing"

func TestVolumeRegionalComGpuAluigadaNpvNegativo(t *testing.T) {
	r := CalcularNPV(ParamsBase(CustoJobGerenciado))
	if r.Npv >= 0 {
		t.Fatalf("npv=%v, esperado < 0", r.Npv)
	}
}

func TestVolumeRegionalComGpuAluigadaNuncaAtingeBreakeven(t *testing.T) {
	r := CalcularNPV(ParamsBase(CustoJobGerenciado))
	if r.MesBreakeven != nil {
		t.Fatalf("mesBreakeven=%v, esperado nil", *r.MesBreakeven)
	}
}

func TestMesmoVolumeComLoraLocalNpvPositivo(t *testing.T) {
	r := CalcularNPV(ParamsBase(CustoLoraLocal))
	if r.Npv <= 0 {
		t.Fatalf("npv=%v, esperado > 0", r.Npv)
	}
}

func TestLoraLocalAtingeBreakevenRapido(t *testing.T) {
	r := CalcularNPV(ParamsBase(CustoLoraLocal))
	if r.MesBreakeven == nil {
		t.Fatal("esperava mesBreakeven não nil")
	}
	if *r.MesBreakeven > 3 {
		t.Fatalf("mesBreakeven=%d, esperado <= 3", *r.MesBreakeven)
	}
}

func TestDiferencaDeNpvEProximaDoCustoFixo(t *testing.T) {
	gerenciado := CalcularNPV(ParamsBase(CustoJobGerenciado))
	lora := CalcularNPV(ParamsBase(CustoLoraLocal))
	diferenca := lora.Npv - gerenciado.Npv
	delta := diferenca - CustoJobGerenciado
	if delta < 0 {
		delta = -delta
	}
	if delta >= 50 {
		t.Fatalf("diferença=%v, esperado próximo de %v", diferenca, CustoJobGerenciado)
	}
}
