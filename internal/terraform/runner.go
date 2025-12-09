package terraform

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

//go:embed templates/main.tf.tpl
var templatesFS embed.FS

type Runner struct {
	WorkDir string
}

type TemplateData struct {
	Name      string
	Timestamp string
	Port      int
}

func NewRunner(workDir string) *Runner {
	return &Runner{WorkDir: workDir}
}

func (r *Runner) PrepareTerraformFiles(data TemplateData) error {
	err := os.MkdirAll(r.WorkDir, 0755)
	if err != nil {
		return err
	}

	// tplPath := "internal/terraform/templates/main.tf.tpl"
	tplContent, err := templatesFS.ReadFile("templates/main.tf.tpl")
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("main").Parse(string(tplContent))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return err
	}

	mainTfPath := filepath.Join(r.WorkDir, "main.tf")
	return os.WriteFile(mainTfPath, buf.Bytes(), 0644)
}

func (r *Runner) RunTerraform(ctx context.Context, logFilePath string) error {
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	commands := []struct {
		Name string
		Args []string
	}{
		{"init", nil},
		{"plan", nil},
		{"apply", []string{"-auto-approve"}},
	}

	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform binary not found in PATH")
	}

	for _, cmdDef := range commands {
		cmd := exec.CommandContext(ctx, "terraform", append([]string{cmdDef.Name}, cmdDef.Args...)...)
		cmd.Dir = r.WorkDir
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error in terraform %s: %w", cmdDef.Name, err)
		}
	}

	return nil
}