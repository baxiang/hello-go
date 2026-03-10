# Temporal 工作流模式

本节介绍 Temporal 中常用的工作流设计模式，包括顺序执行、并行执行、扇入扇出、状态机、Saga 模式等。

## 3.1 顺序执行模式

### 3.1.1 基本顺序执行

```go
// 顺序执行工作流：逐步执行一系列活动
func SequentialWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 步骤 1: 验证订单
    logger.Info("步骤 1: 验证订单")
    if err := workflow.ExecuteActivity(ctx, ValidateOrder, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 步骤 2: 预留库存
    logger.Info("步骤 2: 预留库存")
    if err := workflow.ExecuteActivity(ctx, ReserveInventory, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 步骤 3: 处理支付
    logger.Info("步骤 3: 处理支付")
    if err := workflow.ExecuteActivity(ctx, ProcessPayment, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 步骤 4: 发货
    logger.Info("步骤 4: 发货")
    if err := workflow.ExecuteActivity(ctx, ShipOrder, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 步骤 5: 发送通知
    logger.Info("步骤 5: 发送通知")
    if err := workflow.ExecuteActivity(ctx, SendNotification, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    logger.Info("订单处理完成", "orderID", orderID)
    return nil
}
```

### 3.1.2 带条件的顺序执行

```go
// 带条件的顺序执行
func ConditionalSequentialWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    
    // 验证订单
    if err := workflow.ExecuteActivity(ctx, ValidateOrder, order).Get(ctx, nil); err != nil {
        return err
    }
    
    // 根据订单类型执行不同流程
    if order.Type == "physical" {
        // 实物商品：预留库存 -> 发货
        if err := workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, nil); err != nil {
            return err
        }
        if err := workflow.ExecuteActivity(ctx, ShipOrder, order).Get(ctx, nil); err != nil {
            return err
        }
    } else if order.Type == "digital" {
        // 数字商品：发送下载链接
        if err := workflow.ExecuteActivity(ctx, SendDownloadLink, order).Get(ctx, nil); err != nil {
            return err
        }
    }
    
    // 处理支付
    if err := workflow.ExecuteActivity(ctx, ProcessPayment, order).Get(ctx, nil); err != nil {
        return err
    }
    
    // 发送通知
    return workflow.ExecuteActivity(ctx, SendNotification, order).Get(ctx, nil)
}
```

---

## 3.2 并行执行模式

### 3.2.1 并行执行活动

```go
// 并行执行多个活动
func ParallelActivityWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 并行执行多个活动
    f1 := workflow.ExecuteActivity(ctx, SendEmail, orderID)
    f2 := workflow.ExecuteActivity(ctx, SendSMS, orderID)
    f3 := workflow.ExecuteActivity(ctx, UpdateDashboard, orderID)
    
    // 等待所有活动完成
    if err := f1.Get(ctx, nil); err != nil {
        logger.Error("邮件发送失败", "error", err)
        return err
    }
    
    if err := f2.Get(ctx, nil); err != nil {
        logger.Error("短信发送失败", "error", err)
        return err
    }
    
    if err := f3.Get(ctx, nil); err != nil {
        logger.Error("仪表板更新失败", "error", err)
        return err
    }
    
    logger.Info("所有通知已发送")
    return nil
}
```

### 3.2.2 使用 Future

```go
// 使用 Future 进行更灵活的控制
func FutureWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 启动所有活动
    futures := map[string]workflow.Future{
        "email":    workflow.ExecuteActivity(ctx, SendEmail, orderID),
        "sms":      workflow.ExecuteActivity(ctx, SendSMS, orderID),
        "dashboard": workflow.ExecuteActivity(ctx, UpdateDashboard, orderID),
    }
    
    // 等待任意一个完成
    selector := workflow.NewSelector(ctx)
    for name, future := range futures {
        name := name
        future := future
        selector.AddFuture(future, func(f workflow.Future) {
            var err error
            f.Get(ctx, &err)
            if err != nil {
                logger.Error(name+" 失败", "error", err)
            } else {
                logger.Info(name+" 完成")
            }
        })
    }
    
    // 等待所有完成
    for range futures {
        selector.Select(ctx)
    }
    
    return nil
}
```

