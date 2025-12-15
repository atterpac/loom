package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	taskQueue = "test-task-queue"
)

var (
	mode      = flag.String("mode", "both", "Mode: worker, starter, or both")
	address   = flag.String("address", "localhost:7233", "Temporal server address")
	namespace = flag.String("namespace", "default", "Temporal namespace")
	count     = flag.Int("count", 10, "Number of workflows to start")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	c, err := client.Dial(client.Options{
		HostPort:  *address,
		Namespace: *namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	switch *mode {
	case "worker":
		wg.Add(1)
		go func() {
			defer wg.Done()
			runWorker(ctx, c)
		}()
	case "starter":
		startWorkflows(ctx, c, *count)
		return
	case "both":
		wg.Add(1)
		go func() {
			defer wg.Done()
			runWorker(ctx, c)
		}()
		// Give worker time to start
		time.Sleep(2 * time.Second)
		startWorkflows(ctx, c, *count)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down...")
	cancel()
	wg.Wait()
}

func runWorker(ctx context.Context, c client.Client) {
	w := worker.New(c, taskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(QuickWorkflow)
	w.RegisterWorkflow(SlowWorkflow)
	w.RegisterWorkflow(FailingWorkflow)
	w.RegisterWorkflow(ActivityWorkflow)
	w.RegisterWorkflow(TimerWorkflow)
	w.RegisterWorkflow(SignalWorkflow)
	w.RegisterWorkflow(ChildWorkflow)
	w.RegisterWorkflow(LongRunningWorkflow)
	w.RegisterWorkflow(RetryWorkflow)

	// Register activities
	w.RegisterActivity(ProcessActivity)
	w.RegisterActivity(HeartbeatActivity)
	w.RegisterActivity(FlakeyActivity)
	w.RegisterActivity(SlowActivity)

	fmt.Printf("Starting worker on task queue: %s\n", taskQueue)

	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Printf("Worker stopped: %v", err)
	}
}

func startWorkflows(ctx context.Context, c client.Client, count int) {
	fmt.Printf("Starting %d test workflows...\n\n", count)

	workflowTypes := []struct {
		name    string
		starter func(ctx context.Context, c client.Client, id string) error
		weight  int // Higher = more likely to be chosen
	}{
		{"QuickWorkflow", startQuickWorkflow, 3},
		{"SlowWorkflow", startSlowWorkflow, 2},
		{"FailingWorkflow", startFailingWorkflow, 2},
		{"ActivityWorkflow", startActivityWorkflow, 3},
		{"TimerWorkflow", startTimerWorkflow, 2},
		{"ChildWorkflow", startChildWorkflow, 2},
		{"RetryWorkflow", startRetryWorkflow, 2},
		{"LongRunningWorkflow", startLongRunningWorkflow, 1},
		{"SignalWorkflow", startSignalWorkflow, 1},
	}

	// Build weighted list
	var weighted []int
	for i, wt := range workflowTypes {
		for j := 0; j < wt.weight; j++ {
			weighted = append(weighted, i)
		}
	}

	started := 0
	failed := 0

	for i := 0; i < count; i++ {
		// Pick random workflow type
		idx := weighted[rand.Intn(len(weighted))]
		wt := workflowTypes[idx]

		id := fmt.Sprintf("test-%s-%d-%s", wt.name, i+1, randomString(6))

		err := wt.starter(ctx, c, id)
		if err != nil {
			fmt.Printf("  [FAIL] %s: %v\n", id, err)
			failed++
		} else {
			fmt.Printf("  [OK] %s\n", id)
			started++
		}

		// Small delay between starts
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\nStarted: %d, Failed: %d\n", started, failed)
	fmt.Println("\nWorkflows are running. Press Ctrl+C to stop the worker.")
}

func startQuickWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, QuickWorkflow, randomName())
	return err
}

func startSlowWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, SlowWorkflow, rand.Intn(10)+3)
	return err
}

func startFailingWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, FailingWorkflow, "This workflow intentionally fails")
	return err
}

func startActivityWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, ActivityWorkflow, rand.Intn(5)+2)
	return err
}

func startTimerWorkflow(ctx context.Context, c client.Client, id string) error {
	intervals := make([]int, rand.Intn(3)+2)
	for i := range intervals {
		intervals[i] = rand.Intn(3) + 1
	}
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, TimerWorkflow, intervals)
	return err
}

func startSignalWorkflow(ctx context.Context, c client.Client, id string) error {
	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, SignalWorkflow)
	if err != nil {
		return err
	}

	// Send some signals after a delay
	go func() {
		time.Sleep(2 * time.Second)
		for i := 0; i < 3; i++ {
			c.SignalWorkflow(ctx, we.GetID(), we.GetRunID(), "test-signal", fmt.Sprintf("signal-%d", i+1))
			time.Sleep(time.Second)
		}
	}()

	return nil
}

func startChildWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, ChildWorkflow, rand.Intn(3)+1)
	return err
}

func startLongRunningWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, LongRunningWorkflow, 2) // 2 minutes
	return err
}

func startRetryWorkflow(ctx context.Context, c client.Client, id string) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        id,
		TaskQueue: taskQueue,
	}, RetryWorkflow, rand.Intn(3)+1)
	return err
}

// Helper functions

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomName() string {
	names := []string{
		"Alice", "Bob", "Charlie", "Diana", "Eve",
		"Frank", "Grace", "Henry", "Ivy", "Jack",
		"Kate", "Leo", "Mia", "Noah", "Olivia",
	}
	return names[rand.Intn(len(names))]
}
