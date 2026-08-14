package com.iadeva.nexustools.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Equivalente a consult_runbook, definido em
 * labs/modulo10_remediation.py (RAG sobre runbooks).
 */
@Service
public class RunbookService {

    private final String dataPath;

    public RunbookService(@Value("${app.data-path:./data}") String dataPath) {
        this.dataPath = dataPath;
    }

    /**
     * Lê o runbook oficial de um serviço específico e retorna os passos de
     * remediação — equivalente a consult_runbook.
     */
    public String consultRunbook(String serviceName) {
        Path runbookPath = Path.of(dataPath, "runbook_" + serviceName + ".md");
        try {
            return Files.readString(runbookPath);
        } catch (IOException e) {
            return "Error: Runbook for service '" + serviceName + "' not found.";
        }
    }
}