---

## 3.3 扇入扇出模式

### 3.3.1 扇出模式

```go
// 扇出：向多个接收者发送消息
func FanOutWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    
    // 为每个商品创建活动
    futures := make([]workflow.Future, len(order.Items))
    
    for i, item := range order.Items {
        futures[i] = workflow.ExecuteActivity(ctx, ProcessItem, item)
    }
    
    // 等待所有商品处理完成
    var errors []string
    for i, future := range futures {
        var err error
        future.Get(ctx, &err)
        if err != nil {
            errors = append(errors, fmt.Sprintf("商品 %d: %v", i, err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("处理失败: %s", strings.Join(errors, "; "))
    }
    
    logger.Info("所有商品处理完成")
    return nil
}
```

### 3.3.2 扇入模式

```go
// 扇入：收集多个活动的结果
func FanInWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 启动多个查询活动
    futures := []workflow.Future{
        workflow.ExecuteActivity(ctx, GetOrderDetails, orderID),
        workflow.ExecuteActivity(ctx, GetPaymentInfo, orderID),
        workflow.ExecuteActivity(ctx, GetShippingInfo, orderID),
    }
    
    // 收集结果
    var orderDetails OrderDetails
    var paymentInfo PaymentInfo
    var shippingInfo ShippingInfo
    
    if err := futures[0].Get(ctx, &orderDetails); err != nil {
        return err
    }
    if err := futures[1].Get(ctx, &paymentInfo); err != nil {
        return err
    }
    if err := futures[2].Get(ctx, &shippingInfo); err != nil {
        return err
    }
    
    // 聚合结果
    summary := OrderSummary{
        Details:   orderDetails,
        Payment:   paymentInfo,
        Shipping:  shippingInfo,
    }
    
    // 保存汇总
    return workflow.ExecuteActivity(ctx, SaveOrderSummary, summary).Get(ctx, nil)
}
```

### 3.3.3 动态扇出

```go
// 动态扇出：根据输入动态创建活动
func DynamicFanOutWorkflow(ctx workflow.Context, batch Batch) error {
    logger := workflow.GetLogger(ctx)
    
    // 根据批次大小动态创建活动
    futures := make([]workflow.Future, 0, len(batch.Items))
    
    for _, item := range batch.Items {
        future := workflow.ExecuteActivity(ctx, ProcessBatchItem, item)
        futures = append(futures, future)
    }
    
    // 等待所有活动完成
    completed := 0
    failed := 0
    
    for _, future := range futures {
        var err error
        future.Get(ctx, &err)
        if err != nil {
            failed++
        } else {
            completed++
        }
    }
    
    logger.Info("批处理完成", "completed", completed, "failed", failed)
    
    if failed > 0 {
        return fmt.Errorf("%d 个项目处理失败", failed)
    }
    
    return nil
}
```

---

## 3.4 状态机模式

### 3.4.1 订单状态机

```go
// 订单状态机工作流
func OrderStateMachineWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 初始状态
    state := "created"
    
    for {
        switch state {
        case "created":
            logger.Info("订单已创建，等待支付")
            // 等待支付信号
            signal := workflow.NewSignalChannel(ctx, "payment-received")
            var paymentInfo PaymentInfo
            signal.Receive(ctx, &paymentInfo)
            
            // 验证支付
            if err := workflow.ExecuteActivity(ctx, VerifyPayment, paymentInfo).Get(ctx, nil); err != nil {
                state = "payment-failed"
            } else {
                state = "paid"
            }
            
        case "paid":
            logger.Info("订单已支付，准备处理")
            // 处理订单
            if err := workflow.ExecuteActivity(ctx, ProcessOrder, orderID).Get(ctx, nil); err != nil {
                state = "processing-failed"
            } else {
                state = "processing"
            }
            
        case "processing":
            logger.Info("订单处理中")
            // 模拟处理时间
            workflow.Sleep(ctx, time.Second)
            state = "ready"
            
        case "ready":
            logger.Info("订单已准备好，等待发货")
            // 等待发货信号
            signal := workflow.NewSignalChannel(ctx, "ship-order")
            var shippingInfo ShippingInfo
            signal.Receive(ctx, &shippingInfo)
            
            // 发货
            if err := workflow.ExecuteActivity(ctx, ShipOrder, shippingInfo).Get(ctx, nil); err != nil {
                state = "shipping-failed"
            } else {
                state = "shipped"
            }
            
        case "shipped":
            logger.Info("订单已发货")
            // 发送通知
            workflow.ExecuteActivity(ctx, SendDeliveryNotification, orderID).Get(ctx, nil)
            return nil
            
        case "payment-failed", "processing-failed", "shipping-failed":
            logger.Error("订单处理失败", "state", state)
            // 发送失败通知
            workflow.ExecuteActivity(ctx, SendFailureNotification, orderID).Get(ctx, nil)
            return fmt.Errorf("订单在状态 %s 时失败", state)
        }
    }
}
```

