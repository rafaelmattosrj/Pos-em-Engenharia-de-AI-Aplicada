// Demo: imprime os resultados das calculadoras de trade-off de LoRA/PEFT com
// dado real capturado nesta disciplina. Nao dispara os subprocessos MLX
// (adaptercomparison/rankadapter) nem a chamada HTTP real (managedapi) --
// isso depende de hardware/rede locais; a demo mostra a logica pura.
package main

import (
	"fmt"
	"strings"

	"lora-peft-toolkit/fullvslora"
	"lora-peft-toolkit/lorarank"
	"lora-peft-toolkit/managedapi"
	"lora-peft-toolkit/regionalnpv"
)

func main() {
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("LoRA rank trade-off (Módulo 4.3)")
	fmt.Println(strings.Repeat("=", 72))
	for _, e := range lorarank.ExecucoesReais {
		fmt.Printf("  Rank %d: val loss %.3f -> %.3f (%.2f%% redução), pico mem %.3fGB\n",
			e.Rank, e.ValLossInicial, e.ValLossFinal, lorarank.CalcularReducaoValLoss(e), e.PicoMemGB)
	}
	recomendado := lorarank.RecomendarRankMinimo(lorarank.ExecucoesReais, 0.10)
	fmt.Printf("  Recomendação (margem 10%%): rank %d\n", recomendado.Rank)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("Full fine-tuning vs. LoRA (Módulo 4.4)")
	fmt.Println(strings.Repeat("=", 72))
	decisao := fullvslora.ValeAPenaFullFineTuning(fullvslora.ConfiguracoesReais, 20)
	valeAPena := "NÃO"
	if decisao.ValeAPena {
		valeAPena = "SIM"
	}
	fmt.Printf("  Melhor LoRA: %s | Ganho do full fine-tuning: %.2f%% | Vale a pena (limiar 20%%)? %s\n",
		decisao.MelhorLora, decisao.GanhoPercentual, valeAPena)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("NPV: parceria regional, GPU alugada vs. LoRA local (Módulo 4.1)")
	fmt.Println(strings.Repeat("=", 72))
	npvGerenciado := regionalnpv.CalcularNPV(regionalnpv.ParamsBase(regionalnpv.CustoJobGerenciado))
	npvLora := regionalnpv.CalcularNPV(regionalnpv.ParamsBase(regionalnpv.CustoLoraLocal))
	breakevenGerenciado := "não atinge"
	if npvGerenciado.MesBreakeven != nil {
		breakevenGerenciado = fmt.Sprintf("mês %d", *npvGerenciado.MesBreakeven)
	}
	fmt.Printf("  GPU alugada: NPV em 24 meses = R$ %.2f (breakeven: %s)\n", npvGerenciado.Npv, breakevenGerenciado)
	fmt.Printf("  LoRA local:  NPV em 24 meses = R$ %.2f (breakeven: mês %d)\n", npvLora.Npv, *npvLora.MesBreakeven)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("Preview de requisição LoRA gerenciada (Together AI, Módulo 4.2)")
	fmt.Println(strings.Repeat("=", 72))
	requisicao, err := managedapi.MontarRequisicaoPadrao("", "file-exemplo-amplitude-seguros")
	if err != nil {
		panic(err)
	}
	fmt.Println("  URL:", requisicao.URL)
	fmt.Println("  Authorization:", requisicao.Headers["Authorization"])
	fmt.Printf("  Body: %+v\n", requisicao.Body)
	fmt.Println("  (sem TOGETHER_API_KEY no ambiente -- modo preview, nada é enviado de verdade)")

	fmt.Println()
	fmt.Println("Nota: adaptercomparison e rankadapter (comparação real via mlx_lm.generate,")
	fmt.Println("Módulos 4.2/4.3) não têm demo aqui -- dependem de MLX + modelo/adaptador locais.")
	fmt.Println("Ver os testes dos pacotes para o parsing validado contra saída real.")
}
