package guardian

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// MockTool es un mock de herramienta para testing
type MockTool struct {
	name        string
	description string
	available   bool
	shouldFail  bool
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	if m.shouldFail {
		return &ToolResult{
			Success: false,
			Error:   "mock tool failed",
		}, nil
	}

	return &ToolResult{
		Success: true,
		Message: "mock tool executed successfully",
		Data:    map[string]string{"result": "success"},
	}, nil
}

func (m *MockTool) IsAvailable() bool {
	return m.available
}

func TestNewGuardian(t *testing.T) {
	t.Run("creates guardian with default config", func(t *testing.T) {
		guardian := NewGuardian(nil)

		if guardian == nil {
			t.Fatal("expected guardian to be created")
		}

		if !guardian.config.RequireApproval {
			t.Error("expected RequireApproval to be true by default")
		}

		if guardian.config.MaxRetries != 3 {
			t.Errorf("expected MaxRetries to be 3, got %d", guardian.config.MaxRetries)
		}
	})

	t.Run("creates guardian with custom config", func(t *testing.T) {
		config := &Config{
			RequireApproval: false,
			DryRun:          true,
			MaxRetries:      5,
			Timeout:         10 * time.Minute,
		}

		guardian := NewGuardian(config)

		if guardian.config.RequireApproval {
			t.Error("expected RequireApproval to be false")
		}

		if !guardian.config.DryRun {
			t.Error("expected DryRun to be true")
		}

		if guardian.config.MaxRetries != 5 {
			t.Errorf("expected MaxRetries to be 5, got %d", guardian.config.MaxRetries)
		}
	})
}

func TestRegisterTool(t *testing.T) {
	guardian := NewGuardian(nil)

	t.Run("registers tool successfully", func(t *testing.T) {
		tool := &MockTool{
			name:        "test_tool",
			description: "A test tool",
			available:   true,
		}

		err := guardian.RegisterTool(tool)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(guardian.tools) != 1 {
			t.Errorf("expected 1 tool, got %d", len(guardian.tools))
		}
	})

	t.Run("fails to register nil tool", func(t *testing.T) {
		err := guardian.RegisterTool(nil)
		if err == nil {
			t.Error("expected error when registering nil tool")
		}
	})

	t.Run("fails to register duplicate tool", func(t *testing.T) {
		tool1 := &MockTool{name: "duplicate", available: true}
		tool2 := &MockTool{name: "duplicate", available: true}

		guardian := NewGuardian(nil)
		guardian.RegisterTool(tool1)

		err := guardian.RegisterTool(tool2)
		if err == nil {
			t.Error("expected error when registering duplicate tool")
		}
	})
}

func TestGetAvailableTools(t *testing.T) {
	guardian := NewGuardian(nil)

	tool1 := &MockTool{name: "available_tool", available: true}
	tool2 := &MockTool{name: "unavailable_tool", available: false}
	tool3 := &MockTool{name: "another_available", available: true}

	guardian.RegisterTool(tool1)
	guardian.RegisterTool(tool2)
	guardian.RegisterTool(tool3)

	available := guardian.GetAvailableTools()

	if len(available) != 2 {
		t.Errorf("expected 2 available tools, got %d", len(available))
	}

	for _, tool := range available {
		if !tool.IsAvailable() {
			t.Error("expected all returned tools to be available")
		}
	}
}

func TestCreateRemediationPlan(t *testing.T) {
	guardian := NewGuardian(nil)
	ctx := context.Background()

	t.Run("creates plan successfully", func(t *testing.T) {
		plan, err := guardian.CreateRemediationPlan(ctx, "vuln-001", "missing_security_headers")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if plan == nil {
			t.Fatal("expected plan to be created")
		}

		if plan.VulnerabilityID != "vuln-001" {
			t.Errorf("expected vulnerability ID to be vuln-001, got %s", plan.VulnerabilityID)
		}

		if plan.Status != "pending_approval" {
			t.Errorf("expected status to be pending_approval, got %s", plan.Status)
		}

		if !plan.RequiresBackup {
			t.Error("expected RequiresBackup to be true")
		}
	})
}

func TestExecuteRemediationPlan(t *testing.T) {
	guardian := NewGuardian(nil)
	ctx := context.Background()

	t.Run("fails without approval", func(t *testing.T) {
		plan := &RemediationPlan{
			ID:              "plan-001",
			VulnerabilityID: "vuln-001",
			Status:          "pending_approval",
			Steps:           []RemediationStep{},
		}

		err := guardian.ExecuteRemediationPlan(ctx, plan)

		if err == nil {
			t.Error("expected error when executing plan without approval")
		}
	})

	t.Run("executes in dry run mode", func(t *testing.T) {
		config := &Config{
			RequireApproval: false,
			DryRun:          true,
		}
		guardian := NewGuardian(config)

		plan := &RemediationPlan{
			ID:              "plan-001",
			VulnerabilityID: "vuln-001",
			Status:          "approved",
			Steps:           []RemediationStep{},
		}

		err := guardian.ExecuteRemediationPlan(ctx, plan)

		if err != nil {
			t.Fatalf("expected no error in dry run, got %v", err)
		}

		if plan.Status != "simulated" {
			t.Errorf("expected status to be simulated, got %s", plan.Status)
		}
	})

	t.Run("executes plan successfully", func(t *testing.T) {
		config := &Config{
			RequireApproval: false,
			DryRun:          false,
		}
		guardian := NewGuardian(config)

		tool := &MockTool{
			name:      "test_tool",
			available: true,
		}
		guardian.RegisterTool(tool)

		plan := &RemediationPlan{
			ID:              "plan-001",
			VulnerabilityID: "vuln-001",
			Status:          "approved",
			Steps: []RemediationStep{
				{
					ID:          1,
					ToolName:    "test_tool",
					Description: "Test step",
					Arguments:   map[string]string{"arg1": "value1"},
					Order:       1,
				},
			},
		}

		err := guardian.ExecuteRemediationPlan(ctx, plan)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if plan.Status != "completed" {
			t.Errorf("expected status to be completed, got %s", plan.Status)
		}
	})

	t.Run("fails when tool not found", func(t *testing.T) {
		config := &Config{
			RequireApproval: false,
			DryRun:          false,
		}
		guardian := NewGuardian(config)

		plan := &RemediationPlan{
			ID:              "plan-001",
			VulnerabilityID: "vuln-001",
			Status:          "approved",
			Steps: []RemediationStep{
				{
					ID:          1,
					ToolName:    "nonexistent_tool",
					Description: "Test step",
					Arguments:   map[string]string{},
					Order:       1,
				},
			},
		}

		err := guardian.ExecuteRemediationPlan(ctx, plan)

		if err == nil {
			t.Error("expected error when tool not found")
		}
	})

	t.Run("fails when tool execution fails", func(t *testing.T) {
		config := &Config{
			RequireApproval: false,
			DryRun:          false,
		}
		guardian := NewGuardian(config)

		tool := &MockTool{
			name:       "failing_tool",
			available:  true,
			shouldFail: true,
		}
		guardian.RegisterTool(tool)

		plan := &RemediationPlan{
			ID:              "plan-001",
			VulnerabilityID: "vuln-001",
			Status:          "approved",
			Steps: []RemediationStep{
				{
					ID:          1,
					ToolName:    "failing_tool",
					Description: "Test step",
					Arguments:   map[string]string{},
					Order:       1,
				},
			},
		}

		err := guardian.ExecuteRemediationPlan(ctx, plan)

		if err == nil {
			t.Error("expected error when tool execution fails")
		}
	})
}
