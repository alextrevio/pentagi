package graph

// This file contains GraphQL resolvers for Guardian auto-remediation functionality

import (
	"context"
	"encoding/json"
	"fmt"
	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/guardian"
	"time"

	"github.com/sirupsen/logrus"
)

// Guardian Task Resolvers

// GuardianTasks is the resolver for the guardianTasks query
func (r *queryResolver) GuardianTasks(ctx context.Context, flowID int64) ([]*model.GuardianTask, error) {
	_, err := validatePermissionWithFlowID(ctx, "guardian.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"flow_id": flowID,
	}).Debug("get guardian tasks")

	tasks, err := r.DB.GetGuardianTasks(ctx, flowID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.GuardianTask, len(tasks))
	for i, task := range tasks {
		result[i] = convertGuardianTask(task)
	}

	return result, nil
}

// GuardianTask is the resolver for the guardianTask query
func (r *queryResolver) GuardianTask(ctx context.Context, id int64) (*model.GuardianTask, error) {
	_, err := validatePermission(ctx, "guardian.view")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"task_id": id,
	}).Debug("get guardian task")

	task, err := r.DB.GetGuardianTask(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertGuardianTask(task), nil
}

// CreateGuardianTask is the resolver for the createGuardianTask mutation
func (r *mutationResolver) CreateGuardianTask(ctx context.Context, flowID int64, title string, targetURL string, vulnerabilityType model.VulnerabilityType) (*model.GuardianTask, error) {
	uid, err := validatePermissionWithFlowID(ctx, "guardian.create", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":                uid,
		"flow_id":            flowID,
		"title":              title,
		"target_url":         targetURL,
		"vulnerability_type": vulnerabilityType,
	}).Debug("create guardian task")

	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if targetURL == "" {
		return nil, fmt.Errorf("target URL is required")
	}

	task, err := r.DB.CreateGuardianTask(ctx, database.CreateGuardianTaskParams{
		Title:             title,
		TargetUrl:         targetURL,
		VulnerabilityType: database.VulnerabilityType(vulnerabilityType),
		Status:            database.RemediationStatusPendingApproval,
		FlowID:            flowID,
		TaskID:            database.NullInt64{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	return convertGuardianTask(task), nil
}

// Remediation Plan Resolvers

// RemediationPlans is the resolver for the remediationPlans query
func (r *queryResolver) RemediationPlans(ctx context.Context, flowID int64, status *model.RemediationStatus) ([]*model.RemediationPlan, error) {
	_, err := validatePermissionWithFlowID(ctx, "guardian.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"flow_id": flowID,
		"status":  status,
	}).Debug("get remediation plans")

	var plans []database.RemediationPlan
	if status != nil {
		plans, err = r.DB.GetRemediationPlansByStatus(ctx, database.GetRemediationPlansByStatusParams{
			FlowID: flowID,
			Status: database.RemediationStatus(*status),
		})
	} else {
		plans, err = r.DB.GetRemediationPlans(ctx, flowID)
	}
	if err != nil {
		return nil, err
	}

	result := make([]*model.RemediationPlan, len(plans))
	for i, plan := range plans {
		result[i], err = r.convertRemediationPlan(ctx, plan)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// RemediationPlan is the resolver for the remediationPlan query
func (r *queryResolver) RemediationPlan(ctx context.Context, id int64) (*model.RemediationPlan, error) {
	_, err := validatePermission(ctx, "guardian.view")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"plan_id": id,
	}).Debug("get remediation plan")

	plan, err := r.DB.GetRemediationPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertRemediationPlan(ctx, plan)
}

// CreateRemediationPlan is the resolver for the createRemediationPlan mutation
func (r *mutationResolver) CreateRemediationPlan(ctx context.Context, flowID int64, taskID *int64, vulnerabilityID string, vulnerabilityType model.VulnerabilityType, targetURL string) (*model.RemediationPlan, error) {
	uid, err := validatePermissionWithFlowID(ctx, "guardian.create", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":                uid,
		"flow_id":            flowID,
		"vulnerability_id":   vulnerabilityID,
		"vulnerability_type": vulnerabilityType,
	}).Debug("create remediation plan")

	// Use Guardian to create the plan
	guardianInstance := guardian.NewGuardian(nil)
	plan, err := guardianInstance.CreateRemediationPlan(ctx, vulnerabilityID, string(vulnerabilityType))
	if err != nil {
		return nil, err
	}

	// Save to database
	dbPlan, err := r.DB.CreateRemediationPlan(ctx, database.CreateRemediationPlanParams{
		VulnerabilityID:   plan.VulnerabilityID,
		VulnerabilityType: database.VulnerabilityType(vulnerabilityType),
		Description:       plan.Description,
		RiskLevel:         database.RiskLevel(plan.RiskLevel),
		EstimatedTime:     int32(plan.EstimatedTime),
		RequiresBackup:    plan.RequiresBackup,
		RequiresDowntime:  plan.RequiresDowntime,
		RollbackPlan:      plan.RollbackPlan,
		Status:            database.RemediationStatusPendingApproval,
		FlowID:            flowID,
		TaskID:            database.NullInt64{Valid: taskID != nil, Int64: *taskID},
		GuardianTaskID:    database.NullInt64{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	// Save steps
	for _, step := range plan.Steps {
		argsJSON, _ := json.Marshal(step.Arguments)
		_, err := r.DB.CreateRemediationStep(ctx, database.CreateRemediationStepParams{
			RemediationPlanID: dbPlan.ID,
			ToolName:          step.ToolName,
			Description:       step.Description,
			Arguments:         argsJSON,
			StepOrder:         int32(step.Order),
			RiskLevel:         database.RiskLevel(step.RiskLevel),
			Status:            database.RemediationStatusPendingApproval,
		})
		if err != nil {
			return nil, err
		}
	}

	return r.convertRemediationPlan(ctx, dbPlan)
}

// ApproveRemediationPlan is the resolver for the approveRemediationPlan mutation
func (r *mutationResolver) ApproveRemediationPlan(ctx context.Context, id int64, approvedBy string) (*model.RemediationPlan, error) {
	_, err := validatePermission(ctx, "guardian.approve")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"plan_id":     id,
		"approved_by": approvedBy,
	}).Debug("approve remediation plan")

	plan, err := r.DB.ApproveRemediationPlan(ctx, database.ApproveRemediationPlanParams{
		ApprovedBy: database.NullString{Valid: true, String: approvedBy},
		ID:         id,
	})
	if err != nil {
		return nil, err
	}

	return r.convertRemediationPlan(ctx, plan)
}

// RejectRemediationPlan is the resolver for the rejectRemediationPlan mutation
func (r *mutationResolver) RejectRemediationPlan(ctx context.Context, id int64, reason string) (*model.RemediationPlan, error) {
	_, err := validatePermission(ctx, "guardian.approve")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"plan_id": id,
		"reason":  reason,
	}).Debug("reject remediation plan")

	plan, err := r.DB.RejectRemediationPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertRemediationPlan(ctx, plan)
}

