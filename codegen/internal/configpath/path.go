package configpath

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

type templateData struct {
	Dir  string
	Path string
	Name string
}

type Builder struct {
	tpl *template.Template
}

func (b Builder) Build(protoPath string) (string, error) {
	filePath := strings.TrimSuffix(protoPath, filepath.Ext(protoPath))

	data := templateData{
		Dir:  filepath.Dir(filePath),
		Path: filePath,
		Name: filepath.Base(filePath),
	}
	writer := &strings.Builder{}
	if err := b.tpl.Execute(writer, data); err != nil {
		return "", fmt.Errorf("failed to execute config path template: %w", err)
	}
	return writer.String(), nil
}

func NewBuilder(pattern string) (Builder, error) {
	tpl, err := template.New("filepath").Parse(pattern)
	if err != nil {
		return Builder{}, fmt.Errorf("failed to parse config pattern: %w", err)
	}

	return Builder{tpl: tpl}, nil
}
