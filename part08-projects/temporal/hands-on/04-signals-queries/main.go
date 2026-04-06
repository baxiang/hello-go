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

// CounterWorkflow 演示信号和查询的基本使用
type CounterState struct {
	Value   int
	History []int
}

func CounterWorkflow(ctx workflow.Context, initialValue int) (CounterState, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CounterWorkflow started", "initialValue", initialValue)

	state := CounterState{
		Value:   initialValue,
		History: []int{initialValue},
	}

	// 定义信号通道
	incrementCh := workflow.GetSignalChannel(ctx, "increment")
	decrementCh := workflow.GetSignalChannel(ctx, "decrement")
	resetCh := workflow.GetSignalChannel(ctx, "reset")

	// 注册查询处理器
	err := workflow.SetQueryHandler(ctx, "get_state", func() (CounterState, error) {
		return state, nil
	})
	if err != nil {
		return state, err
	}

	err = workflow.SetQueryHandler(ctx, "get_value", func() (int, error) {
		return state.Value, nil
	})
	if err != nil {
		return state, err
	}

	// 使用选择器处理信号
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(incrementCh, func(c workflow.ReceiveChannel, more bool) {
		var delta int
		c.Receive(ctx, &delta)
		state.Value += delta
		state.History = append(state.History, state.Value)
		logger.Info("Incremented", "delta", delta, "new_value", state.Value)
	})

	selector.AddReceive(decrementCh, func(c workflow.ReceiveChannel, more bool) {
		var delta int
		c.Receive(ctx, &delta)
		state.Value -= delta
		state.History = append(state.History, state.Value)
		logger.Info("Decremented", "delta", delta, "new_value", state.Value)
	})

	selector.AddReceive(resetCh, func(c workflow.ReceiveChannel, more bool) {
		var newValue int
		c.Receive(ctx, &newValue)
		state.Value = newValue
		state.History = append(state.History, state.Value)
		logger.Info("Reset", "new_value", state.Value)
	})

	// 处理信号直到完成
	for {
		selector.Select(ctx)

		// 可以添加完成条件
		if state.Value >= 100 {
			logger.Info("Counter reached 100, completing workflow")
			break
		}
	}

	return state, nil
}

// OrderWorkflow 演示更复杂的信号处理场景
type OrderState struct {
	OrderID    string
	Status     string // "pending", "approved", "rejected", "completed"
	Total      float64
	ApprovedBy string
	Comments   []string
}

func OrderWorkflow(ctx workflow.Context, orderID string, total float64) (OrderState, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderWorkflow started", "orderID", orderID, "total", total)

	state := OrderState{
		OrderID:  orderID,
		Status:   "pending",
		Total:    total,
		Comments: []string{},
	}

	// 信号通道
	approveCh := workflow.GetSignalChannel(ctx, "approve")
	rejectCh := workflow.GetSignalChannel(ctx, "reject")
	commentCh := workflow.GetSignalChannel(ctx, "comment")

	// 查询处理器
	err := workflow.SetQueryHandler(ctx, "get_order", func() (OrderState, error) {
		return state, nil
	})
	if err != nil {
		return state, err
	}

	// 使用选择器处理信号
	selector := workflow.NewSelector(ctx)
	var completed bool

	selector.AddReceive(approveCh, func(c workflow.ReceiveChannel, more bool) {
		var approver string
		c.Receive(ctx, &approver)
		state.Status = "approved"
		state.ApprovedBy = approver
		state.Comments = append(state.Comments, fmt.Sprintf("Approved by %s", approver))
		logger.Info("Order approved", "approver", approver)
		completed = true
	})

	selector.AddReceive(rejectCh, func(c workflow.ReceiveChannel, more bool) {
		var reason string
		c.Receive(ctx, &reason)
		state.Status = "rejected"
		state.Comments = append(state.Comments, fmt.Sprintf("Rejected: %s", reason))
		logger.Info("Order rejected", "reason", reason)
		completed = true
	})

	selector.AddReceive(commentCh, func(c workflow.ReceiveChannel, more bool) {
		var comment string
		c.Receive(ctx, &comment)
		state.Comments = append(state.Comments, comment)
		logger.Info("Comment added", "comment", comment)
	})

	// 等待批准或拒绝
	for !completed {
		selector.Select(ctx)
	}

	// 如果批准，等待完成
	if state.Status == "approved" {
		completeCh := workflow.GetSignalChannel(ctx, "complete")
		selector.AddReceive(completeCh, func(c workflow.ReceiveChannel, more bool) {
			var dummy struct{}
			c.Receive(ctx, &dummy)
			state.Status = "completed"
			logger.Info("Order completed")
		})

		for state.Status != "completed" {
			selector.Select(ctx)
		}
	}

	return state, nil
}

// TimerSignalWorkflow 演示信号和定时器的组合使用
type TimerSignalState struct {
	ReceivedSignals []string
	TimedOut        bool
}

