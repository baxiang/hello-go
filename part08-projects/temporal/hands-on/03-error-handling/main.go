package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// ============================================================
// 自定义错误类型
// ============================================================

var (
	ErrItemNotFound   = errors.New("item not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrServiceTimeout = errors.New("service timeout")
	ErrRetryableError = temporal.NewApplicationError("retryable error", "RetryableError")
	ErrNonRetryable   = temporal.NewNonRetryableApplicationError("non-retryable error", "NonRetryableError", nil)
)

// ============================================================
// Workflow 定义
// ============================================================

// RetryWorkflow 演示错误重试策略
func RetryWorkflow(ctx workflow.Context, input string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("RetryWorkflow started", "input", input)

	// 配置重试策略
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, FlakeyActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed after retries", "error", err)
		return "", err
	}

	return result, nil
}

// TimeoutWorkflow 演示超时处理
func TimeoutWorkflow(ctx workflow.Context, duration time.Duration) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TimeoutWorkflow started", "duration", duration)

	// 配置超时
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Second, // 活动必须在2秒内完成
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SlowActivity, duration).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed due to timeout", "error", err)

		// 检查是否是超时错误
		var timeoutErr *temporal.TimeoutError
		if errors.As(err, &timeoutErr) {
			return "", fmt.Errorf("activity timed out: %w", err)
		}
		return "", err
	}

	return result, nil
}

// CompensationWorkflow 演示补偿事务模式
func CompensationWorkflow(ctx workflow.Context, operations []string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CompensationWorkflow started")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 记录已完成的操作，用于补偿
	var completedOps []string

	// 执行操作序列
	for i, op := range operations {
		var result string
		err := workflow.ExecuteActivity(ctx, ExecuteOperationActivity, op).Get(ctx, &result)
		if err != nil {
			logger.Error("Operation failed, starting compensation", "op", op, "error", err)

			// 执行补偿：逆序回滚已完成操作
			for j := len(completedOps) - 1; j >= 0; j-- {
				compErr := workflow.ExecuteActivity(ctx, CompensateActivity, completedOps[j]).Get(ctx, nil)
				if compErr != nil {
					logger.Error("Compensation failed", "op", completedOps[j], "error", compErr)
				}
			}

			return "", fmt.Errorf("operation %s failed, compensation executed", op)
		}

		completedOps = append(completedOps, op)
		logger.Info("Operation completed", "index", i, "op", op)
	}

	return fmt.Sprintf("All %d operations completed successfully", len(completedOps)), nil
}

// CircuitBreakerWorkflow 演示熔断器模式
func CircuitBreakerWorkflow(ctx workflow.Context, serviceURL string, attempts int) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CircuitBreakerWorkflow started", "url", serviceURL, "attempts", attempts)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 1.5,
			MaximumInterval:    time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var failureCount int
	var lastError error

	for i := 0; i < attempts; i++ {
		var result string
		err := workflow.ExecuteActivity(ctx, ServiceCallActivity, serviceURL).Get(ctx, &result)
		if err != nil {
			failureCount++
			lastError = err
			logger.Warn("Service call failed", "attempt", i+1, "error", err)

			// 连续3次失败后返回错误
			if failureCount >= 3 {
				return "", fmt.Errorf("circuit breaker triggered after %d failures: %w", failureCount, lastError)
			}
			continue
		}

		// 成功后重置计数
		failureCount = 0
		return result, nil
	}

	return "", fmt.Errorf("all attempts failed: %w", lastError)
}

// SagaWorkflow 演示 Saga 分布式事务模式
type SagaStep struct {
	Name        string
	ExecuteArgs interface{}
	Compensate  string
}

func SagaWorkflow(ctx workflow.Context, steps []SagaStep) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SagaWorkflow started")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var executedSteps []string

	for _, step := range steps {
		logger.Info("Executing step", "name", step.Name)

		var result string
		err := workflow.ExecuteActivity(ctx, SagaExecuteActivity, step.Name, step.ExecuteArgs).Get(ctx, &result)
		if err != nil {
			logger.Error("Step failed, starting saga compensation", "step", step.Name, "error", err)

			// 逆序执行补偿
			for i := len(executedSteps) - 1; i >= 0; i-- {
				compStep := executedSteps[i]
				compErr := workflow.ExecuteActivity(ctx, SagaCompensateActivity, compStep).Get(ctx, nil)
				if compErr != nil {
					logger.Error("Compensation failed", "step", compStep, "error", compErr)
				}
			}

			return "", fmt.Errorf("saga failed at step %s: %w", step.Name, err)
		}

		executedSteps = append(executedSteps, step.Name)
	}

	return fmt.Sprintf("Saga completed successfully with %d steps", len(executedSteps)), nil
}

// ============================================================
// Activity 实现
// ============================================================

var callCount int

