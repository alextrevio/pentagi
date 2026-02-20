import {
    AlertTriangle,
    CheckCircle,
    ChevronRight,
    Clock,
    FileText,
    Loader2,
    Shield,
    ShieldAlert,
    XCircle,
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { RemediationStatus, RiskLevel } from '@/graphql/types';

interface RemediationStep {
    id: number;
    toolName: string;
    description: string;
    arguments: Record<string, string>;
    order: number;
    riskLevel: RiskLevel;
    status: RemediationStatus;
    result?: string;
}

interface RemediationPlan {
    id: string;
    vulnerabilityId: string;
    description: string;
    steps: RemediationStep[];
    riskLevel: RiskLevel;
    estimatedTime: number;
    requiresBackup: boolean;
    requiresDowntime: boolean;
    rollbackPlan: string;
    status: RemediationStatus;
    createdAt: string;
    approvedBy?: string;
    approvedAt?: string;
}

// Mock data
const mockPlan: RemediationPlan = {
    id: '1',
    vulnerabilityId: 'vuln-001',
    description: 'Add missing security headers to Nginx configuration',
    riskLevel: RiskLevel.Medium,
    estimatedTime: 5,
    requiresBackup: true,
    requiresDowntime: false,
    rollbackPlan: '1. Stop execution\n2. Restore from backup\n3. Restart services\n4. Verify system state',
    status: RemediationStatus.PendingApproval,
    createdAt: new Date().toISOString(),
    steps: [
        {
            id: 1,
            toolName: 'http_header_scanner',
            description: 'Scan URL to identify missing headers',
            arguments: { url: 'https://example.com' },
            order: 1,
            riskLevel: RiskLevel.Low,
            status: RemediationStatus.PendingApproval,
        },
        {
            id: 2,
            toolName: 'file_backup_manager',
            description: 'Create backup of current configuration',
            arguments: { file_path: '/etc/nginx/nginx.conf', action: 'create' },
            order: 2,
            riskLevel: RiskLevel.Low,
            status: RemediationStatus.PendingApproval,
        },
        {
            id: 3,
            toolName: 'nginx_config_editor',
            description: 'Add security headers to Nginx configuration',
            arguments: { config_path: '/etc/nginx/nginx.conf', backup: 'true' },
            order: 3,
            riskLevel: RiskLevel.Medium,
            status: RemediationStatus.PendingApproval,
        },
        {
            id: 4,
            toolName: 'docker_service_manager',
            description: 'Reload Nginx service',
            arguments: { container_id: 'nginx-container', action: 'reload', service: 'nginx' },
            order: 4,
            riskLevel: RiskLevel.Low,
            status: RemediationStatus.PendingApproval,
        },
        {
            id: 5,
            toolName: 'http_header_scanner',
            description: 'Verify headers were added correctly',
            arguments: { url: 'https://example.com' },
            order: 5,
            riskLevel: RiskLevel.Low,
            status: RemediationStatus.PendingApproval,
        },
    ],
};

const riskLevelConfig: Record<
    RiskLevel,
    { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary'; icon: React.ReactNode }
> = {
    [RiskLevel.Low]: {
        label: 'Low',
        variant: 'secondary',
        icon: <Shield className="h-4 w-4" />,
    },
    [RiskLevel.Medium]: {
        label: 'Medium',
        variant: 'outline',
        icon: <ShieldAlert className="h-4 w-4" />,
    },
    [RiskLevel.High]: {
        label: 'High',
        variant: 'default',
        icon: <ShieldAlert className="h-4 w-4" />,
    },
    [RiskLevel.Critical]: {
        label: 'Critical',
        variant: 'destructive',
        icon: <AlertTriangle className="h-4 w-4" />,
    },
};

const RemediationPlanPage = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [plan] = useState<RemediationPlan>(mockPlan);
    const [isApproving, setIsApproving] = useState(false);
    const [isRejecting, setIsRejecting] = useState(false);

    const handleApprove = async () => {
        setIsApproving(true);
        // TODO: Implementar aprobación real
        setTimeout(() => {
            setIsApproving(false);
            console.log('Plan approved');
        }, 1000);
    };

    const handleReject = async () => {
        setIsRejecting(true);
        // TODO: Implementar rechazo real
        setTimeout(() => {
            setIsRejecting(false);
            console.log('Plan rejected');
        }, 1000);
    };

    const riskConfig = riskLevelConfig[plan.riskLevel];

    return (
        <>
            <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator orientation="vertical" className="mr-2 h-4" />
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <BreadcrumbLink href="/guardian">Guardian Tasks</BreadcrumbLink>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            <BreadcrumbPage>Remediation Plan</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </header>

            <div className="flex flex-1 flex-col gap-4 p-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h2 className="text-2xl font-bold tracking-tight">{plan.description}</h2>
                        <p className="text-sm text-muted-foreground">Vulnerability ID: {plan.vulnerabilityId}</p>
                    </div>
                    {plan.status === RemediationStatus.PendingApproval && (
                        <div className="flex gap-2">
                            <Button
                                variant="outline"
                                onClick={handleReject}
                                disabled={isRejecting || isApproving}
                            >
                                {isRejecting ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Rejecting...
                                    </>
                                ) : (
                                    <>
                                        <XCircle className="mr-2 h-4 w-4" />
                                        Reject
                                    </>
                                )}
                            </Button>
                            <Button onClick={handleApprove} disabled={isApproving || isRejecting}>
                                {isApproving ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Approving...
                                    </>
                                ) : (
                                    <>
                                        <CheckCircle className="mr-2 h-4 w-4" />
                                        Approve & Execute
                                    </>
                                )}
                            </Button>
                        </div>
                    )}
                </div>

                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Risk Level</CardTitle>
                            {riskConfig.icon}
                        </CardHeader>
                        <CardContent>
                            <Badge variant={riskConfig.variant} className="flex items-center gap-1 w-fit">
                                {riskConfig.label}
                            </Badge>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Estimated Time</CardTitle>
                            <Clock className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{plan.estimatedTime} min</div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Requires Backup</CardTitle>
                            <FileText className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{plan.requiresBackup ? 'Yes' : 'No'}</div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Downtime Required</CardTitle>
                            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{plan.requiresDowntime ? 'Yes' : 'No'}</div>
                        </CardContent>
                    </Card>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle>Remediation Steps</CardTitle>
                        <CardDescription>
                            The following steps will be executed in order to remediate the vulnerability
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {plan.steps.map((step, index) => {
                                const stepRiskConfig = riskLevelConfig[step.riskLevel];
                                return (
                                    <div key={step.id} className="flex items-start gap-4">
                                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border bg-background">
                                            {index + 1}
                                        </div>
                                        <div className="flex-1 space-y-2">
                                            <div className="flex items-center justify-between">
                                                <div className="flex items-center gap-2">
                                                    <h4 className="font-semibold">{step.toolName}</h4>
                                                    <Badge variant={stepRiskConfig.variant} className="flex items-center gap-1">
                                                        {stepRiskConfig.icon}
                                                        {stepRiskConfig.label}
                                                    </Badge>
                                                </div>
                                            </div>
                                            <p className="text-sm text-muted-foreground">{step.description}</p>
                                            <div className="rounded-md bg-muted p-3">
                                                <p className="text-xs font-mono">
                                                    {JSON.stringify(step.arguments, null, 2)}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Rollback Plan</CardTitle>
                        <CardDescription>
                            In case of failure, the following steps will be executed to restore the system
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <pre className="whitespace-pre-wrap text-sm">{plan.rollbackPlan}</pre>
                    </CardContent>
                </Card>
            </div>
        </>
    );
};

export default RemediationPlanPage;