// ExecuteRemediationPlan is the resolver for the executeRemediationPlan mutation
func (r *mutationResolver) ExecuteRemediationPlan(ctx context.Context, id int64) (*model.RemediationPlan, error) {
	_, err := validatePermission(ctx, "guardian.execute")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"plan_id": id,
	}).Debug("execute remediation plan")

	// Get the plan
	dbPlan, err := r.DB.GetRemediationPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if approved
	if dbPlan.Status != database.RemediationStatusApproved {
		return nil, fmt.Errorf("plan must be approved before execution")
	}

	// Mark as in progress
	_, err = r.DB.StartRemediationPlanExecution(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get steps
	steps, err := r.DB.GetRemediationSteps(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert to Guardian plan
	guardianPlan := &guardian.RemediationPlan{
		ID:              fmt.Sprintf("%d", dbPlan.ID),
		VulnerabilityID: dbPlan.VulnerabilityID,
		Description:     dbPlan.Description,
		RiskLevel:       string(dbPlan.RiskLevel),
		EstimatedTime:   int(dbPlan.EstimatedTime),
		RequiresBackup:  dbPlan.RequiresBackup,
		RequiresDowntime: dbPlan.RequiresDowntime,
		RollbackPlan:    dbPlan.RollbackPlan,
		Status:          string(dbPlan.Status),
		Steps:           make([]guardian.RemediationStep, len(steps)),
	}

	for i, step := range steps {
		var args map[string]string
		json.Unmarshal(step.Arguments, &args)
		guardianPlan.Steps[i] = guardian.RemediationStep{
			ID:          int(step.ID),
			ToolName:    step.ToolName,
			Description: step.Description,
			Arguments:   args,
			Order:       int(step.StepOrder),
			RiskLevel:   string(step.RiskLevel),
		}
	}

	// Execute with Guardian
	guardianInstance := guardian.NewGuardian(&guardian.Config{
		RequireApproval: false, // Already approved
		DryRun:          false,
	})

	// Execute in background
	go func() {
		execErr := guardianInstance.ExecuteRemediationPlan(context.Background(), guardianPlan)
		if execErr != nil {
			r.Logger.WithError(execErr).Error("failed to execute remediation plan")
			r.DB.FailRemediationPlan(context.Background(), id)
		} else {
			r.DB.CompleteRemediationPlan(context.Background(), id)
		}
	}()

	// Return updated plan
	updatedPlan, err := r.DB.GetRemediationPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertRemediationPlan(ctx, updatedPlan)
}

// RollbackRemediation is the resolver for the rollbackRemediation mutation
func (r *mutationResolver) RollbackRemediation(ctx context.Context, planID int64, backupID string) (*model.RemediationPlan, error) {
	_, err := validatePermission(ctx, "guardian.rollback")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"plan_id":   planID,
		"backup_id": backupID,
	}).Debug("rollback remediation")

	// Get backup
	backup, err := r.DB.GetGuardianBackup(ctx, backupID)
	if err != nil {
		return nil, err
	}

	// TODO: Implement actual rollback logic using Guardian
	// For now, just mark as rolled back
	plan, err := r.DB.RollbackRemediationPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"backup_path": backup.BackupPath,
	}).Info("remediation rolled back")

	return r.convertRemediationPlan(ctx, plan)
}

