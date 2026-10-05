package parser

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v3"
)

func parseFile(sourcePath string) (SourceGroup, error) {
	var sourceGroup SourceGroup

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return SourceGroup{}, fmt.Errorf("read source file %q: %w", sourcePath, err)
	}

	if err := yaml.Unmarshal(data, &sourceGroup); err != nil {
		return SourceGroup{}, fmt.Errorf("parse source file %q: %w", sourcePath, err)
	}

	validate := validator.New()
	if err := validate.Struct(sourceGroup); err != nil {
		return SourceGroup{}, fmt.Errorf("validate source file %q: %w", sourcePath, err)
	}

	return sourceGroup, nil
}

func Parse(sourcesDir string) ([]SourceFile, error) {
	var sourceFiles []SourceFile

	err := filepath.WalkDir(sourcesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk source path %q: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		sourceGroup, err := parseFile(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(sourcesDir, path)
		if err != nil {
			return fmt.Errorf("get relative path for source file %q: %w", path, err)
		}

		dir := filepath.Dir(rel)

		sourceFiles = append(sourceFiles, SourceFile{
			Directory: dir,
			Group:     sourceGroup,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk sources directory %q: %w", sourcesDir, err)
	}

	return sourceFiles, nil
}
