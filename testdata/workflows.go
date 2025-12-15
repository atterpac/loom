package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// === WORKFLOWS ===

// QuickWorkflow completes immediately
func QuickWorkflow(ctx workflow.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}

// SlowWorkflow takes some time to complete
func SlowWorkflow(ctx workflow.Context, seconds int) (string, error) {
	if seconds <= 0 {
		seconds = 5
	}
	err := workflow.Sleep(ctx, time.Duration(seconds)*time.Second)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Slept for %d seconds", seconds), nil
}

// FailingWorkflow always fails
func FailingWorkflow(ctx workflow.Context, message string) (string, error) {
	return "", temporal.NewApplicationError(message, "INTENTIONAL_FAILURE")
}

// ActivityWorkflow runs multiple activities
func ActivityWorkflow(ctx workflow.Context, count int) ([]string, error) {
	if count <= 0 {
		count = 3
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var results []string
	for i := 1; i <= count; i++ {
		var result string
		err := workflow.ExecuteActivity(ctx, ProcessActivity, i).Get(ctx, &result)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// TimerWorkflow uses timers
func TimerWorkflow(ctx workflow.Context, intervals []int) (string, error) {
	if len(intervals) == 0 {
		intervals = []int{1, 2, 1}
	}

	for i, secs := range intervals {
		timerFuture := workflow.NewTimer(ctx, time.Duration(secs)*time.Second)
		err := timerFuture.Get(ctx, nil)
		if err != nil {
			return "", err
		}
		workflow.GetLogger(ctx).Info("Timer fired", "iteration", i+1)
	}

	return fmt.Sprintf("Completed %d timers", len(intervals)), nil
}

// SignalWorkflow waits for signals
func SignalWorkflow(ctx workflow.Context) ([]string, error) {
	var signals []string
	signalChan := workflow.GetSignalChannel(ctx, "test-signal")

	// Wait for up to 3 signals or 60 seconds
	selector := workflow.NewSelector(ctx)
	timerFuture := workflow.NewTimer(ctx, 60*time.Second)

	done := false
	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		var signal string
		c.Receive(ctx, &signal)
		signals = append(signals, signal)
		if len(signals) >= 3 {
			done = true
		}
	})
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		done = true
	})

	for !done {
		selector.Select(ctx)
	}

	return signals, nil
}

// ChildWorkflow spawns child workflows
func ChildWorkflow(ctx workflow.Context, childCount int) ([]string, error) {
	if childCount <= 0 {
		childCount = 2
	}

	var results []string
	for i := 1; i <= childCount; i++ {
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("%s-child-%d", workflow.GetInfo(ctx).WorkflowExecution.ID, i),
		})

		var result string
		err := workflow.ExecuteChildWorkflow(childCtx, QuickWorkflow, fmt.Sprintf("Child-%d", i)).Get(ctx, &result)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// LongRunningWorkflow runs for a long time (for testing "Running" status)
func LongRunningWorkflow(ctx workflow.Context, minutes int) (string, error) {
	if minutes <= 0 {
		minutes = 5
	}

	// Run activities periodically
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 0; i < minutes; i++ {
		var result string
		err := workflow.ExecuteActivity(ctx, HeartbeatActivity, i).Get(ctx, &result)
		if err != nil {
			return "", err
		}
		err = workflow.Sleep(ctx, 1*time.Minute)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Ran for %d minutes", minutes), nil
}

// RetryWorkflow fails a few times then succeeds
func RetryWorkflow(ctx workflow.Context, failCount int) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.5,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, FlakeyActivity, failCount).Get(ctx, &result)
	return result, err
}

// === ACTIVITIES ===

// ProcessActivity simulates processing work
func ProcessActivity(ctx context.Context, iteration int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Processing", "iteration", iteration)

	// Simulate work
	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)

	return fmt.Sprintf("Processed item %d", iteration), nil
}

// HeartbeatActivity reports progress
func HeartbeatActivity(ctx context.Context, minute int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Heartbeat", "minute", minute)

	// Send heartbeats while "working"
	for i := 0; i < 5; i++ {
		activity.RecordHeartbeat(ctx, fmt.Sprintf("tick %d", i))
		time.Sleep(time.Second)
	}

	return fmt.Sprintf("Heartbeat minute %d complete", minute), nil
}

// FlakeyActivity fails N times before succeeding
var flakeyAttempts = make(map[string]int)

func FlakeyActivity(ctx context.Context, failCount int) (string, error) {
	info := activity.GetInfo(ctx)
	key := info.WorkflowExecution.ID

	flakeyAttempts[key]++
	attempt := flakeyAttempts[key]

	if attempt <= failCount {
		return "", errors.New(fmt.Sprintf("intentional failure %d/%d", attempt, failCount))
	}

	delete(flakeyAttempts, key)
	return fmt.Sprintf("Succeeded after %d failures", failCount), nil
}

// SlowActivity takes a while
func SlowActivity(ctx context.Context, seconds int) (string, error) {
	time.Sleep(time.Duration(seconds) * time.Second)
	return fmt.Sprintf("Slow activity completed after %d seconds", seconds), nil
}
