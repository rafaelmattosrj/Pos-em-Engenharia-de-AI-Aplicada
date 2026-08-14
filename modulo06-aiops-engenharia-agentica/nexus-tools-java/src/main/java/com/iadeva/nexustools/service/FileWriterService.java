package com.iadeva.nexustools.service;

import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Equivalente a tools/file_writer.py.
 */
@Service
public class FileWriterService {

    /**
     * Salva o código gerado em um arquivo físico em disco, removendo
     * eventuais cercas de código markdown (```hcl ... ```) — equivalente a
     * write_file.
     */
    public String writeFile(String content, String filename) {
        String target = filename == null || filename.isBlank() ? "main.tf" : filename;
        String cleaned = content.replace("```hcl", "").replace("```", "").strip();

        try {
            Files.writeString(Path.of(target), cleaned);
        } catch (IOException e) {
            return "❌ Erro ao salvar '" + target + "': " + e.getMessage();
        }
        return "✅ File '" + target + "' saved successfully.";
    }
}
