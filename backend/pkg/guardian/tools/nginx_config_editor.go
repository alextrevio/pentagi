package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NginxConfigEditor es una herramienta para editar archivos de configuración de Nginx
type NginxConfigEditor struct {
	backupDir string
}

// NginxConfigEditorArgs representa los argumentos para la edición
type NginxConfigEditorArgs struct {
	ConfigPath string            `json:"config_path"`
	Headers    map[string]string `json:"headers"`
	Backup     bool              `json:"backup"`
}

// NewNginxConfigEditor crea una nueva instancia del editor
func NewNginxConfigEditor(backupDir string) *NginxConfigEditor {
	if backupDir == "" {
		backupDir = "/tmp/nginx_backups"
	}
	
	// Crear directorio de backups si no existe
	os.MkdirAll(backupDir, 0755)
	
	return &NginxConfigEditor{
		backupDir: backupDir,
	}
}

// Name devuelve el nombre de la herramienta
func (n *NginxConfigEditor) Name() string {
	return "nginx_config_editor"
}

// Description devuelve la descripción de la herramienta
func (n *NginxConfigEditor) Description() string {
	return "Edita archivos de configuración de Nginx para añadir cabeceras de seguridad"
}

// IsAvailable indica si la herramienta está disponible
func (n *NginxConfigEditor) IsAvailable() bool {
	// Verificar si nginx está instalado
	_, err := exec.LookPath("nginx")
	return err == nil
}

// Execute ejecuta la edición del archivo de configuración
func (n *NginxConfigEditor) Execute(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var editArgs NginxConfigEditorArgs
	if err := json.Unmarshal(args, &editArgs); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if editArgs.ConfigPath == "" {
		return nil, fmt.Errorf("config_path is required")
	}

	if len(editArgs.Headers) == 0 {
		return nil, fmt.Errorf("at least one header is required")
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(editArgs.ConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", editArgs.ConfigPath)
	}

	// Crear backup si se solicita
	var backupPath string
	if editArgs.Backup {
		var err error
		backupPath, err = n.createBackup(editArgs.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Leer el archivo de configuración
	content, err := os.ReadFile(editArgs.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Añadir las cabeceras de seguridad
	modifiedContent, err := n.addSecurityHeaders(string(content), editArgs.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to add security headers: %w", err)
	}

	// Escribir el archivo modificado
	if err := os.WriteFile(editArgs.ConfigPath, []byte(modifiedContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	// Validar la sintaxis de Nginx
	if err := n.validateNginxConfig(); err != nil {
		// Si la validación falla, restaurar el backup
		if backupPath != "" {
			n.restoreBackup(backupPath, editArgs.ConfigPath)
		}
		return nil, fmt.Errorf("nginx config validation failed: %w", err)
	}

	result := map[string]interface{}{
		"success":     true,
		"message":     "Configuration updated successfully",
		"config_path": editArgs.ConfigPath,
		"backup_path": backupPath,
		"headers_added": len(editArgs.Headers),
	}

	return result, nil
}

// createBackup crea una copia de seguridad del archivo de configuración
func (n *NginxConfigEditor) createBackup(configPath string) (string, error) {
	// Generar nombre único para el backup
	timestamp := fmt.Sprintf("%d", os.Getpid())
	backupName := fmt.Sprintf("%s.backup.%s", filepath.Base(configPath), timestamp)
	backupPath := filepath.Join(n.backupDir, backupName)

	// Leer el archivo original
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Escribir el backup
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

// restoreBackup restaura un archivo desde un backup
func (n *NginxConfigEditor) restoreBackup(backupPath, configPath string) error {
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, content, 0644)
}

// addSecurityHeaders añade cabeceras de seguridad a la configuración de Nginx
func (n *NginxConfigEditor) addSecurityHeaders(content string, headers map[string]string) (string, error) {
	lines := strings.Split(content, "\n")
	modified := false

	// Buscar el bloque server o http
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Buscar el inicio de un bloque server
		if strings.HasPrefix(trimmed, "server") && strings.Contains(trimmed, "{") {
			// Insertar las cabeceras después de la línea del bloque server
			headerLines := make([]string, 0)
			for name, value := range headers {
				headerLines = append(headerLines, fmt.Sprintf("    add_header %s \"%s\" always;", name, value))
			}
			
			// Insertar las nuevas líneas
			newLines := make([]string, 0, len(lines)+len(headerLines))
			newLines = append(newLines, lines[:i+1]...)
			newLines = append(newLines, headerLines...)
			newLines = append(newLines, lines[i+1:]...)
			lines = newLines
			modified = true
			break
		}
	}

	if !modified {
		return "", fmt.Errorf("could not find server block in nginx config")
	}

	return strings.Join(lines, "\n"), nil
}

// validateNginxConfig valida la sintaxis de la configuración de Nginx
func (n *NginxConfigEditor) validateNginxConfig() error {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validation failed: %s", string(output))
	}
	return nil
}
