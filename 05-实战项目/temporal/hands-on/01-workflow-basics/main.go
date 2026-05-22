package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// ============================================================
// Workflow 定义
// ============================================================

// SimpleWorkflow 演示最基本的工作流
func SimpleWorkflow(ctx workflow.Context, name string) (string, error) {
	// 日志输出（注意：Workflow 中使用 workflow.GetLogger）
	logger := workflow.GetLogger(ctx)
	logger.Info("SimpleWorkflow started", "name", name)

	// 配置 Activity 选项
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 执行 Activity
	var result string
	err := workflow.ExecuteActivity(ctx, GreetActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	logger.Info("SimpleWorkflow completed", "result", result)
	return result, nil
}

// TimedWorkflow 演示使用定时器的工作流
func TimedWorkflow(ctx workflow.Context, duration time.Duration) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TimedWorkflow started", "duration", duration)

	// 创建定时器
	timerFuture := workflow.NewTimer(ctx, duration)

	// 使用选择器等待定时器或信号
	var timerFired bool
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		timerFired = true
		logger.Info("Timer fired")
	})

	selector.Select(ctx)

	if timerFired {
		return fmt.Sprintf("Timer completed after %v", duration), nil
	}
	return "Workflow interrupted", nil
}

// ChildWorkflowDemo 演示子工作流调用
func ChildWorkflowDemo(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ChildWorkflowDemo started")

	// 配置子工作流选项
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowID:          "child-workflow-" + name,
		WorkflowRunTimeout:  30 * time.Second,
		WorkflowTaskTimeout: 10 * time.Second,
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	// 执行子工作流
	var result string
	err := workflow.ExecuteChildWorkflow(ctx, SimpleWorkflow, name).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Child workflow result: %s", result), nil
}

// ============================================================
// Activity 实现
// ============================================================

// GreetActivity 简单的问候 Activity
func GreetActivity(name string) (string, error) {
	return fmt.Sprintf("Hello, %s! Welcome to Temporal.", name), nil
}

// ============================================================
// Worker 和 Starter
// ============================================================

const taskQueue = "workflow-basics-task-queue"

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
	w.RegisterWorkflow(SimpleWorkflow)
	w.RegisterWorkflow(TimedWorkflow)
	w.RegisterWorkflow(ChildWorkflowDemo)

	// 注册 Activity
	w.RegisterActivity(GreetActivity)

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

	// 示例 1: 简单工作流
	fmt.Println("\n=== 示例 1: 简单工作流 ===")
	workflowOptions := client.StartWorkflowOptions{
		ID:        "simple-workflow-001",
		TaskQueue: taskQueue,
	}
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, SimpleWorkflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)
	fmt.Printf("WorkflowID: %s, RunID: %s\n", we.GetID(), we.GetRunID())

	// 示例 2: 带定时器的工作流
	fmt.Println("\n=== 示例 2: 带定时器的工作流 ===")
	workflowOptions.ID = "timed-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, TimedWorkflow, 2*time.Second)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Result: %s\n", result)

	// 示例 3: 子工作流
	fmt.Println("\n=== 示例 3: 子工作流 ===")
	workflowOptions.ID = "child-workflow-demo-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, ChildWorkflowDemo, "Child")
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
	fmt.Println("Temporal Workflow 基础示例")
	fmt.Println("========================")
	fmt.Println("本示例演示：")
	fmt.Println("1. 简单工作流的定义和执行")
	fmt.Println("2. 定时器的使用")
	fmt.Println("3. 子工作流的调用")
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
