package trialforge

// Estrategia decide como a reacao paralela (ICF + CSR) trata a falha de um
// dos dois agentes.
type Estrategia int

const (
	// EstrategiaAllSettled e a correcao (paragrafo 72-73 do TP): equivalente a
	// Promise.allSettled (JS) / gather(return_exceptions=True) (Python) — cada
	// resultado e preservado independente do outro ter falhado. E o valor
	// zero (default), igual ao default 'promise.allSettled' do original.
	EstrategiaAllSettled Estrategia = iota
	// EstrategiaAll replica o bug documentado nos paragrafos 68-71: equivalente
	// a Promise.all (JS) / gather sem return_exceptions=True (Python) — a
	// primeira falha derruba o lote inteiro, mesmo que o outro agente tenha
	// terminado bem.
	EstrategiaAll
)

// OpcoesFluxo e o porte do objeto de opcoes de rodarFluxoTrialForge /
// rodar_fluxo_trialforge (JS/Python) — os flags que disparam cada cenario de
// falha simulada, mais a estrategia de reacao paralela. O valor zero
// (OpcoesFluxo{}) ja e valido e reproduz o comportamento "sem nenhuma falha
// simulada, estrategia allSettled", igual ao default do original.
type OpcoesFluxo struct {
	Estrategia            Estrategia
	ForcarFalhaProtocolo  bool
	ForcarFalhaICF        bool
	ForcarTravamentoICF   bool
	ComEmendaEtica        bool
	PersistirFalhaNoRetry bool
}
