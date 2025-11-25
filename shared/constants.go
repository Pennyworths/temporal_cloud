package shared

import "time"

// Workflow constants
const (
	// WorkflowName is the name of the hello workflow
	WorkflowName = "HelloWorkflow"

	// ScheduleWorkflowName is the name of the schedule helper workflow
	ScheduleWorkflowName = "ScheduleWorkflow"

	// DelayWorkflowName is the name of the delay workflow
	DelayWorkflowName = "DelayWorkflow"

	// WorkflowIDPrefix is the prefix for generated workflow IDs
	WorkflowIDPrefix = "hello-workflow-"

	// DefaultWorkflowName is the default name parameter for workflows
	DefaultWorkflowName = "Temporal User"
)

// WorkflowInput defines the payload accepted by HelloWorkflow
type WorkflowInput struct {
	Name      string // name parameter passed to the workflow
	AutoStart bool   // auto-start without waiting for signal (used by schedules)
}

// Signal constants
const (
	// SignalUpdateName is the signal name for updating the workflow name
	SignalUpdateName = "update-name"
)

// Timeout constants
const (
	// DefaultWorkflowTimeout is the default timeout for workflow execution
	DefaultWorkflowTimeout = 24 * time.Hour

	// DefaultWorkflowRunTimeout is the default timeout for a single workflow run
	DefaultWorkflowRunTimeout = 24 * time.Hour

	// DefaultContextTimeout is the default timeout for client operations
	DefaultContextTimeout = 30 * time.Second
)
