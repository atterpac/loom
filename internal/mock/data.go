package mock

import (
	"time"
)

// Namespace represents a Temporal namespace.
type Namespace struct {
	Name            string
	State           string
	RetentionPeriod string
}

// Workflow represents a workflow execution.
type Workflow struct {
	ID        string
	RunID     string
	Type      string
	Status    string
	Namespace string
	TaskQueue string
	StartTime time.Time
	EndTime   *time.Time
	ParentID  *string
}

// HistoryEvent represents a workflow history event.
type HistoryEvent struct {
	ID      int64
	Type    string
	Time    time.Time
	Details string
}

// TaskQueue represents a task queue.
type TaskQueue struct {
	Name        string
	Type        string
	PollerCount int
	Backlog     int
}

// Poller represents a task queue poller.
type Poller struct {
	Identity      string
	LastAccessTime time.Time
	TaskQueueType string
}

func ptr[T any](v T) *T {
	return &v
}

// Namespaces returns mock namespace data.
func Namespaces() []Namespace {
	return []Namespace{
		{Name: "default", State: "Active", RetentionPeriod: "7 days"},
		{Name: "production", State: "Active", RetentionPeriod: "30 days"},
		{Name: "staging", State: "Active", RetentionPeriod: "3 days"},
		{Name: "development", State: "Active", RetentionPeriod: "1 day"},
		{Name: "archived", State: "Deprecated", RetentionPeriod: "90 days"},
	}
}

// Workflows returns mock workflow data for a namespace.
func Workflows(namespace string) []Workflow {
	now := time.Now()
	workflows := []Workflow{
		{
			ID:        "order-processing-abc123",
			RunID:     "run-001-xyz",
			Type:      "OrderWorkflow",
			Status:    "Running",
			Namespace: namespace,
			TaskQueue: "order-tasks",
			StartTime: now.Add(-5 * time.Minute),
		},
		{
			ID:        "payment-xyz789",
			RunID:     "run-002-abc",
			Type:      "PaymentWorkflow",
			Status:    "Completed",
			Namespace: namespace,
			TaskQueue: "payment-tasks",
			StartTime: now.Add(-1 * time.Hour),
			EndTime:   ptr(now.Add(-55 * time.Minute)),
		},
		{
			ID:        "shipment-def456",
			RunID:     "run-003-def",
			Type:      "ShipmentWorkflow",
			Status:    "Failed",
			Namespace: namespace,
			TaskQueue: "shipment-tasks",
			StartTime: now.Add(-30 * time.Minute),
			EndTime:   ptr(now.Add(-25 * time.Minute)),
		},
		{
			ID:        "inventory-check-111",
			RunID:     "run-004-ghi",
			Type:      "InventoryWorkflow",
			Status:    "Running",
			Namespace: namespace,
			TaskQueue: "inventory-tasks",
			StartTime: now.Add(-10 * time.Minute),
		},
		{
			ID:        "user-signup-222",
			RunID:     "run-005-jkl",
			Type:      "UserOnboardingWorkflow",
			Status:    "Completed",
			Namespace: namespace,
			TaskQueue: "user-tasks",
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   ptr(now.Add(-1*time.Hour - 45*time.Minute)),
		},
		{
			ID:        "refund-process-333",
			RunID:     "run-006-mno",
			Type:      "RefundWorkflow",
			Status:    "Canceled",
			Namespace: namespace,
			TaskQueue: "payment-tasks",
			StartTime: now.Add(-45 * time.Minute),
			EndTime:   ptr(now.Add(-40 * time.Minute)),
		},
		{
			ID:        "email-campaign-444",
			RunID:     "run-007-pqr",
			Type:      "EmailCampaignWorkflow",
			Status:    "Running",
			Namespace: namespace,
			TaskQueue: "email-tasks",
			StartTime: now.Add(-15 * time.Minute),
		},
		{
			ID:        "data-sync-555",
			RunID:     "run-008-stu",
			Type:      "DataSyncWorkflow",
			Status:    "Terminated",
			Namespace: namespace,
			TaskQueue: "sync-tasks",
			StartTime: now.Add(-3 * time.Hour),
			EndTime:   ptr(now.Add(-2 * time.Hour)),
		},
		{
			ID:        "report-gen-666",
			RunID:     "run-009-vwx",
			Type:      "ReportWorkflow",
			Status:    "Completed",
			Namespace: namespace,
			TaskQueue: "report-tasks",
			StartTime: now.Add(-4 * time.Hour),
			EndTime:   ptr(now.Add(-3*time.Hour - 30*time.Minute)),
		},
		{
			ID:        "cleanup-job-777",
			RunID:     "run-010-yz0",
			Type:      "CleanupWorkflow",
			Status:    "Running",
			Namespace: namespace,
			TaskQueue: "maintenance-tasks",
			StartTime: now.Add(-2 * time.Minute),
		},
		{
			ID:        "notification-888",
			RunID:     "run-011-123",
			Type:      "NotificationWorkflow",
			Status:    "Completed",
			Namespace: namespace,
			TaskQueue: "notification-tasks",
			StartTime: now.Add(-20 * time.Minute),
			EndTime:   ptr(now.Add(-19 * time.Minute)),
		},
		{
			ID:        "batch-import-999",
			RunID:     "run-012-456",
			Type:      "BatchImportWorkflow",
			Status:    "Failed",
			Namespace: namespace,
			TaskQueue: "import-tasks",
			StartTime: now.Add(-1*time.Hour - 30*time.Minute),
			EndTime:   ptr(now.Add(-1 * time.Hour)),
		},
		{
			ID:        "child-workflow-aaa",
			RunID:     "run-013-789",
			Type:      "ChildWorkflow",
			Status:    "Completed",
			Namespace: namespace,
			TaskQueue: "order-tasks",
			StartTime: now.Add(-4 * time.Minute),
			EndTime:   ptr(now.Add(-3 * time.Minute)),
			ParentID:  ptr("order-processing-abc123"),
		},
	}
	return workflows
}

