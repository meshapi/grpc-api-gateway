package configpath

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

type templateData struct {
	Path string
	Name string
}

func Build(protoPath, configPathTemplate string) (string, error) {
	tpl, err := template.New("filepath").Parse(configPathTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse config path template: %w", err)
	}

	filePath := strings.TrimSuffix(protoPath, filepath.Ext(protoPath))

	data := templateData{
		Path: filePath,
		Name: filepath.Base(filePath),
	}
	writer := &strings.Builder{}
	if err := tpl.Execute(writer, data); err != nil {
		return "", fmt.Errorf("failed to execute config path template: %w", err)
	}

	return writer.String(), nil
}
