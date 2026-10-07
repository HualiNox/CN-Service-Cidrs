package generator

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func writeDomainRules(rules []string, path, name string) error {
	if len(rules) == 0 {
		log.Printf("no domain rules found for %q; skipping output", filepath.Join(path, name))
		return nil
	}
	if err := mkdirAll(path, false); err != nil {
		return fmt.Errorf("prepare output directory %q: %w", path, err)
	}
	outputPath := filepath.Join(path, name+"-domain.txt")
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}
	defer func() { _ = file.Close() }()
	for _, rule := range rules {
		if _, err := io.WriteString(file, rule+"\n"); err != nil {
			return fmt.Errorf("write output file %q: %w", outputPath, err)
		}
	}
	return nil
}

func writeDomainDirectorySummary(directoryRules map[string][]string, path string) error {
	for dir, rules := range directoryRules {
		if len(rules) == 0 {
			continue
		}
		targetDir := filepath.Join(path, filepath.Dir(dir))
		name := filepath.Base(dir)
		if err := writeDomainRules(rules, targetDir, name); err != nil {
			return fmt.Errorf("write domain directory summary %q: %w", dir, err)
		}
	}
	return nil
}
