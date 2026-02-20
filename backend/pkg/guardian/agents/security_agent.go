package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SecurityAgent es un agente especializado en correcciones de seguridad
type SecurityAgent struct {
	config *SecurityAgentConfig
}

// SecurityAgentConfig contiene la configuración del agente de seguridad
type SecurityAgentConfig struct {
	MaxConcurrentTasks int
	DefaultTimeout     time.Duration
	EnableAutoApproval bool // Solo para tareas de bajo riesgo
}

// VulnerabilityAnalysis representa el análisis de una vulnerabilidad
type VulnerabilityAnalysis struct {
	VulnerabilityID   string            `json:"vulnerability_id"`
	Type              string            `json:"type"`
	Severity          string            `json:"severity"`
	Description       string            `json:"description"`
	AffectedResources []string          `json:"affected_resources"`
	Recommendations   []string          `json:"recommendations"`
	RemediationSteps  []RemediationStep `json:"remediation_steps"`
	EstimatedImpact   string            `json:"estimated_impact"`
	RequiresDowntime  bool              `json:"requires_downtime"`
}

// RemediationStep representa un paso de corrección
type RemediationStep struct {
	ID          int               `json:"id"`
	ToolName    string            `json:"tool_name"`
	Description string            `json:"description"`
	Arguments   map[string]string `json:"arguments"`
	Order       int               `json:"order"`
	RiskLevel   string            `json:"risk_level"`
}

// NewSecurityAgent crea una nueva instancia del agente de seguridad
func NewSecurityAgent(config *SecurityAgentConfig) *SecurityAgent {
	if config == nil {
		config = &SecurityAgentConfig{
			MaxConcurrentTasks: 5,
			DefaultTimeout:     30 * time.Minute,
			EnableAutoApproval: false,
		}
	}

	return &SecurityAgent{
		config: config,
	}
}

// AnalyzeVulnerability analiza una vulnerabilidad y determina el plan de corrección
func (s *SecurityAgent) AnalyzeVulnerability(ctx context.Context, vulnerabilityType string, targetURL string) (*VulnerabilityAnalysis, error) {
	// Esta es una implementación básica que se expandirá con IA en el futuro
	analysis := &VulnerabilityAnalysis{
		VulnerabilityID:   fmt.Sprintf("vuln-%d", time.Now().Unix()),
		Type:              vulnerabilityType,
		AffectedResources: []string{targetURL},
		Recommendations:   make([]string, 0),
		RemediationSteps:  make([]RemediationStep, 0),
	}

	// Determinar la severidad y los pasos según el tipo de vulnerabilidad
	switch vulnerabilityType {
	case "missing_security_headers":
		analysis.Severity = "medium"
		analysis.Description = "El servidor web no incluye cabeceras de seguridad HTTP recomendadas"
		analysis.EstimatedImpact = "Bajo - Solo requiere recarga del servicio web"
		analysis.RequiresDowntime = false
		
		analysis.Recommendations = []string{
			"Añadir Content-Security-Policy para prevenir XSS",
			"Añadir Strict-Transport-Security para forzar HTTPS",
			"Añadir X-Frame-Options para prevenir clickjacking",
			"Añadir X-Content-Type-Options para prevenir MIME sniffing",
		}

		analysis.RemediationSteps = []RemediationStep{
			{
				ID:          1,
				ToolName:    "http_header_scanner",
				Description: "Escanear URL para identificar cabeceras faltantes",
				Arguments: map[string]string{
					"url": targetURL,
				},
				Order:     1,
				RiskLevel: "low",
			},
			{
				ID:          2,
				ToolName:    "file_backup_manager",
				Description: "Crear backup de la configuración actual",
				Arguments: map[string]string{
					"file_path": "/etc/nginx/nginx.conf",
					"action":    "create",
				},
				Order:     2,
				RiskLevel: "low",
			},
			{
				ID:          3,
				ToolName:    "nginx_config_editor",
				Description: "Añadir cabeceras de seguridad a la configuración de Nginx",
				Arguments: map[string]string{
					"config_path": "/etc/nginx/nginx.conf",
					"backup":      "true",
				},
				Order:     3,
				RiskLevel: "medium",
			},
			{
				ID:          4,
				ToolName:    "docker_service_manager",
				Description: "Recargar el servicio Nginx",
				Arguments: map[string]string{
					"container_id": "nginx-container",
					"action":       "reload",
					"service":      "nginx",
				},
				Order:     4,
				RiskLevel: "low",
			},
			{
				ID:          5,
				ToolName:    "http_header_scanner",
				Description: "Verificar que las cabeceras se añadieron correctamente",
				Arguments: map[string]string{
					"url": targetURL,
				},
				Order:     5,
				RiskLevel: "low",
			},
		}

	case "weak_ssl_configuration":
		analysis.Severity = "high"
		analysis.Description = "El servidor utiliza configuración SSL/TLS débil o desactualizada"
		analysis.EstimatedImpact = "Medio - Requiere reinicio del servicio web"
		analysis.RequiresDowntime = true
		
		analysis.Recommendations = []string{
			"Actualizar a TLS 1.2 o superior",
			"Deshabilitar cifrados débiles",
			"Configurar Perfect Forward Secrecy",
		}

	case "outdated_software":
		analysis.Severity = "high"
		analysis.Description = "El software del servidor está desactualizado y contiene vulnerabilidades conocidas"
		analysis.EstimatedImpact = "Alto - Requiere actualización de paquetes y posible reinicio"
		analysis.RequiresDowntime = true
		
		analysis.Recommendations = []string{
			"Actualizar paquetes del sistema",
			"Aplicar parches de seguridad",
			"Reiniciar servicios afectados",
		}

	default:
		return nil, fmt.Errorf("unknown vulnerability type: %s", vulnerabilityType)
	}

	return analysis, nil
}

// DetermineRiskLevel determina el nivel de riesgo de una corrección
func (s *SecurityAgent) DetermineRiskLevel(analysis *VulnerabilityAnalysis) string {
	// Lógica simple para determinar el riesgo
	if analysis.RequiresDowntime {
		return "high"
	}

	if analysis.Severity == "critical" || analysis.Severity == "high" {
		return "medium"
	}

	return "low"
}

// CanAutoApprove determina si una corrección puede ser auto-aprobada
func (s *SecurityAgent) CanAutoApprove(analysis *VulnerabilityAnalysis) bool {
	if !s.config.EnableAutoApproval {
		return false
	}

	// Solo auto-aprobar correcciones de bajo riesgo
	riskLevel := s.DetermineRiskLevel(analysis)
	return riskLevel == "low" && !analysis.RequiresDowntime
}

// GenerateRollbackPlan genera un plan de reversión
func (s *SecurityAgent) GenerateRollbackPlan(analysis *VulnerabilityAnalysis) string {
	rollbackSteps := []string{
		"1. Detener la ejecución de pasos adicionales",
		"2. Restaurar archivos desde el backup más reciente",
		"3. Reiniciar servicios afectados",
		"4. Verificar que el sistema vuelve a su estado anterior",
		"5. Notificar al administrador del fallo",
	}

	rollbackPlan := "Plan de Reversión:\n"
	for _, step := range rollbackSteps {
		rollbackPlan += step + "\n"
	}

	return rollbackPlan
}