### 3.4.2 审批工作流

```go
// 审批状态机
func ApprovalWorkflow(ctx workflow.Context, request ApprovalRequest) error {
    logger := workflow.GetLogger(ctx)
    
    state := "pending"
    approvers := request.Approvers
    
    for len(approvers) > 0 {
        switch state {
        case "pending":
            // 发送给第一个审批者
            currentApprover := approvers[0]
            logger.Info("等待审批", "approver", currentApprover)
            
            // 等待审批信号
            signal := workflow.NewSignalChannel(ctx, "approval")
            var approval Approval
            signal.Receive(ctx, &approval)
            
            if approval.Approved {
                approvers = approvers[1:]
                if len(approvers) == 0 {
                    state = "approved"
                }
            } else {
                state = "rejected"
                break
            }
            
        case "approved":
            logger.Info("审批通过")
            workflow.ExecuteActivity(ctx, NotifyApprovalComplete, request).Get(ctx, nil)
            return nil
            
        case "rejected":
            logger.Info("审批被拒绝")
            workflow.ExecuteActivity(ctx, NotifyApprovalRejected, request).Get(ctx, nil)
            return fmt.Errorf("审批被拒绝")
        }
    }
    
    return nil
}
```

---

## 3.5 Saga 模式

### 3.5.1 基本 Saga

```go
// Saga 补偿工作流
func OrderSagaWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    
    // 步骤 1: 创建订单
    logger.Info("步骤 1: 创建订单")
    var orderID string
    if err := workflow.ExecuteActivity(ctx, CreateOrder, order).Get(ctx, &orderID); err != nil {
        return err
    }
    
    // 步骤 2: 预留库存
    logger.Info("步骤 2: 预留库存")
    if err := workflow.ExecuteActivity(ctx, ReserveInventory, orderID).Get(ctx, nil); err != nil {
        // 补偿：取消订单
        workflow.ExecuteActivity(ctx, CancelOrder, orderID)
        return err
    }
    
    // 步骤 3: 处理支付
    logger.Info("步骤 3: 处理支付")
    if err := workflow.ExecuteActivity(ctx, ProcessPayment, orderID).Get(ctx, nil); err != nil {
        // 补偿：释放库存
        workflow.ExecuteActivity(ctx, ReleaseInventory, orderID)
        // 补偿：取消订单
        workflow.ExecuteActivity(ctx, CancelOrder, orderID)
        return err
    }
    
    // 步骤 4: 发货
    logger.Info("步骤 4: 发货")
    if err := workflow.ExecuteActivity(ctx, ShipOrder, orderID).Get(ctx, nil); err != nil {
        // 补偿：退款
        workflow.ExecuteActivity(ctx, RefundPayment, orderID)
        // 补偿：释放库存
        workflow.ExecuteActivity(ctx, ReleaseInventory, orderID)
        // 补偿：取消订单
        workflow.ExecuteActivity(ctx, CancelOrder, orderID)
        return err
    }
    
    logger.Info("订单处理完成", "orderID", orderID)
    return nil
}
```

### 3.5.2 高级 Saga