// Events returns mock event history for a workflow.
func Events(workflowID string) []HistoryEvent {
	now := time.Now()
	return []HistoryEvent{
		{ID: 1, Type: "WorkflowExecutionStarted", Time: now.Add(-5 * time.Minute), Details: "WorkflowType: OrderWorkflow, TaskQueue: order-tasks"},
		{ID: 2, Type: "WorkflowTaskScheduled", Time: now.Add(-5 * time.Minute), Details: "TaskQueue: order-tasks"},
		{ID: 3, Type: "WorkflowTaskStarted", Time: now.Add(-5 * time.Minute), Details: "Identity: worker-1@host"},
		{ID: 4, Type: "WorkflowTaskCompleted", Time: now.Add(-5 * time.Minute), Details: "ScheduledEventId: 2"},
		{ID: 5, Type: "ActivityTaskScheduled", Time: now.Add(-4*time.Minute - 50*time.Second), Details: "ActivityType: ValidateOrder, TaskQueue: order-tasks"},
		{ID: 6, Type: "ActivityTaskStarted", Time: now.Add(-4*time.Minute - 45*time.Second), Details: "Identity: worker-1@host, Attempt: 1"},
		{ID: 7, Type: "ActivityTaskCompleted", Time: now.Add(-4*time.Minute - 40*time.Second), Details: "ScheduledEventId: 5, Result: {valid: true}"},
		{ID: 8, Type: "WorkflowTaskScheduled", Time: now.Add(-4*time.Minute - 40*time.Second), Details: "TaskQueue: order-tasks"},
		{ID: 9, Type: "WorkflowTaskStarted", Time: now.Add(-4*time.Minute - 35*time.Second), Details: "Identity: worker-1@host"},
		{ID: 10, Type: "WorkflowTaskCompleted", Time: now.Add(-4*time.Minute - 35*time.Second), Details: "ScheduledEventId: 8"},
		{ID: 11, Type: "ActivityTaskScheduled", Time: now.Add(-4*time.Minute - 30*time.Second), Details: "ActivityType: ReserveInventory, TaskQueue: inventory-tasks"},
		{ID: 12, Type: "ActivityTaskStarted", Time: now.Add(-4*time.Minute - 25*time.Second), Details: "Identity: worker-2@host, Attempt: 1"},
		{ID: 13, Type: "ActivityTaskCompleted", Time: now.Add(-4*time.Minute - 10*time.Second), Details: "ScheduledEventId: 11, Result: {reserved: true}"},
		{ID: 14, Type: "WorkflowTaskScheduled", Time: now.Add(-4*time.Minute - 10*time.Second), Details: "TaskQueue: order-tasks"},
		{ID: 15, Type: "WorkflowTaskStarted", Time: now.Add(-4*time.Minute - 5*time.Second), Details: "Identity: worker-1@host"},
		{ID: 16, Type: "WorkflowTaskCompleted", Time: now.Add(-4*time.Minute - 5*time.Second), Details: "ScheduledEventId: 14"},
		{ID: 17, Type: "ActivityTaskScheduled", Time: now.Add(-4 * time.Minute), Details: "ActivityType: ProcessPayment, TaskQueue: payment-tasks"},
		{ID: 18, Type: "ActivityTaskStarted", Time: now.Add(-3*time.Minute - 55*time.Second), Details: "Identity: worker-3@host, Attempt: 1"},
		{ID: 19, Type: "ActivityTaskFailed", Time: now.Add(-3*time.Minute - 50*time.Second), Details: "ScheduledEventId: 17, Failure: PaymentDeclined"},
		{ID: 20, Type: "ActivityTaskScheduled", Time: now.Add(-3*time.Minute - 45*time.Second), Details: "ActivityType: ProcessPayment, TaskQueue: payment-tasks, Retry: 2"},
		{ID: 21, Type: "ActivityTaskStarted", Time: now.Add(-3*time.Minute - 40*time.Second), Details: "Identity: worker-3@host, Attempt: 2"},
		{ID: 22, Type: "ActivityTaskCompleted", Time: now.Add(-3*time.Minute - 20*time.Second), Details: "ScheduledEventId: 20, Result: {transactionId: txn-123}"},
		{ID: 23, Type: "WorkflowTaskScheduled", Time: now.Add(-3*time.Minute - 20*time.Second), Details: "TaskQueue: order-tasks"},
		{ID: 24, Type: "WorkflowTaskStarted", Time: now.Add(-3*time.Minute - 15*time.Second), Details: "Identity: worker-1@host"},
		{ID: 25, Type: "WorkflowTaskCompleted", Time: now.Add(-3*time.Minute - 15*time.Second), Details: "ScheduledEventId: 23"},
		{ID: 26, Type: "StartChildWorkflowExecutionInitiated", Time: now.Add(-3*time.Minute - 10*time.Second), Details: "WorkflowType: ShipmentWorkflow, WorkflowId: shipment-def456"},
		{ID: 27, Type: "ChildWorkflowExecutionStarted", Time: now.Add(-3*time.Minute - 5*time.Second), Details: "WorkflowId: shipment-def456, RunId: run-child-001"},
		{ID: 28, Type: "TimerStarted", Time: now.Add(-3 * time.Minute), Details: "TimerId: wait-for-shipment, Duration: 24h"},
		{ID: 29, Type: "SignalExternalWorkflowExecutionInitiated", Time: now.Add(-2*time.Minute - 30*time.Second), Details: "WorkflowId: notification-888, SignalName: order-update"},
		{ID: 30, Type: "ExternalWorkflowExecutionSignaled", Time: now.Add(-2*time.Minute - 25*time.Second), Details: "WorkflowId: notification-888"},
	}
}