// Backup Resolvers

// Backups is the resolver for the backups query
func (r *queryResolver) Backups(ctx context.Context, filePath *string) ([]*model.BackupInfo, error) {
	_, err := validatePermission(ctx, "guardian.view")
	if err != nil {
		return nil, err
	}

	r.Logger.Debug("get guardian backups")

	var backups []database.GuardianBackup
	if filePath != nil {
		backups, err = r.DB.GetGuardianBackupsByFilePath(ctx, *filePath)
	} else {
		backups, err = r.DB.GetGuardianBackups(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]*model.BackupInfo, len(backups))
	for i, backup := range backups {
		result[i] = convertBackupInfo(backup)
	}

	return result, nil
}

// CreateBackup is the resolver for the createBackup mutation
func (r *mutationResolver) CreateBackup(ctx context.Context, filePath string) (*model.BackupInfo, error) {
	_, err := validatePermission(ctx, "guardian.create")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"file_path": filePath,
	}).Debug("create backup")

	// TODO: Use Guardian FileBackupManager to create actual backup
	backupID := fmt.Sprintf("backup_%d", time.Now().Unix())
	backupPath := fmt.Sprintf("/tmp/guardian_backups/%s", backupID)

	backup, err := r.DB.CreateGuardianBackup(ctx, database.CreateGuardianBackupParams{
		ID:                 backupID,
		FilePath:           filePath,
		BackupPath:         backupPath,
		Size:               0, // TODO: Get actual size
		RemediationPlanID:  database.NullInt64{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	return convertBackupInfo(backup), nil
}

// DeleteBackup is the resolver for the deleteBackup mutation
func (r *mutationResolver) DeleteBackup(ctx context.Context, backupID string) (bool, error) {
	_, err := validatePermission(ctx, "guardian.admin")
	if err != nil {
		return false, err
	}

	r.Logger.WithFields(logrus.Fields{
		"backup_id": backupID,
	}).Debug("delete backup")

	err = r.DB.DeleteGuardianBackup(ctx, backupID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Helper functions

func convertGuardianTask(task database.GuardianTask) *model.GuardianTask {
	return &model.GuardianTask{
		ID:                fmt.Sprintf("%d", task.ID),
		Title:             task.Title,
		TargetURL:         task.TargetUrl,
		VulnerabilityType: model.VulnerabilityType(task.VulnerabilityType),
		Status:            model.StatusType(task.Status),
		FlowID:            fmt.Sprintf("%d", task.FlowID),
		TaskID:            nil, // TODO: Convert if valid
		CreatedAt:         task.CreatedAt.Time,
		UpdatedAt:         task.UpdatedAt.Time,
	}
}

func (r *queryResolver) convertRemediationPlan(ctx context.Context, plan database.RemediationPlan) (*model.RemediationPlan, error) {
	// Get steps
	steps, err := r.DB.GetRemediationSteps(ctx, plan.ID)
	if err != nil {
		return nil, err
	}

	modelSteps := make([]*model.RemediationStep, len(steps))
	for i, step := range steps {
		modelSteps[i] = convertRemediationStep(step)
	}

	return &model.RemediationPlan{
		ID:               fmt.Sprintf("%d", plan.ID),
		VulnerabilityID:  plan.VulnerabilityID,
		VulnerabilityType: model.VulnerabilityType(plan.VulnerabilityType),
		Description:      plan.Description,
		Steps:            modelSteps,
		RiskLevel:        model.RiskLevel(plan.RiskLevel),
		EstimatedTime:    int(plan.EstimatedTime),
		RequiresBackup:   plan.RequiresBackup,
		RequiresDowntime: plan.RequiresDowntime,
		RollbackPlan:     plan.RollbackPlan,
		Status:           model.RemediationStatus(plan.Status),
		FlowID:           fmt.Sprintf("%d", plan.FlowID),
		TaskID:           nil, // TODO: Convert if valid
		CreatedAt:        plan.CreatedAt.Time,
		UpdatedAt:        plan.UpdatedAt.Time,
		ApprovedBy:       &plan.ApprovedBy.String,
		ApprovedAt:       &plan.ApprovedAt.Time,
		ExecutedAt:       &plan.ExecutedAt.Time,
		CompletedAt:      &plan.CompletedAt.Time,
	}, nil
}

func convertRemediationStep(step database.RemediationStep) *model.RemediationStep {
	var args map[string]interface{}
	json.Unmarshal(step.Arguments, &args)

	return &model.RemediationStep{
		ID:          int(step.ID),
		ToolName:    step.ToolName,
		Description: step.Description,
		Arguments:   args,
		Order:       int(step.StepOrder),
		RiskLevel:   model.RiskLevel(step.RiskLevel),
		Status:      model.RemediationStatus(step.Status),
		Result:      &step.Result.String,
		ExecutedAt:  &step.ExecutedAt.Time,
	}
}

func convertBackupInfo(backup database.GuardianBackup) *model.BackupInfo {
	return &model.BackupInfo{
		ID:         backup.ID,
		FilePath:   backup.FilePath,
		BackupPath: backup.BackupPath,
		CreatedAt:  backup.CreatedAt.Time,
		Size:       int(backup.Size),
	}
}