```go
// 带重试的高级 Saga
func AdvancedSagaWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    
    // 定义步骤
    steps := []struct {
        Name        string
        Execute     interface{}
        Compensate  interface{}
    }{
        {
            Name:       "create-order",
            Execute:    CreateOrder,
            Compensate: CancelOrder,
        },
        {
            Name:       "reserve-inventory",
            Execute:    ReserveInventory,
            Compensate: ReleaseInventory,
        },
        {
            Name:       "process-payment",
            Execute:    ProcessPayment,
            Compensate: RefundPayment,
        },
        {
            Name:       "ship-order",
            Execute:    ShipOrder,
            Compensate: CancelShipment,
        },
    }
    
    // 执行步骤
    executed := make([]int, 0)
    
    for i, step := range steps {
        logger.Info("执行步骤", "step", step.Name)
        
        // 设置重试
        ao := workflow.ActivityOptions{
            RetryPolicy: &temporal.RetryPolicy{
                InitialInterval:    time.Second,
                BackoffCoefficient: 2.0,
                MaximumAttempts:    3,
            },
        }
        ctx = workflow.WithActivityOptions(ctx, ao)
        
        err := workflow.ExecuteActivity(ctx, step.Execute, order).Get(ctx, nil)
        
        if err != nil {
            logger.Error("步骤失败，开始补偿", "step", step.Name, "error", err)
            
            // 逆向补偿已执行的步骤
            for j := len(executed) - 1; j >= 0; j-- {
                stepIdx := executed[j]
                compensate := steps[stepIdx].Compensate
                logger.Info("执行补偿", "step", steps[stepIdx].Name)
                workflow.ExecuteActivity(ctx, compensate, order).Get(ctx, nil)
            }
            
            return err
        }
        
        executed = append(executed, i)
    }
    
    return nil
}
```

---

## 3.6 管道模式

### 3.6.1 线性管道

```go
// 线性管道：每个阶段的输出是下一个阶段的输入
func PipelineWorkflow(ctx workflow.Context, data Data) (Result, error) {
    logger := workflow.GetLogger(ctx)
    
    // 阶段 1: 数据验证
    logger.Info("管道阶段 1: 数据验证")
    validated, err := validateData(ctx, data)
    if err != nil {
        return nil, err
    }
    
    // 阶段 2: 数据转换
    logger.Info("管道阶段 2: 数据转换")
    transformed, err := transformData(ctx, validated)
    if err != nil {
        return nil, err
    }
    
    // 阶段 3: 数据增强
    logger.Info("管道阶段 3: 数据增强")
    enriched, err := enrichData(ctx, transformed)
    if err != nil {
        return nil, err
    }
    
    // 阶段 4: 数据保存
    logger.Info("管道阶段 4: 数据保存")
    result, err := saveData(ctx, enriched)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

func validateData(ctx workflow.Context, data Data) (Data, error) {
    var result Data
    err := workflow.ExecuteActivity(ctx, ValidateActivity, data).Get(ctx, &result)
    return result, err
}

func transformData(ctx workflow.Context, data Data) (Data, error) {
    var result Data
    err := workflow.ExecuteActivity(ctx, TransformActivity, data).Get(ctx, &result)
    return result, err
}

func enrichData(ctx workflow.Context, data Data) (Data, error) {
    var result Data
    err := workflow.ExecuteActivity(ctx, EnrichActivity, data).Get(ctx, &result)
    return result, err
}

func saveData(ctx workflow.Context, data Data) (Result, error) {
    var result Result
    err := workflow.ExecuteActivity(ctx, SaveActivity, data).Get(ctx, &result)
    return result, err
}
```

### 3.6.2 并行管道

```go
// 并行管道：多个阶段同时执行
func ParallelPipelineWorkflow(ctx workflow.Context, data Data) (Result, error) {
    logger := workflow.GetLogger(ctx)
    
    // 并行执行多个阶段
    var validated, transformed, enriched Data
    
    f1 := workflow.ExecuteActivity(ctx, ValidateActivity, data)
    f2 := workflow.ExecuteActivity(ctx, TransformActivity, data)
    f3 := workflow.ExecuteActivity(ctx, EnrichActivity, data)
    
    // 等待所有阶段完成
    if err := f1.Get(ctx, &validated); err != nil {
        return nil, err
    }
    if err := f2.Get(ctx, &transformed); err != nil {
        return nil, err
    }
    if err := f3.Get(ctx, &enriched); err != nil {
        return nil, err
    }
    
    // 合并结果
    merged := mergeData(validated, transformed, enriched)
    
    // 保存结果
    var result Result
    if err := workflow.ExecuteActivity(ctx, SaveActivity, merged).Get(ctx, &result); err != nil {
        return nil, err
    }
    
    return result, nil
}

func mergeData(d1, d2, d3 Data) Data {
    // 合并逻辑
    return Data{}
}
```

