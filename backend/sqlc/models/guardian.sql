-- Guardian Tasks Queries

-- name: GetGuardianTasks :many
SELECT
  gt.*
FROM guardian_tasks gt
INNER JOIN flows f ON gt.flow_id = f.id
WHERE gt.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY gt.created_at DESC;

-- name: GetGuardianTask :one
SELECT
  gt.*
FROM guardian_tasks gt
INNER JOIN flows f ON gt.flow_id = f.id
WHERE gt.id = $1 AND f.deleted_at IS NULL;

-- name: GetGuardianTasksByStatus :many
SELECT
  gt.*
FROM guardian_tasks gt
INNER JOIN flows f ON gt.flow_id = f.id
WHERE gt.flow_id = $1 AND gt.status = $2 AND f.deleted_at IS NULL
ORDER BY gt.created_at DESC;

-- name: CreateGuardianTask :one
INSERT INTO guardian_tasks (
  title,
  target_url,
  vulnerability_type,
  status,
  flow_id,
  task_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateGuardianTaskStatus :one
UPDATE guardian_tasks
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: DeleteGuardianTask :exec
DELETE FROM guardian_tasks
WHERE id = $1;

-- Remediation Plans Queries

-- name: GetRemediationPlans :many
SELECT
  rp.*
FROM remediation_plans rp
INNER JOIN flows f ON rp.flow_id = f.id
WHERE rp.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY rp.created_at DESC;

-- name: GetRemediationPlan :one
SELECT
  rp.*
FROM remediation_plans rp
INNER JOIN flows f ON rp.flow_id = f.id
WHERE rp.id = $1 AND f.deleted_at IS NULL;

-- name: GetRemediationPlansByStatus :many
SELECT
  rp.*
FROM remediation_plans rp
INNER JOIN flows f ON rp.flow_id = f.id
WHERE rp.flow_id = $1 AND rp.status = $2 AND f.deleted_at IS NULL
ORDER BY rp.created_at DESC;

-- name: GetRemediationPlansByGuardianTask :many
SELECT
  rp.*
FROM remediation_plans rp
WHERE rp.guardian_task_id = $1
ORDER BY rp.created_at DESC;

-- name: CreateRemediationPlan :one
INSERT INTO remediation_plans (
  vulnerability_id,
  vulnerability_type,
  description,
  risk_level,
  estimated_time,
  requires_backup,
  requires_downtime,
  rollback_plan,
  status,
  flow_id,
  task_id,
  guardian_task_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: UpdateRemediationPlanStatus :one
UPDATE remediation_plans
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: ApproveRemediationPlan :one
UPDATE remediation_plans
SET 
  status = 'approved',
  approved_by = $1,
  approved_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: RejectRemediationPlan :one
UPDATE remediation_plans
SET 
  status = 'rejected',
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: StartRemediationPlanExecution :one
UPDATE remediation_plans
SET 
  status = 'in_progress',
  executed_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CompleteRemediationPlan :one
UPDATE remediation_plans
SET 
  status = 'completed',
  completed_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: FailRemediationPlan :one
UPDATE remediation_plans
SET 
  status = 'failed',
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: RollbackRemediationPlan :one
UPDATE remediation_plans
SET 
  status = 'rolled_back',
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRemediationPlan :exec
DELETE FROM remediation_plans
WHERE id = $1;

-- Remediation Steps Queries

-- name: GetRemediationSteps :many
SELECT
  rs.*
FROM remediation_steps rs
WHERE rs.remediation_plan_id = $1
ORDER BY rs.step_order ASC;

-- name: GetRemediationStep :one
SELECT
  rs.*
FROM remediation_steps rs
WHERE rs.id = $1;

-- name: CreateRemediationStep :one
INSERT INTO remediation_steps (
  remediation_plan_id,
  tool_name,
  description,
  arguments,
  step_order,
  risk_level,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateRemediationStepStatus :one
UPDATE remediation_steps
SET status = $1
WHERE id = $2
RETURNING *;

-- name: CompleteRemediationStep :one
UPDATE remediation_steps
SET 
  status = 'completed',
  result = $1,
  executed_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: FailRemediationStep :one
UPDATE remediation_steps
SET 
  status = 'failed',
  result = $1,
  executed_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: DeleteRemediationSteps :exec
DELETE FROM remediation_steps
WHERE remediation_plan_id = $1;

-- Vulnerability Analysis Queries

-- name: GetVulnerabilityAnalysis :one
SELECT
  va.*
FROM vulnerability_analyses va
WHERE va.vulnerability_id = $1;

-- name: GetVulnerabilityAnalysisByTask :one
SELECT
  va.*
FROM vulnerability_analyses va
WHERE va.guardian_task_id = $1;

-- name: CreateVulnerabilityAnalysis :one
INSERT INTO vulnerability_analyses (
  vulnerability_id,
  vulnerability_type,
  severity,
  description,
  affected_resources,
  recommendations,
  estimated_impact,
  requires_downtime,
  guardian_task_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: DeleteVulnerabilityAnalysis :exec
DELETE FROM vulnerability_analyses
WHERE id = $1;

-- Guardian Backups Queries

-- name: GetGuardianBackups :many
SELECT
  gb.*
FROM guardian_backups gb
ORDER BY gb.created_at DESC;

-- name: GetGuardianBackupsByFilePath :many
SELECT
  gb.*
FROM guardian_backups gb
WHERE gb.file_path = $1
ORDER BY gb.created_at DESC;

-- name: GetGuardianBackupsByPlan :many
SELECT
  gb.*
FROM guardian_backups gb
WHERE gb.remediation_plan_id = $1
ORDER BY gb.created_at DESC;

-- name: GetGuardianBackup :one
SELECT
  gb.*
FROM guardian_backups gb
WHERE gb.id = $1;

-- name: CreateGuardianBackup :one
INSERT INTO guardian_backups (
  id,
  file_path,
  backup_path,
  size,
  remediation_plan_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: DeleteGuardianBackup :exec
DELETE FROM guardian_backups
WHERE id = $1;

-- name: DeleteOldGuardianBackups :exec
DELETE FROM guardian_backups
WHERE created_at < NOW() - INTERVAL '30 days';
