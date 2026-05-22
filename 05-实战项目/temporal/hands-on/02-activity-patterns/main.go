package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// ============================================================
// Workflow 定义
// ============================================================

// SequenceWorkflow 演示顺序执行多个 Activity
func SequenceWorkflow(ctx workflow.Context, items []string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SequenceWorkflow started", "items", items)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var results []string
	for i, item := range items {
		var result string
		err := workflow.ExecuteActivity(ctx, ProcessItemActivity, item, i).Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed", "index", i, "error", err)
			return "", err
		}
		results = append(results, result)
	}

	return fmt.Sprintf("Processed %d items: %v", len(results), results), nil
}

// ParallelWorkflow 演示并行执行多个 Activity
func ParallelWorkflow(ctx workflow.Context, items []string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ParallelWorkflow started", "items", items)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 启动所有 Activity（不等待）
	futures := make([]workflow.Future, len(items))
	for i, item := range items {
		futures[i] = workflow.ExecuteActivity(ctx, ProcessItemActivity, item, i)
	}

	// 等待所有 Activity 完成
	results := make([]string, len(items))
	for i, future := range futures {
		err := future.Get(ctx, &results[i])
		if err != nil {
			logger.Error("Activity failed", "index", i, "error", err)
			return "", err
		}
	}

	return fmt.Sprintf("Processed %d items in parallel: %v", len(results), results), nil
}

// SelectFirstWorkflow 演示使用 Selector 选择最快完成的 Activity
func SelectFirstWorkflow(ctx workflow.Context, urls []string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SelectFirstWorkflow started", "urls", urls)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 使用 Selector 等待第一个完成的 Activity
	selector := workflow.NewSelector(ctx)
	var result string
	var completedIndex int

	for i, url := range urls {
		// 注意：在循环中使用局部变量捕获
		idx := i
		future := workflow.ExecuteActivity(ctx, FetchURLActivity, url)
		selector.AddFuture(future, func(f workflow.Future) {
			err := f.Get(ctx, &result)
			if err == nil {
				completedIndex = idx
				logger.Info("Activity completed", "index", idx, "result", result)
			}
		})
	}

	// 只等待一个完成
	selector.Select(ctx)

	return fmt.Sprintf("First completed: %s (index %d)", result, completedIndex), nil
}

// BatchedWorkflow 演示分批执行 Activity
func BatchedWorkflow(ctx workflow.Context, items []string, batchSize int) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("BatchedWorkflow started", "total", len(items), "batchSize", batchSize)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var allResults []string

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		// 并行执行当前批次
		futures := make([]workflow.Future, len(batch))
		for j, item := range batch {
			futures[j] = workflow.ExecuteActivity(ctx, ProcessItemActivity, item, i+j)
		}

		// 等待当前批次完成
		for _, future := range futures {
			var result string
			err := future.Get(ctx, &result)
			if err != nil {
				logger.Error("Activity failed in batch", "error", err)
				return "", err
			}
			allResults = append(allResults, result)
		}
	}

	return fmt.Sprintf("Processed %d items in batches", len(allResults)), nil
}

// HeartbeatActivityWorkflow 演示带心跳的长时间 Activity
func HeartbeatActivityWorkflow(ctx workflow.Context, count int) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("HeartbeatActivityWorkflow started", "count", count)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, LongRunningActivity, count).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// ============================================================
// Activity 实现
// ============================================================

// ProcessItemActivity 处理单个项目
func ProcessItemActivity(ctx context.Context, item string, index int) (string, error) {
	time.Sleep(100 * time.Millisecond) // 模拟处理时间
	return fmt.Sprintf("processed-%s-%d", item, index), nil
}

// FetchURLActivity 模拟获取 URL 内容
func FetchURLActivity(ctx context.Context, url string) (string, error) {
	// 模拟不同的响应时间
	delay := time.Duration(100+len(url)*50) * time.Millisecond
	time.Sleep(delay)
	return fmt.Sprintf("content-from-%s", url), nil
}

// LongRunningActivity 演示长时间运行的 Activity（带心跳）
func LongRunningActivity(ctx context.Context, count int) (string, error) {
	for i := 0; i < count; i++ {
		// 检查是否需要取消
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// 模拟工作
		time.Sleep(500 * time.Millisecond)

		// 发送心跳
		activity.RecordHeartbeat(ctx, fmt.Sprintf("Progress: %d/%d", i+1, count))
	}

	return fmt.Sprintf("Completed %d iterations", count), nil
}

// ============================================================
// Worker 和 Starter
// ============================================================

const taskQueue = "activity-patterns-task-queue"

func runWorker() {
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})

	// 注册 Workflow
	w.RegisterWorkflow(SequenceWorkflow)
	w.RegisterWorkflow(ParallelWorkflow)
	w.RegisterWorkflow(SelectFirstWorkflow)
	w.RegisterWorkflow(BatchedWorkflow)
	w.RegisterWorkflow(HeartbeatActivityWorkflow)

	// 注册 Activity
	w.RegisterActivity(ProcessItemActivity)
	w.RegisterActivity(FetchURLActivity)
	w.RegisterActivity(LongRunningActivity)

	log.Println("Starting worker...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func runStarter() {
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	items := []string{"A", "B", "C", "D", "E"}
	urls := []string{"http://api1.com", "http://api2.com", "http://api3.com"}

	// 示例 1: 顺序执行
	fmt.Println("\n=== 示例 1: 顺序执行 ===")
	workflowOptions := client.StartWorkflowOptions{
		ID:        "sequence-workflow-001",
		TaskQueue: taskQueue,
	}
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, SequenceWorkflow, items)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)

	// 示例 2: 并行执行
	fmt.Println("\n=== 示例 2: 并行执行 ===")
	workflowOptions.ID = "parallel-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, ParallelWorkflow, items)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)

	// 示例 3: 选择最快完成
	fmt.Println("\n=== 示例 3: 选择最快完成 ===")
	workflowOptions.ID = "select-first-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, SelectFirstWorkflow, urls)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)

	// 示例 4: 分批执行
	fmt.Println("\n=== 示例 4: 分批执行 ===")
	workflowOptions.ID = "batched-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, BatchedWorkflow, append(items, "F", "G", "H"), 3)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)

	// 示例 5: 带心跳的长时间 Activity
	fmt.Println("\n=== 示例 5: 带心跳的长时间 Activity ===")
	workflowOptions.ID = "heartbeat-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, HeartbeatActivityWorkflow, 5)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)
}

func main() {
	fmt.Println("Temporal Activity 模式示例")
	fmt.Println("=========================")
	fmt.Println("本示例演示：")
	fmt.Println("1. 顺序执行多个 Activity")
	fmt.Println("2. 并行执行多个 Activity")
	fmt.Println("3. 使用 Selector 选择最快完成")
	fmt.Println("4. 分批执行 Activity")
	fmt.Println("5. 带心跳的长时间 Activity")
	fmt.Println()
	fmt.Println("使用方法：")
	fmt.Println("  作为 Worker 运行: go run main.go worker")
	fmt.Println("  作为 Starter 运行: go run main.go starter")
	fmt.Println()

	if len(os.Args) < 2 {
		fmt.Println("请指定运行模式: worker 或 starter")
		return
	}

	switch os.Args[1] {
	case "worker":
		runWorker()
	case "starter":
		runStarter()
	default:
		fmt.Printf("未知模式: %s\n", os.Args[1])
	}
}