// FlakeyActivity 模拟不稳定的服务
func FlakeyActivity(ctx context.Context, input string) (string, error) {
	callCount++
	// 前2次调用失败，第3次成功
	if callCount <= 2 {
		return "", ErrRetryableError
	}
	return fmt.Sprintf("Success after %d attempts: %s", callCount, input), nil
}

// SlowActivity 模拟慢操作
func SlowActivity(ctx context.Context, duration time.Duration) (string, error) {
	time.Sleep(duration)
	return fmt.Sprintf("Completed after %v", duration), nil
}

// ExecuteOperationActivity 执行操作
func ExecuteOperationActivity(ctx context.Context, op string) (string, error) {
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("Executed: %s", op), nil
}

// CompensateActivity 执行补偿操作
func CompensateActivity(ctx context.Context, op string) error {
	log.Printf("Compensating: %s", op)
	time.Sleep(50 * time.Millisecond)
	return nil
}

// ServiceCallActivity 模拟服务调用
func ServiceCallActivity(ctx context.Context, url string) (string, error) {
	// 模拟服务不可用
	return "", fmt.Errorf("service %s unavailable", url)
}

// SagaExecuteActivity Saga 步骤执行
func SagaExecuteActivity(ctx context.Context, stepName string, args interface{}) (string, error) {
	time.Sleep(100 * time.Millisecond)

	// 模拟某个步骤失败
	if stepName == "step3" {
		return "", fmt.Errorf("step3 failed intentionally")
	}

	return fmt.Sprintf("Executed %s", stepName), nil
}

// SagaCompensateActivity Saga 步骤补偿
func SagaCompensateActivity(ctx context.Context, stepName string) error {
	log.Printf("Saga compensating: %s", stepName)
	time.Sleep(50 * time.Millisecond)
	return nil
}

// ============================================================
// Worker 和 Starter
// ============================================================

const taskQueue = "error-handling-task-queue"

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
	w.RegisterWorkflow(RetryWorkflow)
	w.RegisterWorkflow(TimeoutWorkflow)
	w.RegisterWorkflow(CompensationWorkflow)
	w.RegisterWorkflow(CircuitBreakerWorkflow)
	w.RegisterWorkflow(SagaWorkflow)

	// 注册 Activity
	w.RegisterActivity(FlakeyActivity)
	w.RegisterActivity(SlowActivity)
	w.RegisterActivity(ExecuteOperationActivity)
	w.RegisterActivity(CompensateActivity)
	w.RegisterActivity(ServiceCallActivity)
	w.RegisterActivity(SagaExecuteActivity)
	w.RegisterActivity(SagaCompensateActivity)

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

	// 示例 1: 重试策略
	fmt.Println("\n=== 示例 1: 重试策略 ===")
	callCount = 0 // 重置计数器
	workflowOptions := client.StartWorkflowOptions{
		ID:        "retry-workflow-001",
		TaskQueue: taskQueue,
	}
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, RetryWorkflow, "test-input")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		fmt.Printf("Workflow failed (expected): %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}

	// 示例 2: 超时处理
	fmt.Println("\n=== 示例 2: 超时处理 ===")
	workflowOptions.ID = "timeout-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, TimeoutWorkflow, 5*time.Second)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		fmt.Printf("Workflow failed (expected timeout): %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}

	// 示例 3: 补偿事务
	fmt.Println("\n=== 示例 3: 补偿事务 ===")
	workflowOptions.ID = "compensation-workflow-001"
	ops := []string{"step1", "step2", "fail_step", "step4"}
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, CompensationWorkflow, ops)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		fmt.Printf("Workflow failed (compensation executed): %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}

	// 示例 4: 熔断器模式
	fmt.Println("\n=== 示例 4: 熔断器模式 ===")
	workflowOptions.ID = "circuit-breaker-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, CircuitBreakerWorkflow, "http://api.example.com", 5)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		fmt.Printf("Workflow failed (circuit breaker triggered): %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}

	// 示例 5: Saga 模式
	fmt.Println("\n=== 示例 5: Saga 模式 ===")
	workflowOptions.ID = "saga-workflow-001"
	sagaSteps := []SagaStep{
		{Name: "step1", ExecuteArgs: "arg1", Compensate: "comp1"},
		{Name: "step2", ExecuteArgs: "arg2", Compensate: "comp2"},
		{Name: "step3", ExecuteArgs: "arg3", Compensate: "comp3"},
		{Name: "step4", ExecuteArgs: "arg4", Compensate: "comp4"},
	}
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, SagaWorkflow, sagaSteps)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(context.Background(), &result)
	if err != nil {
		fmt.Printf("Workflow failed (saga compensation executed): %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
}

func main() {
	fmt.Println("Temporal 错误处理示例")
	fmt.Println("====================")
	fmt.Println("本示例演示：")
	fmt.Println("1. 重试策略配置")
	fmt.Println("2. 超时处理")
	fmt.Println("3. 补偿事务模式")
	fmt.Println("4. 熔断器模式")
	fmt.Println("5. Saga 分布式事务")
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
