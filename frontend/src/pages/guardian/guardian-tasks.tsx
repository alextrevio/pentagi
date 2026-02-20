import {
    CheckCircle,
    Clock,
    Eye,
    Loader2,
    MoreHorizontal,
    Plus,
    Shield,
    ShieldAlert,
    Trash,
    XCircle,
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { RemediationStatus, RiskLevel, VulnerabilityType } from '@/graphql/types';

// Mock data - será reemplazado por datos reales de GraphQL
interface GuardianTask {
    id: string;
    title: string;
    targetUrl: string;
    vulnerabilityType: VulnerabilityType;
    status: RemediationStatus;
    riskLevel: RiskLevel;
    createdAt: string;
    updatedAt: string;
}

const mockTasks: GuardianTask[] = [
    {
        id: '1',
        title: 'Fix Missing Security Headers',
        targetUrl: 'https://example.com',
        vulnerabilityType: VulnerabilityType.MissingSecurityHeaders,
        status: RemediationStatus.PendingApproval,
        riskLevel: RiskLevel.Medium,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
    },
];

const statusConfig: Record<
    RemediationStatus,
    { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary' }
> = {
    [RemediationStatus.PendingApproval]: {
        label: 'Pending Approval',
        variant: 'outline',
    },
    [RemediationStatus.Approved]: {
        label: 'Approved',
        variant: 'default',
    },
    [RemediationStatus.Rejected]: {
        label: 'Rejected',
        variant: 'destructive',
    },
    [RemediationStatus.InProgress]: {
        label: 'In Progress',
        variant: 'default',
    },
    [RemediationStatus.Completed]: {
        label: 'Completed',
        variant: 'secondary',
    },
    [RemediationStatus.Failed]: {
        label: 'Failed',
        variant: 'destructive',
    },
    [RemediationStatus.RolledBack]: {
        label: 'Rolled Back',
        variant: 'outline',
    },
};

const riskLevelConfig: Record<
    RiskLevel,
    { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary'; icon: React.ReactNode }
> = {
    [RiskLevel.Low]: {
        label: 'Low',
        variant: 'secondary',
        icon: <Shield className="h-3 w-3" />,
    },
    [RiskLevel.Medium]: {
        label: 'Medium',
        variant: 'outline',
        icon: <ShieldAlert className="h-3 w-3" />,
    },
    [RiskLevel.High]: {
        label: 'High',
        variant: 'default',
        icon: <ShieldAlert className="h-3 w-3" />,
    },
    [RiskLevel.Critical]: {
        label: 'Critical',
        variant: 'destructive',
        icon: <ShieldAlert className="h-3 w-3" />,
    },
};

const GuardianTasks = () => {
    const navigate = useNavigate();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingTask, setDeletingTask] = useState<GuardianTask | null>(null);
    const [isLoading] = useState(false);
    const [tasks] = useState<GuardianTask[]>(mockTasks);

    const handleTaskOpen = (taskId: string) => {
        navigate(`/guardian/${taskId}`);
    };

    const handleTaskDeleteDialogOpen = (task: GuardianTask) => {
        setDeletingTask(task);
        setIsDeleteDialogOpen(true);
    };

    const handleTaskDelete = async () => {
        if (!deletingTask) {
            return;
        }

        // TODO: Implementar eliminación real
        console.log('Deleting task:', deletingTask.id);
        setIsDeleteDialogOpen(false);
        setDeletingTask(null);
    };

    const handleNewTask = () => {
        navigate('/guardian/new');
    };

    const columns = [
        {
            key: 'title',
            label: 'Title',
            render: (task: GuardianTask) => (
                <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">{task.title}</span>
                </div>
            ),
        },
        {
            key: 'targetUrl',
            label: 'Target URL',
            render: (task: GuardianTask) => (
                <span className="text-sm text-muted-foreground">{task.targetUrl}</span>
            ),
        },
        {
            key: 'vulnerabilityType',
            label: 'Vulnerability Type',
            render: (task: GuardianTask) => (
                <Badge variant="outline">{task.vulnerabilityType.replace(/_/g, ' ')}</Badge>
            ),
        },
        {
            key: 'riskLevel',
            label: 'Risk Level',
            render: (task: GuardianTask) => {
                const config = riskLevelConfig[task.riskLevel];
                return (
                    <Badge variant={config.variant} className="flex items-center gap-1 w-fit">
                        {config.icon}
                        {config.label}
                    </Badge>
                );
            },
        },
        {
            key: 'status',
            label: 'Status',
            render: (task: GuardianTask) => {
                const config = statusConfig[task.status];
                return <Badge variant={config.variant}>{config.label}</Badge>;
            },
        },
        {
            key: 'createdAt',
            label: 'Created',
            render: (task: GuardianTask) => (
                <span className="text-sm text-muted-foreground">
                    {new Date(task.createdAt).toLocaleDateString()}
                </span>
            ),
        },
        {
            key: 'actions',
            label: '',
            render: (task: GuardianTask) => (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button size="icon" variant="ghost">
                            <MoreHorizontal className="h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => handleTaskOpen(task.id)}>
                            <Eye className="mr-2 h-4 w-4" />
                            View Details
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                            onClick={() => handleTaskDeleteDialogOpen(task)}
                            className="text-destructive"
                        >
                            <Trash className="mr-2 h-4 w-4" />
                            Delete
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            ),
        },
    ];

    const getStatusCounts = () => {
        const counts = {
            total: tasks.length,
            pending: tasks.filter((t) => t.status === RemediationStatus.PendingApproval).length,
            inProgress: tasks.filter((t) => t.status === RemediationStatus.InProgress).length,
            completed: tasks.filter((t) => t.status === RemediationStatus.Completed).length,
        };
        return counts;
    };

    const statusCounts = getStatusCounts();

    return (
        <>
            <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator orientation="vertical" className="mr-2 h-4" />
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <BreadcrumbPage>Guardian Tasks</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </header>

            <div className="flex flex-1 flex-col gap-4 p-4">
                <div className="grid gap-4 md:grid-cols-4">
                    <StatusCard
                        title="Total Tasks"
                        value={statusCounts.total}
                        icon={<Shield className="h-4 w-4 text-muted-foreground" />}
                    />
                    <StatusCard
                        title="Pending Approval"
                        value={statusCounts.pending}
                        icon={<Clock className="h-4 w-4 text-muted-foreground" />}
                    />
                    <StatusCard
                        title="In Progress"
                        value={statusCounts.inProgress}
                        icon={<Loader2 className="h-4 w-4 text-muted-foreground animate-spin" />}
                    />
                    <StatusCard
                        title="Completed"
                        value={statusCounts.completed}
                        icon={<CheckCircle className="h-4 w-4 text-muted-foreground" />}
                    />
                </div>

                <div className="flex items-center justify-between">
                    <h2 className="text-2xl font-bold tracking-tight">Guardian Tasks</h2>
                    <Button onClick={handleNewTask}>
                        <Plus className="mr-2 h-4 w-4" />
                        New Task
                    </Button>
                </div>

                <DataTable
                    columns={columns}
                    data={tasks}
                    isLoading={isLoading}
                    onRowClick={(task) => handleTaskOpen(task.id)}
                />
            </div>

            <ConfirmationDialog
                open={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                onConfirm={handleTaskDelete}
                title="Delete Guardian Task"
                description={`Are you sure you want to delete "${deletingTask?.title}"? This action cannot be undone.`}
                confirmText="Delete"
                cancelText="Cancel"
            />
        </>
    );
};

export default GuardianTasks;