---

## 3.7 子工作流模式

### 3.7.1 同步子工作流

```go
// 同步子工作流
func ParentWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    
    logger.Info("启动子工作流")
    
    // 调用子工作流
    var result string
    childOpts := workflow.ChildWorkflowOptions{
        WorkflowID: "child-" + order.ID,
    }
    ctx = workflow.WithChildOptions(ctx, childOpts)
    
    err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, order).Get(ctx, &result)
    if err != nil {
        return err
    }
    
    logger.Info("子工作流完成", "result", result)
    return nil
}

func ChildWorkflow(ctx workflow.Context, order Order) (string, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("子工作流执行中")
    
    // 子工作流逻辑
    return "child-completed", nil
}
```

### 3.7.2 异步子工作流

```go
// 异步子工作流
func ParentAsyncWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    
    // 启动子工作流（不等待）
    childOpts := workflow.ChildWorkflowOptions{
        WorkflowID:        "child-async-" + order.ID,
        ParentClosePolicy: temporal.ParentClosePolicyRequestCancel,
    }
    ctx = workflow.WithChildOptions(ctx, childOpts)
    
    workflow.ExecuteChildWorkflow(ctx, LongRunningChildWorkflow, order)
    
    logger.Info("子工作流已启动，继续执行主工作流")
    
    // 主工作流继续执行其他任务
    workflow.ExecuteActivity(ctx, NotifyParentStarted, order).Get(ctx, nil)
    
    return nil
}
```

### 3.7.3 子工作流数组

```go
// 启动多个子工作流
func MultipleChildrenWorkflow(ctx workflow.Context, orders []Order) error {
    logger := workflow.GetLogger(ctx)
    
    // 为每个订单启动子工作流
    futures := make([]workflow.Future, len(orders))
    
    for i, order := range orders {
        childOpts := workflow.ChildWorkflowOptions{
            WorkflowID: "order-child-" + order.ID,
        }
        ctx = workflow.WithChildOptions(ctx, childOpts)
        
        futures[i] = workflow.ExecuteChildWorkflow(ctx, ProcessOrderWorkflow, order)
    }
    
    // 等待所有子工作流完成
    var errors []string
    for i, future := range futures {
        var err error
        future.Get(ctx, &err)
        if err != nil {
            errors = append(errors, fmt.Sprintf("订单 %s: %v", orders[i].ID, err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("子工作流失败: %s", strings.Join(errors, "; "))
    }
    
    logger.Info("所有子工作流完成")
    return nil
}
```

---

## 3.8 回调模式

### 3.8.1 外部回调

```go
// 等待外部回调的工作流
func CallbackWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 启动处理
    logger.Info("开始处理订单，等待外部回调")
    if err := workflow.ExecuteActivity(ctx, ProcessOrder, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 等待回调信号
    signal := workflow.NewSignalChannel(ctx, "external-callback")
    var callback CallbackData
    signal.Receive(ctx, &callback)
    
    logger.Info("收到回调", "status", callback.Status)
    
    // 根据回调结果处理
    if callback.Status == "approved" {
        return workflow.ExecuteActivity(ctx, CompleteOrder, orderID).Get(ctx, nil)
    } else {
        return workflow.ExecuteActivity(ctx, RejectOrder, orderID).Get(ctx, nil)
    }
}
```

### 3.8.2 定时回调

