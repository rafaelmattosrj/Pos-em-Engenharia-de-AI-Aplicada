package service

import (
	"fmt"
	"os"
	"strings"
)

// FileWriterService — equivalente a tools/file_writer.py.
type FileWriterService struct{}

// WriteFile salva o código gerado em um arquivo físico em disco, removendo
// eventuais cercas de código markdown (```hcl ... ```) — equivalente a
// write_file.
func (FileWriterService) WriteFile(content, filename string) string {
	target := filename
	if target == "" {
		target = "main.tf"
	}

	cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(content, "```hcl", ""), "```", ""))

	if err := os.WriteFile(target, []byte(cleaned), 0o644); err != nil {
		return fmt.Sprintf("❌ Erro ao salvar '%s': %v", target, err)
	}
	return fmt.Sprintf("✅ File '%s' saved successfully.", target)
}
