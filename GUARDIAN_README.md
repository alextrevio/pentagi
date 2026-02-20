# PentAGI Guardian - Auto-Remediation Module

## Overview

**PentAGI Guardian** is an extension to PentAGI that adds autonomous security remediation capabilities. While PentAGI identifies vulnerabilities, Guardian can automatically fix them with human approval.

## Features

### Core Capabilities

- **Autonomous Remediation**: Automatically fixes security vulnerabilities after human approval
- **Risk Assessment**: Evaluates the risk level of each remediation action
- **Backup & Rollback**: Creates backups before any modification and can rollback on failure
- **Multi-Tool Integration**: Uses specialized tools for different remediation tasks
- **Real-time Monitoring**: Tracks remediation progress and status

### Security Features

- ✅ **Mandatory Human Approval**: All remediation plans require explicit approval
- ✅ **Automatic Backups**: Files are backed up before any modification
- ✅ **Syntax Validation**: Configuration changes are validated before applying
- ✅ **Automatic Rollback**: Failed operations are automatically reverted
- ✅ **Dry-Run Mode**: Test remediation plans without making actual changes

## Architecture

### Backend Components

```
backend/pkg/guardian/
├── guardian.go                    # Core Guardian module
├── agents/
│   └── security_agent.go         # Security-focused AI agent
└── tools/
    ├── http_header_scanner.go    # Scans for missing HTTP headers
    ├── nginx_config_editor.go    # Safely edits Nginx configuration
    ├── docker_service_manager.go # Manages Docker services
    └── file_backup_manager.go    # Handles file backups


### Frontend Components

```
frontend/src/pages/guardian/
├── guardian-tasks.tsx           # Main dashboard for Guardian tasks
├── remediation-plan.tsx         # Detailed view of remediation plans
└── new-guardian-task.tsx        # Create new remediation tasks
```

### GraphQL API

Guardian extends PentAGI's GraphQL schema with:

- **Types**: `RemediationPlan`, `GuardianTask`, `VulnerabilityAnalysis`, `BackupInfo`
- **Queries**: `guardianTasks`, `remediationPlans`, `analyzeVulnerability`
- **Mutations**: `createGuardianTask`, `approveRemediationPlan`, `executeRemediationPlan`
- **Subscriptions**: `guardianTaskUpdated`, `remediationPlanUpdated`

## Supported Vulnerabilities

### Phase 1 (PoC) - Implemented

- ✅ **Missing Security Headers**: Automatically adds HTTP security headers to Nginx

### Phase 2 (Planned)

- ⏳ **Weak SSL Configuration**: Updates SSL/TLS settings
- ⏳ **Outdated Software**: Updates packages and dependencies
- ⏳ **Open Ports**: Closes unnecessary exposed ports
- ⏳ **Misconfiguration**: Fixes common server misconfigurations

## Installation & Setup

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for development)
- Node.js 18+ (for frontend development)

### Quick Start

1. **Clone the repository** (if not already done):
   ```bash
   git clone https://github.com/vxcontrol/pentagi.git
   cd pentagi
   git checkout feature/guardian-auto-remediation
   ```

2. **Start the test environment**:
   ```bash
   docker-compose -f docker-compose-guardian-test.yml up -d
   ```

3. **Access the vulnerable test server**:
   ```
   http://localhost:8080
   ```

4. **Run Guardian to remediate**:
   ```bash
   # From PentAGI interface, create a new Guardian task
   # Target: http://vulnerable-nginx
   # Type: Missing Security Headers
   ```

## Usage Guide

### Creating a Remediation Task

1. Navigate to **Guardian Tasks** in the sidebar
2. Click **"New Task"**
3. Fill in the form:
   - **Title**: Descriptive name for the task
   - **Target URL**: URL of the vulnerable application
   - **Vulnerability Type**: Select from the dropdown
4. Click **"Create Task"**

### Approving a Remediation Plan

1. Guardian will analyze the vulnerability and create a remediation plan
2. Review the plan details:
   - **Risk Level**: Understand the potential impact
   - **Steps**: See exactly what will be executed
   - **Rollback Plan**: Know how to revert if needed
3. Click **"Approve & Execute"** to proceed
4. Monitor the execution in real-time

### Rolling Back a Remediation

If a remediation fails or causes issues:

1. Navigate to the remediation plan
2. Click **"Rollback"**
3. Select the backup to restore
4. Confirm the rollback

## Development

### Running Tests

```bash
# Backend tests
cd backend
go test -v ./pkg/guardian/...

