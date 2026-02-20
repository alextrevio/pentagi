import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Shield } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { z } from 'zod';

import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { VulnerabilityType } from '@/graphql/types';

const formSchema = z.object({
    title: z.string().min(3, 'Title must be at least 3 characters'),
    targetUrl: z.string().url('Must be a valid URL'),
    vulnerabilityType: z.nativeEnum(VulnerabilityType),
});

type FormValues = z.infer<typeof formSchema>;

const vulnerabilityTypeOptions = [
    {
        value: VulnerabilityType.MissingSecurityHeaders,
        label: 'Missing Security Headers',
        description: 'HTTP security headers are not configured',
    },
    {
        value: VulnerabilityType.WeakSslConfiguration,
        label: 'Weak SSL Configuration',
        description: 'SSL/TLS configuration is weak or outdated',
    },
    {
        value: VulnerabilityType.OutdatedSoftware,
        label: 'Outdated Software',
        description: 'Server software contains known vulnerabilities',
    },
    {
        value: VulnerabilityType.OpenPorts,
        label: 'Open Ports',
        description: 'Unnecessary ports are exposed',
    },
    {
        value: VulnerabilityType.Misconfiguration,
        label: 'Misconfiguration',
        description: 'Server or application is misconfigured',
    },
];

const NewGuardianTask = () => {
    const navigate = useNavigate();
    const [isSubmitting, setIsSubmitting] = useState(false);

    const form = useForm<FormValues>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            title: '',
            targetUrl: '',
            vulnerabilityType: VulnerabilityType.MissingSecurityHeaders,
        },
    });

    const onSubmit = async (values: FormValues) => {
        setIsSubmitting(true);
        
        try {
            // TODO: Implementar creación real de tarea
            console.log('Creating Guardian task:', values);
            
            // Simular delay
            await new Promise((resolve) => setTimeout(resolve, 1000));
            
            // Navegar a la lista de tareas
            navigate('/guardian');
        } catch (error) {
            console.error('Error creating task:', error);
        } finally {
            setIsSubmitting(false);
        }
    };

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
                            <BreadcrumbPage>New Task</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </header>

            <div className="flex flex-1 flex-col gap-4 p-4">
                <div className="flex items-center gap-2">
                    <Shield className="h-6 w-6" />
                    <h2 className="text-2xl font-bold tracking-tight">Create Guardian Task</h2>
                </div>

                <Card className="max-w-2xl">
                    <CardHeader>
                        <CardTitle>Task Configuration</CardTitle>
                        <CardDescription>
                            Configure a new auto-remediation task to fix security vulnerabilities
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                                <FormField
                                    control={form.control}
                                    name="title"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Task Title</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="e.g., Fix Missing Security Headers"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormDescription>
                                                A descriptive title for this remediation task
                                            </FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="targetUrl"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Target URL</FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="url"
                                                    placeholder="https://example.com"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormDescription>
                                                The URL of the web application to remediate
                                            </FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="vulnerabilityType"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Vulnerability Type</FormLabel>
                                            <Select
                                                onValueChange={field.onChange}
                                                defaultValue={field.value}
                                            >
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select vulnerability type" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    {vulnerabilityTypeOptions.map((option) => (
                                                        <SelectItem key={option.value} value={option.value}>
                                                            <div className="flex flex-col">
                                                                <span className="font-medium">{option.label}</span>
                                                                <span className="text-xs text-muted-foreground">
                                                                    {option.description}
                                                                </span>
                                                            </div>
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            <FormDescription>
                                                The type of vulnerability to remediate
                                            </FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <div className="flex gap-2">
                                    <Button
                                        type="button"
                                        variant="outline"
                                        onClick={() => navigate('/guardian')}
                                        disabled={isSubmitting}
                                    >
                                        Cancel
                                    </Button>
                                    <Button type="submit" disabled={isSubmitting}>
                                        {isSubmitting ? (
                                            <>
                                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                                Creating...
                                            </>
                                        ) : (
                                            'Create Task'
                                        )}
                                    </Button>
                                </div>
                            </form>
                        </Form>
                    </CardContent>
                </Card>
            </div>
        </>
    );
};

export default NewGuardianTask;
