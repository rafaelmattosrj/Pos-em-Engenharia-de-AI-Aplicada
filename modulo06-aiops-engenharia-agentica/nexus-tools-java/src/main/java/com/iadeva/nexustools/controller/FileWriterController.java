package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.FileWriterService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente ao endpoint da tool de tools/file_writer.py.
 */
@RestController
public class FileWriterController {

    private final FileWriterService service;

    public FileWriterController(FileWriterService service) {
        this.service = service;
    }

    public record WriteFileRequest(String content, String filename) {}

    @PostMapping("/tools/write-file")
    public ToolResponse writeFile(@RequestBody WriteFileRequest req) {
        return new ToolResponse(service.writeFile(req.content(), req.filename()));
    }
}
