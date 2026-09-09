package cleaning

type RelatorioPorCaso struct {
	ContagensAntes  map[string]int
	DistAntes       map[string]float64
	EntropiaAntes   float64
	NEfetivoAntes   float64
	ContagensDepois map[string]int
	DistDepois      map[string]float64
	EntropiaDepois  float64
	NEfetivoDepois  float64
	Alocacao        map[string]int
}

type ResultadoPipeline struct {
	Original               int
	AposDedup              int
	DuplicatasRemovidas    int
	ResultadosDedupPorCaso map[string]ResultadoPorCaso
	TotalParesForcaBruta   int64
	TotalCandidatosLSH     int64
	Final                  int
	RelatorioPorCaso       map[string]RelatorioPorCaso
	ExemplosFinal          []Exemplo
}

// LimparEBalancear e o pipeline completo: dedup -> balanceamento -> diversidade.
func LimparEBalancear(exemplos []Exemplo, alpha float64, alvos map[string]int) ResultadoPipeline {
	dedup := RemoverQuaseDuplicatas(exemplos)
	atual := dedup.Mantidos
	relatorioPorCaso := map[string]RelatorioPorCaso{}

	for _, caso := range []string{"amplitude-auto", "amplitude-saude-empresarial"} {
		contagensAntes := ContarPorFonte(atual, caso)
		distAntes := DistribuicaoDe(contagensAntes)

		resultado := BalancearPorTemperatura(atual, caso, alpha, alvos[caso])
		atual = resultado.Exemplos

		contagensDepois := ContarPorFonte(atual, caso)
		distDepois := DistribuicaoDe(contagensDepois)

		relatorioPorCaso[caso] = RelatorioPorCaso{
			ContagensAntes: contagensAntes, DistAntes: distAntes,
			EntropiaAntes: EntropiaShannon(distAntes), NEfetivoAntes: NumeroEfetivoFontes(distAntes),
			ContagensDepois: contagensDepois, DistDepois: distDepois,
			EntropiaDepois: EntropiaShannon(distDepois), NEfetivoDepois: NumeroEfetivoFontes(distDepois),
			Alocacao: resultado.Alocacao,
		}
	}

	return ResultadoPipeline{
		Original: len(exemplos), AposDedup: len(dedup.Mantidos), DuplicatasRemovidas: dedup.Removidos,
		ResultadosDedupPorCaso: dedup.ResultadosPorCaso, TotalParesForcaBruta: dedup.TotalParesForcaBruta,
		TotalCandidatosLSH: dedup.TotalCandidatosLSH, Final: len(atual), RelatorioPorCaso: relatorioPorCaso,
		ExemplosFinal: atual,
	}
}
