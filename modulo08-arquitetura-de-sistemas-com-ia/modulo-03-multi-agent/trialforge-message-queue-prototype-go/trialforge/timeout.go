package trialforge

import (
	"fmt"
	"time"
)

// ErroDeTimeout e o porte de ErroDeTimeout (JS) / ErroDeTimeout (Python):
// sinaliza que um agente nao respondeu dentro do timeout — descoberto de
// fora, pela corrida contra um temporizador real, nunca por uma excecao
// lancada pelo proprio agente.
type ErroDeTimeout struct {
	NomeAgente string
	TimeoutMs  int
}

func (e *ErroDeTimeout) Error() string {
	return fmt.Sprintf("%s: timeout de %dms excedido — agente não respondeu a tempo.", e.NomeAgente, e.TimeoutMs)
}

// resultado carrega o par (valor, erro) de uma operacao assincrona pelo canal —
// equivalente ao par (resolve, reject) de uma Promise (JS) / Future (Python).
type resultado[T any] struct {
	valor T
	erro  error
}

// iniciarComTimeout dispara fn em uma goroutine e devolve IMEDIATAMENTE um
// canal com o resultado — nao bloqueia quem chamou. Isso e o que permite
// iniciar ICF e CSR concorrentemente (Parallel) antes de esperar por
// qualquer um dos dois, exatamente como Promise.race (JS) / asyncio.wait_for
// (Python) rodando ao lado de outra corrida.
//
// Se fn nao responder dentro de timeoutMs, quem "ganha a corrida" no select
// e o temporizador (time.After) — ninguem precisa que o agente "avise" que
// travou; o Supervisor descobre sozinho, de fora. A goroutine de fn, se
// travada, continua rodando em segundo plano e e descartada quando termina
// (o canal e bufferizado, entao o send nunca bloqueia) — o mesmo
// comportamento do setTimeout "orfao" no JS/Python original.
func iniciarComTimeout[T any](timeoutMs int, nomeAgente string, fn func() (T, error)) <-chan resultado[T] {
	saida := make(chan resultado[T], 1)
	go func() {
		interno := make(chan resultado[T], 1)
		go func() {
			v, err := fn()
			interno <- resultado[T]{valor: v, erro: err}
		}()
		select {
		case r := <-interno:
			saida <- r
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			var zero T
			saida <- resultado[T]{valor: zero, erro: &ErroDeTimeout{NomeAgente: nomeAgente, TimeoutMs: timeoutMs}}
		}
	}()
	return saida
}

// comTimeout e a variante bloqueante de iniciarComTimeout, para o caso comum
// (Sequential) de uma unica chamada aguardada na hora.
func comTimeout[T any](timeoutMs int, nomeAgente string, fn func() (T, error)) (T, error) {
	r := <-iniciarComTimeout(timeoutMs, nomeAgente, fn)
	return r.valor, r.erro
}
