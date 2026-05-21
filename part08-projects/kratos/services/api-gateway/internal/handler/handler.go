// Package handler 提供 HTTP 处理器，将 HTTP 请求转换为后端 gRPC 调用
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	orderV1 "services/api/order/v1"
	paymentV1 "services/api/payment/v1"
	productV1 "services/api/product/v1"
	userV1 "services/api/user/v1"
	"services/api-gateway/internal/client"
)

// Handler API 网关处理器
type Handler struct {
	clients *client.Clients
	log     *zap.Logger
}

// New 创建 Handler
func New(clients *client.Clients, log *zap.Logger) *Handler {
	return &Handler{clients: clients, log: log}
}

// Register 注册路由
func (h *Handler) Register(r *mux.Router) {
	// 健康检查
	r.HandleFunc("/health", h.health).Methods(http.MethodGet)

	// 认证
	r.HandleFunc("/api/v1/auth/login", h.login).Methods(http.MethodPost)

	// 用户
	r.HandleFunc("/api/v1/users", h.createUser).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/users/{id:[0-9]+}", h.getUser).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users", h.listUsers).Methods(http.MethodGet)

	// 商品
	r.HandleFunc("/api/v1/products", h.createProduct).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/products", h.listProducts).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/products/{id:[0-9]+}", h.getProduct).Methods(http.MethodGet)

	// 订单
	r.HandleFunc("/api/v1/orders", h.createOrder).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/orders/{id:[0-9]+}", h.getOrder).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/orders", h.listOrders).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/orders/{id:[0-9]+}/cancel", h.cancelOrder).Methods(http.MethodPost)

	// 支付
	r.HandleFunc("/api/v1/payments", h.createPayment).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/payments/{id:[0-9]+}", h.getPayment).Methods(http.MethodGet)
}

// ===== 健康检查 =====
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ===== 认证 =====
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req userV1.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.clients.User.Login(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ===== 用户 =====
func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req userV1.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.clients.User.CreateUser(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	resp, err := h.clients.User.GetUser(r.Context(), &userV1.GetUserRequest{Id: id})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.clients.User.ListUser(r.Context(), &userV1.ListUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Keyword:  r.URL.Query().Get("keyword"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ===== 商品 =====
func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req productV1.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.clients.Product.CreateProduct(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	resp, err := h.clients.Product.GetProduct(r.Context(), &productV1.GetProductRequest{Id: id})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.clients.Product.ListProduct(r.Context(), &productV1.ListProductRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Category: r.URL.Query().Get("category"),
		Keyword:  r.URL.Query().Get("keyword"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ===== 订单 =====
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req orderV1.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.clients.Order.CreateOrder(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	resp, err := h.clients.Order.GetOrder(r.Context(), &orderV1.GetOrderRequest{Id: id})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	resp, err := h.clients.Order.ListOrder(r.Context(), &orderV1.ListOrderRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		UserId:   userID,
		Status:   r.URL.Query().Get("status"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	resp, err := h.clients.Order.CancelOrder(r.Context(), &orderV1.CancelOrderRequest{
		Id:     id,
		Reason: body.Reason,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ===== 支付 =====
func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req paymentV1.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.clients.Payment.CreatePayment(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	resp, err := h.clients.Payment.GetPayment(r.Context(), &paymentV1.GetPaymentRequest{Id: id})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ===== 工具方法 =====
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
