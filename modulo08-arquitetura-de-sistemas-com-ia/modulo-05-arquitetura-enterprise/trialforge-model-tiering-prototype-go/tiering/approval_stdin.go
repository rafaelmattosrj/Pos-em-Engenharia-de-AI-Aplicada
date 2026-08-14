package tiering

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// StdinApprovalPrompt é a implementação real de ApprovalPrompt: lê uma linha
// de um io.Reader (tipicamente os.Stdin) — funciona tanto em terminal
// interativo quanto com stdin redirecionado (ex.: `echo s | ./programa`).
// Reaproveita um único *bufio.Reader por toda a execução, evitando o bug real
// documentado no original em JS (recriar um readline.Interface a cada chamada
// perde entrada vinda de stdin não-interativo).
type StdinApprovalPrompt struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewStdinApprovalPrompt(in io.Reader, out io.Writer) *StdinApprovalPrompt {
	return &StdinApprovalPrompt{reader: bufio.NewReader(in), writer: out}
}

func (s *StdinApprovalPrompt) Approve(rascunho string) (bool, error) {
	fmt.Fprintln(s.writer, "\n[Approval Gate] Rascunho aguardando aprovação antes de virar oficial:")
	fmt.Fprintln(s.writer, "   ", rascunho)
	fmt.Fprint(s.writer, "\nAprovar? (s/n) ")

	linha, err := s.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("falha ao ler resposta de aprovação: %w", err)
	}
	resposta := strings.ToLower(strings.TrimSpace(linha))
	return resposta == "s", nil
}
