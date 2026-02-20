package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// DockerServiceManager gestiona servicios en contenedores Docker
type DockerServiceManager struct {
	client *client.Client
}

// DockerServiceManagerArgs representa los argumentos para gestionar servicios
type DockerServiceManagerArgs struct {
	ContainerID string `json:"container_id"`
	Action      string `json:"action"` // reload, restart, stop, start
	Service     string `json:"service"` // nginx, apache, etc.
}

// NewDockerServiceManager crea una nueva instancia del gestor
func NewDockerServiceManager() (*DockerServiceManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerServiceManager{
		client: cli,
	}, nil
}

// Name devuelve el nombre de la herramienta
func (d *DockerServiceManager) Name() string {
	return "docker_service_manager"
}

// Description devuelve la descripción de la herramienta
func (d *DockerServiceManager) Description() string {
	return "Gestiona servicios en contenedores Docker (reload, restart, stop, start)"
}

// IsAvailable indica si la herramienta está disponible
func (d *DockerServiceManager) IsAvailable() bool {
	if d.client == nil {
		return false
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := d.client.Ping(ctx)
	return err == nil
}

// Execute ejecuta una acción sobre un servicio en un contenedor
func (d *DockerServiceManager) Execute(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var serviceArgs DockerServiceManagerArgs
	if err := json.Unmarshal(args, &serviceArgs); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if serviceArgs.ContainerID == "" {
		return nil, fmt.Errorf("container_id is required")
	}

	if serviceArgs.Action == "" {
		return nil, fmt.Errorf("action is required")
	}

	if serviceArgs.Service == "" {
		return nil, fmt.Errorf("service is required")
	}

	// Ejecutar la acción correspondiente
	var err error
	var message string

	switch serviceArgs.Action {
	case "reload":
		message, err = d.reloadService(ctx, serviceArgs.ContainerID, serviceArgs.Service)
	case "restart":
		message, err = d.restartService(ctx, serviceArgs.ContainerID, serviceArgs.Service)
	case "stop":
		message, err = d.stopService(ctx, serviceArgs.ContainerID, serviceArgs.Service)
	case "start":
		message, err = d.startService(ctx, serviceArgs.ContainerID, serviceArgs.Service)
	default:
		return nil, fmt.Errorf("unknown action: %s", serviceArgs.Action)
	}

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"success":      true,
		"message":      message,
		"container_id": serviceArgs.ContainerID,
		"service":      serviceArgs.Service,
		"action":       serviceArgs.Action,
	}

	return result, nil
}

// reloadService recarga un servicio sin reiniciar el contenedor
func (d *DockerServiceManager) reloadService(ctx context.Context, containerID, service string) (string, error) {
	var cmd []string
	
	switch service {
	case "nginx":
		cmd = []string{"nginx", "-s", "reload"}
	case "apache", "apache2", "httpd":
		cmd = []string{"apachectl", "graceful"}
	default:
		return "", fmt.Errorf("unsupported service for reload: %s", service)
	}

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := d.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	if err := d.client.ContainerExecStart(ctx, execID.ID, container.ExecStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start exec: %w", err)
	}

	// Esperar a que termine la ejecución
	for {
		inspect, err := d.client.ContainerExecInspect(ctx, execID.ID)
		if err != nil {
			return "", fmt.Errorf("failed to inspect exec: %w", err)
		}

		if !inspect.Running {
			if inspect.ExitCode != 0 {
				return "", fmt.Errorf("exec failed with exit code %d", inspect.ExitCode)
			}
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Sprintf("Service %s reloaded successfully", service), nil
}

// restartService reinicia un servicio dentro del contenedor
func (d *DockerServiceManager) restartService(ctx context.Context, containerID, service string) (string, error) {
	// Para servicios, primero intentamos un reload suave
	message, err := d.reloadService(ctx, containerID, service)
	if err == nil {
		return message, nil
	}

	// Si el reload falla, reiniciamos el contenedor completo
	timeout := int(30)
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	if err := d.client.ContainerRestart(ctx, containerID, stopOptions); err != nil {
		return "", fmt.Errorf("failed to restart container: %w", err)
	}

	return fmt.Sprintf("Container %s restarted successfully", containerID), nil
}

// stopService detiene un servicio
func (d *DockerServiceManager) stopService(ctx context.Context, containerID, service string) (string, error) {
	timeout := int(30)
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	if err := d.client.ContainerStop(ctx, containerID, stopOptions); err != nil {
		return "", fmt.Errorf("failed to stop container: %w", err)
	}

	return fmt.Sprintf("Container %s stopped successfully", containerID), nil
}

// startService inicia un servicio
func (d *DockerServiceManager) startService(ctx context.Context, containerID, service string) (string, error) {
	if err := d.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return fmt.Sprintf("Container %s started successfully", containerID), nil
}

// Close cierra la conexión con Docker
func (d *DockerServiceManager) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}