```go
// 定时回调工作流
func TimedCallbackWorkflow(ctx workflow.Context, task Task) error {
    logger := workflow.GetLogger(ctx)
    
    // 创建定时器
    timer := workflow.NewTimer(ctx, task.Timeout)
    
    // 等待任务完成或超时
    selector := workflow.NewSelector(ctx)
    
    taskCompleted := make(chan bool, 1)
    workflow.Go(ctx, func(ctx workflow.Context) {
        // 监听任务完成信号
        signal := workflow.NewSignalChannel(ctx, "task-completed")
        var completed bool
        signal.Receive(ctx, &completed)
        taskCompleted <- completed
    })
    
    selector.AddFuture(timer, func(f workflow.Future) {
        f.Get(ctx, nil)
        logger.Info("任务超时")
    })
    selector.AddReceive(workflow.NewReceiveChannel(ctx, taskCompleted), func(c workflow.ReceiveChannel, more bool) {
        var completed bool
        c.Receive(ctx, &completed)
        logger.Info("任务完成", "completed", completed)
    })
    
    selector.Select(ctx)
    
    return nil
}
```

---

## 3.9 错误处理模式

### 3.9.1 重试模式

```go
// 带重试的工作流
func RetryWorkflow(ctx workflow.Context, data Data) error {
    logger := workflow.GetLogger(ctx)
    
    // 设置活动选项（包含重试策略）
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    5,
            NonRetryableErrorTypes: []string{
                "InvalidInput",
                "AuthenticationFailed",
            },
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    logger.Info("开始执行活动")
    return workflow.ExecuteActivity(ctx, RiskyActivity, data).Get(ctx, nil)
}
```

### 3.9.2 降级模式

```go
// 降级处理工作流
func DegradationWorkflow(ctx workflow.Context, request Request) error {
    logger := workflow.GetLogger(ctx)
    
    // 尝试主要服务
    var primaryResult Result
    err := workflow.ExecuteActivity(ctx, PrimaryService, request).Get(ctx, &primaryResult)
    
    if err != nil {
        logger.Warn("主要服务失败，尝试降级服务", "error", err)
        
        // 降级到备用服务
        var fallbackResult Result
        err = workflow.ExecuteActivity(ctx, FallbackService, request).Get(ctx, &fallbackResult)
        if err != nil {
            logger.Error("降级服务也失败", "error", err)
            return err
        }
        
        logger.Info("降级服务成功")
    } else {
        logger.Info("主要服务成功")
    }
    
    return nil
}
```

### 3.9.3 超时处理

```go
// 超时处理工作流
func TimeoutWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 设置超时
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 执行活动
    future := workflow.ExecuteActivity(ctx, LongRunningActivity, orderID)
    
    // 等待活动完成或超时
    select {
    case <-future.GetChannel(ctx):
        var err error
        future.Get(ctx, &err)
        if err != nil {
            logger.Error("活动失败", "error", err)
            return err
        }
        logger.Info("活动完成")
        
    case <-ctx.Done():
        logger.Warn("活动超时")
        // 处理超时
        workflow.ExecuteActivity(ctx, HandleTimeout, orderID).Get(ctx, nil)
    }
    
    return nil
}
```

---

## 3.10 最佳实践

### 3.10.1 模式选择

| 场景 | 推荐模式 |
|------|----------|
| 顺序步骤 | 顺序执行 |
| 批量处理 | 扇出/扇入 |
| 复杂状态 | 状态机 |
| 跨服务事务 | Saga |
| 数据处理管道 | 管道 |
| 隔离执行 | 子工作流 |

### 3.10.2 性能优化

```go
// 优化并行执行
func OptimizedParallelWorkflow(ctx workflow.Context, items []Item) error {
    // 限制并发数
    semaphore := make(chan struct{}, 10)
    
    futures := make([]workflow.Future, len(items))
    
    for i, item := range items {
        // 获取信号量
        ch := workflow.NewChannel(ctx)
        workflow.Go(ctx, func(ctx workflow.Context) {
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            futures[i] = workflow.ExecuteActivity(ctx, ProcessItem, item)
            ch.Send(ctx, nil)
        })
        ch.Receive(ctx, nil)
    }
    
    // 等待所有完成
    for _, future := range futures {
        future.Get(ctx, nil)
    }
    
    return nil
}
```