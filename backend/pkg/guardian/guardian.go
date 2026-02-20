package guardian

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Guardian representa el agente de auto-remediación de seguridad
type Guardian struct {
	config *Config
	tools  []Tool
}

// Config contiene la configuración del Guardian
type Config struct {
	// RequireApproval indica si se requiere aprobación humana antes de ejecutar correcciones
	RequireApproval bool
	// DryRun indica si solo se debe simular la corrección sin aplicarla
	DryRun bool
	// MaxRetries es el número máximo de reintentos en caso de fallo
	MaxRetries int
	// Timeout es el tiempo máximo de espera para una corrección
	Timeout time.Duration
}

// Tool representa una herramienta que el Guardian puede usar
type Tool interface {
	// Name devuelve el nombre de la herramienta
	Name() string
	// Description devuelve la descripción de la herramienta
	Description() string
	// Execute ejecuta la herramienta con los argumentos proporcionados
	Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
	// IsAvailable indica si la herramienta está disponible para su uso
	IsAvailable() bool
}

// ToolResult representa el resultado de la ejecución de una herramienta
type ToolResult struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// RemediationPlan representa un plan de corrección propuesto por el Guardian
type RemediationPlan struct {
	ID              string            `json:"id"`
	VulnerabilityID string            `json:"vulnerability_id"`
	Description     string            `json:"description"`
	Steps           []RemediationStep `json:"steps"`
	RiskLevel       string            `json:"risk_level"`
	EstimatedTime   time.Duration     `json:"estimated_time"`
	RequiresBackup  bool              `json:"requires_backup"`
	RollbackPlan    string            `json:"rollback_plan"`
	CreatedAt       time.Time         `json:"created_at"`
	Status          string            `json:"status"`
}

// RemediationStep representa un paso individual en un plan de corrección
type RemediationStep struct {
	ID          int               `json:"id"`
	ToolName    string            `json:"tool_name"`
	Description string            `json:"description"`
	Arguments   map[string]string `json:"arguments"`
	Order       int               `json:"order"`
}

// NewGuardian crea una nueva instancia del Guardian
func NewGuardian(config *Config) *Guardian {
	if config == nil {
		config = &Config{
			RequireApproval: true, // Por defecto, siempre requerir aprobación
			DryRun:          false,
			MaxRetries:      3,
			Timeout:         30 * time.Minute,
		}
	}

	return &Guardian{
		config: config,
		tools:  make([]Tool, 0),
	}
}

// RegisterTool registra una nueva herramienta en el Guardian
func (g *Guardian) RegisterTool(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	// Verificar que no exista una herramienta con el mismo nombre
	for _, t := range g.tools {
		if t.Name() == tool.Name() {
			return fmt.Errorf("tool with name %s already registered", tool.Name())
		}
	}

	g.tools = append(g.tools, tool)
	return nil
}

// GetAvailableTools devuelve una lista de herramientas disponibles
func (g *Guardian) GetAvailableTools() []Tool {
	available := make([]Tool, 0)
	for _, tool := range g.tools {
		if tool.IsAvailable() {
			available = append(available, tool)
		}
	}
	return available
}

// CreateRemediationPlan crea un plan de corrección para una vulnerabilidad
func (g *Guardian) CreateRemediationPlan(ctx context.Context, vulnerabilityID string, vulnerabilityType string) (*RemediationPlan, error) {
	// Esta es una implementación básica que se expandirá en fases posteriores
	plan := &RemediationPlan{
		ID:              fmt.Sprintf("plan-%d", time.Now().Unix()),
		VulnerabilityID: vulnerabilityID,
		Description:     fmt.Sprintf("Plan de corrección para vulnerabilidad %s", vulnerabilityType),
		Steps:           make([]RemediationStep, 0),
		RiskLevel:       "medium",
		EstimatedTime:   5 * time.Minute,
		RequiresBackup:  true,
		RollbackPlan:    "Restaurar desde backup automático",
		CreatedAt:       time.Now(),
		Status:          "pending_approval",
	}

	return plan, nil
}

// ExecuteRemediationPlan ejecuta un plan de corrección
func (g *Guardian) ExecuteRemediationPlan(ctx context.Context, plan *RemediationPlan) error {
	if plan == nil {
		return fmt.Errorf("remediation plan cannot be nil")
	}

	// Verificar si se requiere aprobación
	if g.config.RequireApproval && plan.Status != "approved" {
		return fmt.Errorf("remediation plan requires approval before execution")
	}

	// Si está en modo DryRun, solo simular
	if g.config.DryRun {
		plan.Status = "simulated"
		return nil
	}

	// Ejecutar cada paso del plan
	for _, step := range plan.Steps {
		// Buscar la herramienta correspondiente
		tool := g.findToolByName(step.ToolName)
		if tool == nil {
			return fmt.Errorf("tool %s not found", step.ToolName)
		}

		// Convertir argumentos a JSON
		argsJSON, err := json.Marshal(step.Arguments)
		if err != nil {
			return fmt.Errorf("failed to marshal arguments: %w", err)
		}

		// Ejecutar la herramienta
		result, err := tool.Execute(ctx, argsJSON)
		if err != nil {
			return fmt.Errorf("failed to execute tool %s: %w", step.ToolName, err)
		}

		if !result.Success {
			return fmt.Errorf("tool %s failed: %s", step.ToolName, result.Error)
		}
	}

	plan.Status = "completed"
	return nil
}

// findToolByName busca una herramienta por su nombre
func (g *Guardian) findToolByName(name string) Tool {
	for _, tool := range g.tools {
		if tool.Name() == name {
			return tool
		}
	}
	return nil
}
