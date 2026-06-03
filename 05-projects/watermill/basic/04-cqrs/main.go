// Package main 演示 Watermill CQRS 组件：CommandBus 与 EventBus
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// CreateOrder 是命令
type CreateOrder struct {
	OrderID string
	Amount  float64
}

// OrderCreated 是事件
type OrderCreated struct {
	OrderID string
	Amount  float64
}

// PaymentCompleted 是事件
type PaymentCompleted struct {
	OrderID string
}

func main() {
	logger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		log.Fatal(err)
	}

	// 创建 CQRS facade（marshaler 使用 JSON）
	marshaler := cqrs.JSONMarshaler{}
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			return "commands." + commandName
		},
		GenerateEventsTopic: func(eventName string) string {
			return "events." + eventName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				cqrs.NewCommandHandler(
					"create_order",
					func(ctx context.Context, cmd *CreateOrder) error {
						fmt.Printf("[Command] 创建订单: %s 金额: %.2f\n", cmd.OrderID, cmd.Amount)
						return eb.Publish(ctx, &OrderCreated{
							OrderID: cmd.OrderID,
							Amount:  cmd.Amount,
						})
					},
				),
			}
		},
		EventHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				cqrs.NewEventHandler(
					"order_created_handler",
					func(ctx context.Context, event *OrderCreated) error {
						fmt.Printf("[Event] 订单已创建: %s 金额: %.2f\n", event.OrderID, event.Amount)
						return eb.Publish(ctx, &PaymentCompleted{
							OrderID: event.OrderID,
						})
					},
				),
				cqrs.NewEventHandler(
					"payment_completed_handler",
					func(ctx context.Context, event *PaymentCompleted) error {
						fmt.Printf("[Event] 支付已完成: %s\n", event.OrderID)
						return nil
					},
				),
			}
		},
		CommandsPublisher:               pubSub,
		CommandsSubscriberConstructor:   func(handlerName string) (message.Subscriber, error) { return pubSub, nil },
		EventsPublisher:                 pubSub,
		EventsSubscriberConstructor:     func(handlerName string) (message.Subscriber, error) { return pubSub, nil },
		Router:                          router,
		Logger:                          logger,
		CommandEventMarshaler:           marshaler,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := router.Run(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	<-router.Running()

	fmt.Println("发送 CreateOrder 命令...")
	if err := cqrsFacade.CommandBus().Send(ctx, &CreateOrder{
		OrderID: "ORD-001",
		Amount:  99.99,
	}); err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\n04-cqrs 演示完成")
}