// TaskQueues returns mock task queue data.
func TaskQueues() []TaskQueue {
	return []TaskQueue{
		{Name: "order-tasks", Type: "Workflow", PollerCount: 5, Backlog: 12},
		{Name: "order-tasks", Type: "Activity", PollerCount: 10, Backlog: 45},
		{Name: "payment-tasks", Type: "Workflow", PollerCount: 3, Backlog: 0},
		{Name: "payment-tasks", Type: "Activity", PollerCount: 8, Backlog: 3},
		{Name: "shipment-tasks", Type: "Workflow", PollerCount: 2, Backlog: 5},
		{Name: "shipment-tasks", Type: "Activity", PollerCount: 4, Backlog: 15},
		{Name: "notification-tasks", Type: "Workflow", PollerCount: 2, Backlog: 0},
		{Name: "notification-tasks", Type: "Activity", PollerCount: 6, Backlog: 100},
	}
}

// Pollers returns mock poller data for a task queue.
func Pollers(taskQueue string) []Poller {
	now := time.Now()
	return []Poller{
		{Identity: "worker-1@host-001", LastAccessTime: now.Add(-5 * time.Second), TaskQueueType: "Workflow"},
		{Identity: "worker-1@host-001", LastAccessTime: now.Add(-3 * time.Second), TaskQueueType: "Activity"},
		{Identity: "worker-2@host-002", LastAccessTime: now.Add(-10 * time.Second), TaskQueueType: "Workflow"},
		{Identity: "worker-2@host-002", LastAccessTime: now.Add(-2 * time.Second), TaskQueueType: "Activity"},
		{Identity: "worker-3@host-003", LastAccessTime: now.Add(-1 * time.Second), TaskQueueType: "Activity"},
	}
}

// FormatDuration formats a duration as a human-readable string.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return string(rune('0'+mins/10)) + string(rune('0'+mins%10)) + "m ago"
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return string(rune('0'+hours/10)) + string(rune('0'+hours%10)) + "h ago"
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1d ago"
	}
	return string(rune('0'+days/10)) + string(rune('0'+days%10)) + "d ago"
}

// FormatTime formats a time relative to now.
func FormatTime(t time.Time) string {
	return FormatDuration(time.Since(t))
}
