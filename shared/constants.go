package shared

import "time"

// Workflow constants
const (
	// WorkflowName is the name of the hello workflow
	WorkflowName = "HelloWorkflow"

	// WorkflowIDPrefix is the prefix for generated workflow IDs
	WorkflowIDPrefix = "hello-workflow-"

	// DefaultWorkflowName is the default name parameter for workflows
	DefaultWorkflowName = "Temporal User"
)

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
