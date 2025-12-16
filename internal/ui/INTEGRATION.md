# UI Integration Guide

This document describes how to integrate the loom UI components with real Temporal data.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        view/app.go                          │
│                    (Main Controller)                        │
├─────────────────────────────────────────────────────────────┤
│   view/namespace_list.go  │  view/workflow_list.go  │ ...   │
│        (Views)            │       (Views)           │       │
├─────────────────────────────────────────────────────────────┤
│                      ui/ (Primitives)                       │
│   App, Table, Menu, Crumbs, Pages, Actions, Styles          │
├─────────────────────────────────────────────────────────────┤
│                    Data Layer (YOUR CODE)                   │
│              Temporal SDK Client / Repository               │
└─────────────────────────────────────────────────────────────┘
```

## Component Interface

All views implement the `Component` interface defined in `pages.go`:

```go
type Component interface {
    tview.Primitive
    Name() string      // Unique identifier for the view
    Start()            // Called when view becomes active
    Stop()             // Called when view is deactivated
    Hints() []KeyHint  // Keybindings for menu display
}
```

## Integration Steps

### 1. Create a Data Provider Interface

Create an interface that abstracts data fetching. This allows swapping mock data for real Temporal SDK calls:

```go
// internal/temporal/provider.go
package temporal

import (
    "context"
    "time"
)

type Namespace struct {
    Name            string
    State           string
    RetentionPeriod time.Duration
}

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

type HistoryEvent struct {
    ID      int64
    Type    string
    Time    time.Time
    Details map[string]interface{}
}

type Provider interface {
    // Namespace operations
    ListNamespaces(ctx context.Context) ([]Namespace, error)

    // Workflow operations
    ListWorkflows(ctx context.Context, namespace string, opts ListOptions) ([]Workflow, error)
    GetWorkflow(ctx context.Context, namespace, workflowID, runID string) (*Workflow, error)
    GetWorkflowHistory(ctx context.Context, namespace, workflowID, runID string) ([]HistoryEvent, error)

    // Task queue operations
    DescribeTaskQueue(ctx context.Context, namespace, taskQueue string) (*TaskQueueInfo, error)
}

type ListOptions struct {
    PageSize  int
    PageToken []byte
    Query     string // For workflow filtering
}
```

### 2. Implement the Provider with Temporal SDK

```go
// internal/temporal/client.go
package temporal

import (
    "context"

    "go.temporal.io/sdk/client"
    "go.temporal.io/api/workflowservice/v1"
)

type Client struct {
    client client.Client
}

func NewClient(hostPort, namespace string) (*Client, error) {
    c, err := client.Dial(client.Options{
        HostPort:  hostPort,
        Namespace: namespace,
    })
    if err != nil {
        return nil, err
    }
    return &Client{client: c}, nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]Namespace, error) {
    resp, err := c.client.WorkflowService().ListNamespaces(ctx, &workflowservice.ListNamespacesRequest{})
    if err != nil {
        return nil, err
    }

    var namespaces []Namespace
    for _, ns := range resp.Namespaces {
        namespaces = append(namespaces, Namespace{
            Name:            ns.NamespaceInfo.Name,
            State:           ns.NamespaceInfo.State.String(),
            RetentionPeriod: ns.Config.WorkflowExecutionRetentionTtl.AsDuration(),
        })
    }
    return namespaces, nil
}

func (c *Client) ListWorkflows(ctx context.Context, namespace string, opts ListOptions) ([]Workflow, error) {
    resp, err := c.client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
        Namespace: namespace,
        PageSize:  int32(opts.PageSize),
        Query:     opts.Query,
    })
    if err != nil {
        return nil, err
    }

    var workflows []Workflow
    for _, exec := range resp.Executions {
        workflows = append(workflows, Workflow{
            ID:        exec.Execution.WorkflowId,
            RunID:     exec.Execution.RunId,
            Type:      exec.Type.Name,
            Status:    exec.Status.String(),
            StartTime: exec.StartTime.AsTime(),
            // ... map other fields
        })
    }
    return workflows, nil
}

// Implement remaining methods...
```

### 3. Inject Provider into Views

Modify the view constructors to accept a data provider:

```go
// internal/view/app.go
package view

import "github.com/atterpac/loom/internal/temporal"

type App struct {
    ui       *ui.App
    provider temporal.Provider
    // ...
}

func NewApp(provider temporal.Provider) *App {
    a := &App{
        ui:       ui.NewApp(),
        provider: provider,
    }
    a.setup()
    return a
}
```

```go
// internal/view/namespace_list.go
package view

type NamespaceList struct {
    *ui.Table
    app      *App
    provider temporal.Provider
}

func NewNamespaceList(app *App) *NamespaceList {
    nl := &NamespaceList{
        Table:    ui.NewTable(),
        app:      app,
        provider: app.provider,
    }
    nl.setup()
    return nl
}

func (nl *NamespaceList) setup() {
    nl.SetHeaders("NAME", "STATE", "RETENTION")
    // ... table setup

    // Load data asynchronously
    go nl.loadData()
}