func TimerSignalWorkflow(ctx workflow.Context, timeout time.Duration) (TimerSignalState, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TimerSignalWorkflow started", "timeout", timeout)

	state := TimerSignalState{
		ReceivedSignals: []string{},
		TimedOut:        false,
	}

	signalCh := workflow.GetSignalChannel(ctx, "message")
	timerFuture := workflow.NewTimer(ctx, timeout)

	selector := workflow.NewSelector(ctx)

	selector.AddReceive(signalCh, func(c workflow.ReceiveChannel, more bool) {
		var msg string
		c.Receive(ctx, &msg)
		state.ReceivedSignals = append(state.ReceivedSignals, msg)
		logger.Info("Signal received", "message", msg)
	})

	selector.AddFuture(timerFuture, func(f workflow.Future) {
		state.TimedOut = true
		logger.Info("Timer expired")
	})

	// 处理信号直到定时器到期
	for !state.TimedOut {
		selector.Select(ctx)
	}

	return state, nil
}

// UpdateWorkflow 演示使用 Update（更新）功能（Go SDK 1.21+）
type UpdateState struct {
	Counter    int
	LastUpdate string
}

func UpdateWorkflow(ctx workflow.Context, initial int) (UpdateState, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("UpdateWorkflow started")

	state := UpdateState{
		Counter:    initial,
		LastUpdate: "",
	}

	// 查询处理器
	err := workflow.SetQueryHandler(ctx, "get_counter", func() (int, error) {
		return state.Counter, nil
	})
	if err != nil {
		return state, err
	}

	// 信号处理器 - 用于增量更新
	signalCh := workflow.GetSignalChannel(ctx, "update")
	for {
		var delta int
		signalCh.Receive(ctx, &delta)
		state.Counter += delta
		state.LastUpdate = fmt.Sprintf("Added %d at %v", delta, workflow.Now(ctx))
		logger.Info("Counter updated", "counter", state.Counter, "delta", delta)
	}
}

// ============================================================
// Worker 和 Starter
// ============================================================

const taskQueue = "signals-queries-task-queue"

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
	w.RegisterWorkflow(CounterWorkflow)
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterWorkflow(TimerSignalWorkflow)
	w.RegisterWorkflow(UpdateWorkflow)

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

	// 示例 1: Counter 工作流 - 信号和查询
	fmt.Println("\n=== 示例 1: Counter 工作流 ===")
	workflowOptions := client.StartWorkflowOptions{
		ID:        "counter-workflow-001",
		TaskQueue: taskQueue,
	}
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, CounterWorkflow, 0)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	fmt.Printf("Started workflow: %s\n", we.GetID())

	// 发送信号
	time.Sleep(1 * time.Second)
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "increment", 10)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Sent increment signal with value 10")

	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "increment", 20)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Sent increment signal with value 20")

	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "decrement", 5)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Sent decrement signal with value 5")

	// 查询状态
	time.Sleep(1 * time.Second)
	var state CounterState
	resp, err := c.QueryWorkflow(context.Background(), we.GetID(), "", "get_state")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	err = resp.Get(&state)
	if err != nil {
		log.Fatalln("Unable to decode query result", err)
	}
	fmt.Printf("Current state: Value=%d, History=%v\n", state.Value, state.History)

	// 完成工作流（达到 100）
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "increment", 100)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	// 等待工作流完成
	err = we.Get(context.Background(), &state)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Final state: Value=%d, History=%v\n", state.Value, state.History)

	// 示例 2: Order 工作流 - 复杂信号处理
	fmt.Println("\n=== 示例 2: Order 工作流 ===")
	workflowOptions.ID = "order-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, OrderWorkflow, "ORD-001", 199.99)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	fmt.Printf("Started order workflow: %s\n", we.GetID())

	// 添加评论
	time.Sleep(1 * time.Second)
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "comment", "Waiting for manager approval")
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Added comment")

	// 查询订单状态
	resp, err = c.QueryWorkflow(context.Background(), we.GetID(), "", "get_order")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var orderState OrderState
	err = resp.Get(&orderState)
	if err != nil {
		log.Fatalln("Unable to decode query result", err)
	}
	fmt.Printf("Order status: %s, Total: %.2f\n", orderState.Status, orderState.Total)

	// 批准订单
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "approve", "manager@example.com")
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Order approved")

	// 完成订单
	time.Sleep(1 * time.Second)
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "complete", struct{}{})
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	err = we.Get(context.Background(), &orderState)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Final order state: %+v\n", orderState)

	// 示例 3: Timer + Signal 工作流
	fmt.Println("\n=== 示例 3: Timer + Signal 工作流 ===")
	workflowOptions.ID = "timer-signal-workflow-001"
	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, TimerSignalWorkflow, 3*time.Second)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	fmt.Printf("Started timer workflow: %s\n", we.GetID())

	// 发送多个信号
	time.Sleep(500 * time.Millisecond)
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "message", "Hello")
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Sent message: Hello")

	time.Sleep(500 * time.Millisecond)
	err = c.SignalWorkflow(context.Background(), we.GetID(), "", "message", "World")
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	fmt.Println("Sent message: World")

	// 等待工作流完成
	var timerState TimerSignalState
	err = we.Get(context.Background(), &timerState)
	if err != nil {
		log.Fatalln("Workflow failed", err)
	}
	fmt.Printf("Timer workflow completed: Signals=%v, TimedOut=%v\n", timerState.ReceivedSignals, timerState.TimedOut)
}

func main() {
	fmt.Println("Temporal 信号与查询示例")
	fmt.Println("======================")
	fmt.Println("本示例演示：")
	fmt.Println("1. 发送信号到 Workflow")
	fmt.Println("2. 查询 Workflow 状态")
	fmt.Println("3. 使用 Selector 处理多个信号")
	fmt.Println("4. 定时器与信号的组合")
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
