-- +goose Up
-- +goose StatementBegin

-- Enum types for Guardian
CREATE TYPE VULNERABILITY_TYPE AS ENUM (
    'missing_security_headers',
    'weak_ssl_configuration',
    'outdated_software',
    'sql_injection',
    'xss',
    'csrf',
    'open_ports',
    'weak_authentication',
    'misconfiguration'
);

CREATE TYPE REMEDIATION_STATUS AS ENUM (
    'pending_approval',
    'approved',
    'rejected',
    'in_progress',
    'completed',
    'failed',
    'rolled_back'
);

CREATE TYPE RISK_LEVEL AS ENUM (
    'low',
    'medium',
    'high',
    'critical'
);

-- Guardian tasks table
CREATE TABLE guardian_tasks (
    id              BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title           TEXT            NOT NULL,
    target_url      TEXT            NOT NULL,
    vulnerability_type VULNERABILITY_TYPE NOT NULL,
    status          REMEDIATION_STATUS NOT NULL DEFAULT 'pending_approval',
    flow_id         BIGINT          NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
    task_id         BIGINT          REFERENCES tasks(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT guardian_tasks_title_check CHECK (LENGTH(title) >= 3)
);

CREATE INDEX guardian_tasks_flow_id_idx ON guardian_tasks(flow_id);
CREATE INDEX guardian_tasks_task_id_idx ON guardian_tasks(task_id);
CREATE INDEX guardian_tasks_status_idx ON guardian_tasks(status);
CREATE INDEX guardian_tasks_created_at_idx ON guardian_tasks(created_at DESC);

-- Remediation plans table
CREATE TABLE remediation_plans (
    id                  BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    vulnerability_id    TEXT            NOT NULL,
    vulnerability_type  VULNERABILITY_TYPE NOT NULL,
    description         TEXT            NOT NULL,
    risk_level          RISK_LEVEL      NOT NULL,
    estimated_time      INTEGER         NOT NULL DEFAULT 0,
    requires_backup     BOOLEAN         NOT NULL DEFAULT true,
    requires_downtime   BOOLEAN         NOT NULL DEFAULT false,
    rollback_plan       TEXT            NOT NULL,
    status              REMEDIATION_STATUS NOT NULL DEFAULT 'pending_approval',
    flow_id             BIGINT          NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
    task_id             BIGINT          REFERENCES tasks(id) ON DELETE CASCADE,
    guardian_task_id    BIGINT          REFERENCES guardian_tasks(id) ON DELETE CASCADE,
    approved_by         TEXT            NULL,
    approved_at         TIMESTAMPTZ     NULL,
    executed_at         TIMESTAMPTZ     NULL,
    completed_at        TIMESTAMPTZ     NULL,
    created_at          TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT remediation_plans_estimated_time_check CHECK (estimated_time >= 0)
);

CREATE INDEX remediation_plans_flow_id_idx ON remediation_plans(flow_id);
CREATE INDEX remediation_plans_task_id_idx ON remediation_plans(task_id);
CREATE INDEX remediation_plans_guardian_task_id_idx ON remediation_plans(guardian_task_id);
CREATE INDEX remediation_plans_status_idx ON remediation_plans(status);
CREATE INDEX remediation_plans_created_at_idx ON remediation_plans(created_at DESC);

-- Remediation steps table
CREATE TABLE remediation_steps (
    id                  BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    remediation_plan_id BIGINT          NOT NULL REFERENCES remediation_plans(id) ON DELETE CASCADE,
    tool_name           TEXT            NOT NULL,
    description         TEXT            NOT NULL,
    arguments           JSONB           NOT NULL DEFAULT '{}',
    step_order          INTEGER         NOT NULL,
    risk_level          RISK_LEVEL      NOT NULL,
    status              REMEDIATION_STATUS NOT NULL DEFAULT 'pending_approval',
    result              TEXT            NULL,
    executed_at         TIMESTAMPTZ     NULL,
    created_at          TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT remediation_steps_order_check CHECK (step_order > 0),
    CONSTRAINT remediation_steps_unique_order UNIQUE (remediation_plan_id, step_order)
);

CREATE INDEX remediation_steps_plan_id_idx ON remediation_steps(remediation_plan_id);
CREATE INDEX remediation_steps_order_idx ON remediation_steps(remediation_plan_id, step_order);

-- Vulnerability analysis table
CREATE TABLE vulnerability_analyses (
    id                  BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    vulnerability_id    TEXT            NOT NULL UNIQUE,
    vulnerability_type  VULNERABILITY_TYPE NOT NULL,
    severity            RISK_LEVEL      NOT NULL,
    description         TEXT            NOT NULL,
    affected_resources  TEXT[]          NOT NULL DEFAULT '{}',
    recommendations     TEXT[]          NOT NULL DEFAULT '{}',
    estimated_impact    TEXT            NOT NULL,
    requires_downtime   BOOLEAN         NOT NULL DEFAULT false,
    guardian_task_id    BIGINT          REFERENCES guardian_tasks(id) ON DELETE CASCADE,
    created_at          TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX vulnerability_analyses_task_id_idx ON vulnerability_analyses(guardian_task_id);
CREATE INDEX vulnerability_analyses_type_idx ON vulnerability_analyses(vulnerability_type);

-- Backups table
CREATE TABLE guardian_backups (
    id              TEXT            PRIMARY KEY,
    file_path       TEXT            NOT NULL,
    backup_path     TEXT            NOT NULL,
    size            BIGINT          NOT NULL,
    remediation_plan_id BIGINT      REFERENCES remediation_plans(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT guardian_backups_size_check CHECK (size >= 0)
);

CREATE INDEX guardian_backups_plan_id_idx ON guardian_backups(remediation_plan_id);
CREATE INDEX guardian_backups_file_path_idx ON guardian_backups(file_path);
CREATE INDEX guardian_backups_created_at_idx ON guardian_backups(created_at DESC);

-- Add Guardian privileges
INSERT INTO privileges (role_id, name) VALUES
    (1, 'guardian.admin'),
    (1, 'guardian.create'),
    (1, 'guardian.view'),
    (1, 'guardian.approve'),
    (1, 'guardian.execute'),
    (1, 'guardian.rollback'),
    (2, 'guardian.view')
    ON CONFLICT DO NOTHING;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_guardian_tasks_updated_at BEFORE UPDATE ON guardian_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_remediation_plans_updated_at BEFORE UPDATE ON remediation_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_remediation_plans_updated_at ON remediation_plans;
DROP TRIGGER IF EXISTS update_guardian_tasks_updated_at ON guardian_tasks;
DROP FUNCTION IF EXISTS update_updated_at_column();

DELETE FROM privileges WHERE name LIKE 'guardian.%';

DROP TABLE IF EXISTS guardian_backups;
DROP TABLE IF EXISTS vulnerability_analyses;
DROP TABLE IF EXISTS remediation_steps;
DROP TABLE IF EXISTS remediation_plans;
DROP TABLE IF EXISTS guardian_tasks;

DROP TYPE IF EXISTS RISK_LEVEL;
DROP TYPE IF EXISTS REMEDIATION_STATUS;
DROP TYPE IF EXISTS VULNERABILITY_TYPE;

-- +goose StatementEnd