func (nl *NamespaceList) loadData() {
    ctx := context.Background()
    namespaces, err := nl.provider.ListNamespaces(ctx)
    if err != nil {
        // Handle error - show in UI
        nl.app.UI().QueueUpdateDraw(func() {
            nl.showError(err)
        })
        return
    }

    // Update UI on main thread
    nl.app.UI().QueueUpdateDraw(func() {
        nl.populateTable(namespaces)
    })
}
```

### 4. Handle Async Data Loading

Use `tview.Application.QueueUpdateDraw()` for thread-safe UI updates:

```go
func (wl *WorkflowList) refresh() {
    // Show loading indicator
    wl.setLoading(true)

    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        workflows, err := wl.provider.ListWorkflows(ctx, wl.namespace, ListOptions{
            PageSize: 100,
        })

        wl.app.UI().QueueUpdateDraw(func() {
            wl.setLoading(false)
            if err != nil {
                wl.showError(err)
                return
            }
            wl.populateTable(workflows)
        })
    }()
}
```

### 5. Add Refresh and Polling

Implement periodic refresh for live data:

```go
type WorkflowList struct {
    // ...
    refreshTicker *time.Ticker
    stopRefresh   chan struct{}
}

func (wl *WorkflowList) Start() {
    // Start periodic refresh
    wl.refreshTicker = time.NewTicker(5 * time.Second)
    wl.stopRefresh = make(chan struct{})

    go func() {
        for {
            select {
            case <-wl.refreshTicker.C:
                wl.refresh()
            case <-wl.stopRefresh:
                return
            }
        }
    }()

    // Initial load
    wl.refresh()
}

func (wl *WorkflowList) Stop() {
    if wl.refreshTicker != nil {
        wl.refreshTicker.Stop()
    }
    if wl.stopRefresh != nil {
        close(wl.stopRefresh)
    }
}
```

### 6. Add Loading and Error States

Create reusable loading/error components:

```go
// internal/ui/loading.go
package ui

func (t *Table) SetLoading(loading bool) {
    if loading {
        t.Clear()
        t.SetCell(0, 0, tview.NewTableCell(" Loading...").
            SetTextColor(ColorFgDim).
            SetSelectable(false))
    }
}

func (t *Table) SetError(err error) {
    t.Clear()
    t.SetCell(0, 0, tview.NewTableCell(" "+IconFailed+" Error: "+err.Error()).
        SetTextColor(ColorFailed).
        SetSelectable(false))
}
```

## Main Entry Point

Update `cmd/main.go` to initialize with real provider:

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/atterpac/loom/internal/temporal"
    "github.com/atterpac/loom/internal/view"
)

func main() {
    hostPort := flag.String("address", "localhost:7233", "Temporal server address")
    namespace := flag.String("namespace", "default", "Default namespace")
    flag.Parse()

    // Create Temporal client
    provider, err := temporal.NewClient(*hostPort, *namespace)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
        os.Exit(1)
    }
    defer provider.Close()

    // Create and run app
    app := view.NewApp(provider)
    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Status Mapping

Map Temporal workflow statuses to UI status colors:

```go
// internal/temporal/status.go
package temporal

import "go.temporal.io/api/enums/v1"

func MapWorkflowStatus(status enums.WorkflowExecutionStatus) string {
    switch status {
    case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
        return "Running"
    case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
        return "Completed"
    case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
        return "Failed"
    case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
        return "Canceled"
    case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
        return "Terminated"
    case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
        return "TimedOut"
    default:
        return "Unknown"
    }
}
```

The UI `styles.go` already handles these status strings for coloring.

## Keybinding Patterns

Add action handlers that interact with Temporal:

```go
func (wd *WorkflowDetail) Start() {
    wd.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Rune() {
        case 'c': // Cancel workflow
            wd.confirmCancel()
            return nil
        case 't': // Terminate workflow
            wd.confirmTerminate()
            return nil
        case 's': // Send signal
            wd.showSignalDialog()
            return nil
        case 'r': // Refresh
            wd.refresh()
            return nil
        }
        return event
    })
}

func (wd *WorkflowDetail) confirmCancel() {
    // Show confirmation modal, then:
    go func() {
        ctx := context.Background()
        err := wd.provider.CancelWorkflow(ctx, wd.namespace, wd.workflowID, wd.runID)
        wd.app.UI().QueueUpdateDraw(func() {
            if err != nil {
                wd.showError(err)
            } else {
                wd.refresh()
            }
        })
    }()
}
```

## File Structure After Integration

```
loom/
├── cmd/
│   └── main.go                 # CLI flags, provider init
├── internal/
│   ├── ui/                     # Reusable primitives (unchanged)
│   │   ├── app.go
│   │   ├── table.go
│   │   ├── styles.go
│   │   └── ...
│   ├── view/                   # Views (inject provider)
│   │   ├── app.go
│   │   ├── namespace_list.go
│   │   └── ...
│   ├── temporal/               # NEW: Temporal integration
│   │   ├── provider.go         # Interface definition
│   │   ├── client.go           # SDK implementation
│   │   ├── status.go           # Status mapping
│   │   └── mock.go             # Mock for testing
│   └── mock/                   # Can be removed or kept for testing
│       └── data.go
├── go.mod
└── go.sum
```

## Dependencies

Add to `go.mod`:

```
go get go.temporal.io/sdk
go get go.temporal.io/api
```