# Frontend tests
cd frontend
npm test
```

### Building for Production

```bash
# Build backend
cd backend
go build -o guardian ./cmd/pentagi

# Build frontend
cd frontend
npm run build
```

## Configuration

### Environment Variables

```env
# Guardian Configuration
GUARDIAN_REQUIRE_APPROVAL=true
GUARDIAN_DRY_RUN=false
GUARDIAN_MAX_RETRIES=3
GUARDIAN_TIMEOUT=1800
GUARDIAN_BACKUP_DIR=/tmp/guardian_backups

# Auto-approval (use with caution)
GUARDIAN_ENABLE_AUTO_APPROVAL=false
```

### Config File

Create a `guardian.yml` file:

```yaml
guardian:
  require_approval: true
  dry_run: false
  max_retries: 3
  timeout: 1800s
  backup_dir: /tmp/guardian_backups
  
  # Risk levels that can be auto-approved (if enabled)
  auto_approve_risk_levels:
    - low
```

## Security Considerations

### Best Practices

1. **Always Review Plans**: Never auto-approve without reviewing the remediation plan
2. **Test in Staging**: Run Guardian in a staging environment first
3. **Backup Everything**: Ensure backups are working before running in production
4. **Monitor Closely**: Watch the execution in real-time
5. **Have Rollback Ready**: Know how to quickly rollback if needed

### Risk Levels

- **Low**: Safe operations with minimal impact (e.g., adding HTTP headers)
- **Medium**: Operations that may require service reload (e.g., config changes)
- **High**: Operations that require service restart (e.g., SSL updates)
- **Critical**: Operations that may cause downtime (e.g., major updates)

## Troubleshooting

### Common Issues

#### Guardian can't connect to target

```bash
# Check network connectivity
docker exec guardian-test-runner ping vulnerable-nginx

# Verify target is running
curl http://localhost:8080
```

#### Backup creation fails

```bash
# Check backup directory permissions
ls -la /tmp/guardian_backups

# Ensure sufficient disk space
df -h
```

#### Nginx validation fails

```bash
# Test Nginx configuration manually
docker exec guardian-test-nginx nginx -t

# View Nginx error logs
docker logs guardian-test-nginx
```

## Roadmap

### Phase 1: PoC (Current)
- ✅ Core Guardian module
- ✅ HTTP header scanner
- ✅ Nginx config editor
- ✅ Backup manager
- ✅ Frontend UI
- ✅ Test environment

### Phase 2: Expansion
- ⏳ SSL/TLS remediation
- ⏳ Package updates
- ⏳ Firewall configuration
- ⏳ Apache support
- ⏳ Kubernetes integration

### Phase 3: Intelligence
- ⏳ AI-powered decision making
- ⏳ Learning from past remediations
- ⏳ Predictive vulnerability detection
- ⏳ Automated testing after remediation

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Write tests
5. Commit with descriptive messages
6. Push and create a Pull Request

## License

This project follows the same license as PentAGI. See the main LICENSE file for details.

## Support

For issues, questions, or contributions:

- **GitHub Issues**: https://github.com/vxcontrol/pentagi/issues
- **Documentation**: https://pentagi.com/docs/guardian
- **Community**: Join the PentAGI Discord/Telegram

## Acknowledgments

Built on top of PentAGI by VXControl. Special thanks to the open-source security community.

---

**⚠️ Important**: Guardian is a powerful tool that can modify production systems. Always test in a safe environment first and never disable the approval requirement in production.
