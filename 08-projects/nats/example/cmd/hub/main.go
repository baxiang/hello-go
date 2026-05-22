// Package main Hub 服务主程序
// 对应项目中的 livis-claw-hub 服务
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nats-mvp-example/internal/config"
	"nats-mvp-example/internal/hub"
)

func main() {
	// 解析命令行参数
	natsURL := flag.String("nats", config.DefaultNATSURL, "NATS 服务器地址")
	httpAddr := flag.String("http", ":8080", "HTTP 服务地址")
	flag.Parse()

	log.Printf("启动 Hub 服务: nats=%s, http=%s", *natsURL, *httpAddr)

	// 创建 Hub 配置
	cfg := &config.HubConfig{
		NATSURL:          *natsURL,
		ExecuteTimeout:   60e9, // 60s
		HeartbeatTimeout: 90e9, // 90s
	}

	// 创建 Hub 实例
	h, err := hub.New(cfg)
	if err != nil {
		log.Fatalf("创建 Hub 失败: %v", err)
	}

	// 启动 Hub
	if err := h.Start(); err != nil {
		log.Fatalf("启动 Hub 失败: %v", err)
	}

	// 创建 HTTP 服务
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/health/readiness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	// 设备状态查询
	mux.HandleFunc("/api/v1/device/status", func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID == "" {
			http.Error(w, "device_id is required", http.StatusBadRequest)
			return
		}

		status, err := h.GetDeviceStatus(r.Context(), deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"device_id":"` + status.DeviceID + `","status":"` + status.Status + `"}`))
	})

	// 设备列表
	mux.HandleFunc("/api/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		devices, err := h.ListOnlineDevices(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"devices":[`))
		for i, d := range devices {
			if i > 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte(`{"device_id":"` + d.DeviceID + `","status":"` + d.Status + `"}`))
		}
		w.Write([]byte(`]}`))
	})

	// 执行命令（SSE）
	mux.HandleFunc("/api/v1/device/execute", func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Query().Get("device_id")
		if deviceID == "" {
			http.Error(w, "device_id is required", http.StatusBadRequest)
			return
		}

		// 解析请求
		var req struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// 检查设备在线
		if !h.IsDeviceOnline(r.Context(), deviceID) {
			http.Error(w, "device not online", http.StatusNotFound)
			return
		}

		// 设置 SSE 响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// 执行命令
		respCh, err := h.ExecuteCommand(r.Context(), deviceID, req.Query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 流式响应
		for resp := range respCh {
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

			if resp.Type == "response_done" || resp.Type == "error" {
				break
			}
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	// 启动 HTTP 服务
	server := &http.Server{
		Addr:    *httpAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP 服务启动: %s", *httpAddr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务错误: %v", err)
		}
	}()

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("收到退出信号，正在关闭...")

	// 关闭 HTTP 服务
	server.Shutdown(context.Background())

	// 关闭 Hub
	h.Stop()

	log.Println("Hub 服务已停止")
}