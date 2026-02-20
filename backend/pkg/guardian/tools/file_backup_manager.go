package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileBackupManager gestiona backups de archivos
type FileBackupManager struct {
	backupDir string
}

// FileBackupManagerArgs representa los argumentos para gestionar backups
type FileBackupManagerArgs struct {
	FilePath string `json:"file_path"`
	Action   string `json:"action"` // create, restore, list, delete
	BackupID string `json:"backup_id,omitempty"`
}

// BackupInfo contiene información sobre un backup
type BackupInfo struct {
	ID        string    `json:"id"`
	FilePath  string    `json:"file_path"`
	BackupPath string   `json:"backup_path"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
}

// NewFileBackupManager crea una nueva instancia del gestor de backups
func NewFileBackupManager(backupDir string) *FileBackupManager {
	if backupDir == "" {
		backupDir = "/tmp/guardian_backups"
	}
	
	// Crear directorio de backups si no existe
	os.MkdirAll(backupDir, 0755)
	
	return &FileBackupManager{
		backupDir: backupDir,
	}
}

// Name devuelve el nombre de la herramienta
func (f *FileBackupManager) Name() string {
	return "file_backup_manager"
}

// Description devuelve la descripción de la herramienta
func (f *FileBackupManager) Description() string {
	return "Gestiona backups de archivos antes de modificaciones (create, restore, list, delete)"
}

// IsAvailable indica si la herramienta está disponible
func (f *FileBackupManager) IsAvailable() bool {
	// Verificar que el directorio de backups existe y es escribible
	_, err := os.Stat(f.backupDir)
	return err == nil
}

// Execute ejecuta una acción de gestión de backups
func (f *FileBackupManager) Execute(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var backupArgs FileBackupManagerArgs
	if err := json.Unmarshal(args, &backupArgs); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if backupArgs.Action == "" {
		return nil, fmt.Errorf("action is required")
	}

	var result map[string]interface{}
	var err error

	switch backupArgs.Action {
	case "create":
		if backupArgs.FilePath == "" {
			return nil, fmt.Errorf("file_path is required for create action")
		}
		result, err = f.createBackup(backupArgs.FilePath)
	case "restore":
		if backupArgs.BackupID == "" {
			return nil, fmt.Errorf("backup_id is required for restore action")
		}
		result, err = f.restoreBackup(backupArgs.BackupID)
	case "list":
		result, err = f.listBackups(backupArgs.FilePath)
	case "delete":
		if backupArgs.BackupID == "" {
			return nil, fmt.Errorf("backup_id is required for delete action")
		}
		result, err = f.deleteBackup(backupArgs.BackupID)
	default:
		return nil, fmt.Errorf("unknown action: %s", backupArgs.Action)
	}

	if err != nil {
		return nil, err
	}

	result["success"] = true
	return result, nil
}

// createBackup crea un backup de un archivo
func (f *FileBackupManager) createBackup(filePath string) (map[string]interface{}, error) {
	// Verificar que el archivo existe
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}

	// Generar ID único para el backup
	timestamp := time.Now().Unix()
	backupID := fmt.Sprintf("%s_%d", filepath.Base(filePath), timestamp)
	backupPath := filepath.Join(f.backupDir, backupID)

	// Copiar el archivo
	if err := f.copyFile(filePath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Guardar metadata del backup
	metadata := BackupInfo{
		ID:         backupID,
		FilePath:   filePath,
		BackupPath: backupPath,
		CreatedAt:  time.Now(),
		Size:       fileInfo.Size(),
	}

	metadataPath := backupPath + ".meta.json"
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	result := map[string]interface{}{
		"message":     "Backup created successfully",
		"backup_id":   backupID,
		"backup_path": backupPath,
		"file_path":   filePath,
		"size":        fileInfo.Size(),
		"created_at":  metadata.CreatedAt,
	}

	return result, nil
}

// restoreBackup restaura un archivo desde un backup
func (f *FileBackupManager) restoreBackup(backupID string) (map[string]interface{}, error) {
	backupPath := filepath.Join(f.backupDir, backupID)
	metadataPath := backupPath + ".meta.json"

	// Leer metadata
	metadataJSON, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %w", err)
	}

	var metadata BackupInfo
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Verificar que el backup existe
	if _, err := os.Stat(backupPath); err != nil {
		return nil, fmt.Errorf("backup file not found: %w", err)
	}

	// Restaurar el archivo
	if err := f.copyFile(backupPath, metadata.FilePath); err != nil {
		return nil, fmt.Errorf("failed to restore backup: %w", err)
	}

	result := map[string]interface{}{
		"message":     "Backup restored successfully",
		"backup_id":   backupID,
		"file_path":   metadata.FilePath,
		"restored_at": time.Now(),
	}

	return result, nil
}

// listBackups lista todos los backups disponibles
func (f *FileBackupManager) listBackups(filePath string) (map[string]interface{}, error) {
	files, err := os.ReadDir(f.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	backups := make([]BackupInfo, 0)

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			// Leer metadata
			metadataPath := filepath.Join(f.backupDir, file.Name())
			metadataJSON, err := os.ReadFile(metadataPath)
			if err != nil {
				continue
			}

			var metadata BackupInfo
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				continue
			}

			// Filtrar por filePath si se especificó
			if filePath != "" && metadata.FilePath != filePath {
				continue
			}

			backups = append(backups, metadata)
		}
	}

	result := map[string]interface{}{
		"message": fmt.Sprintf("Found %d backups", len(backups)),
		"backups": backups,
		"count":   len(backups),
	}

	return result, nil
}

// deleteBackup elimina un backup
func (f *FileBackupManager) deleteBackup(backupID string) (map[string]interface{}, error) {
	backupPath := filepath.Join(f.backupDir, backupID)
	metadataPath := backupPath + ".meta.json"

	// Eliminar archivo de backup
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to delete backup file: %w", err)
	}

	// Eliminar metadata
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to delete metadata file: %w", err)
	}

	result := map[string]interface{}{
		"message":   "Backup deleted successfully",
		"backup_id": backupID,
	}

	return result, nil
}

// copyFile copia un archivo de origen a destino
func (f *FileBackupManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copiar permisos
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
