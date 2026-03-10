package data

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"gorm.io/gorm"

	"kratos/internal/config"
	"kratos/internal/database"
	"kratos/internal/nats"
	"kratos/internal/redis"
)

// Data 数据层
type Data struct {
	DB       *database.DB
	NATS     *nats.Client
	Redis    *redis.Client
	Logger   *log.Helper
	jwtSecret string
}

// NewData 创建数据层
func NewData(
	db *database.DB,
	natsClient *nats.Client,
	redisClient *redis.Client,
	logger log.Logger,
	cfg *config.Config,
) (*Data, func(), error) {
	l := log.NewHelper(log.With(logger, "module", "data"))

	l.Info("初始化数据层")

	cleanup := func() {
		l.Info("关闭数据层连接")
		if natsClient != nil {
			natsClient.Close()
		}
		if redisClient != nil {
			redisClient.Close()
		}
		if db != nil {
			db.Close()
		}
	}

	return &Data{
		DB:       db,
		NATS:     natsClient,
		Redis:    redisClient,
		Logger:   l,
		jwtSecret: cfg.JWT.Secret,
	}, cleanup, nil
}

// DBClient 数据库客户端
type DBClient struct {
	*gorm.DB
}

// NewDBClient 创建数据库客户端
func NewDBClient(cfg *config.DatabaseConfig, logger log.Logger) (*DBClient, error) {
	db, err := database.NewDB(&database.Config{
		Driver:  cfg.Driver,
		Source:  cfg.Source,
		MaxOpen: cfg.MaxOpen,
		MaxIdle: cfg.MaxIdle,
		MaxLife: cfg.MaxLife,
	}, logger)
	if err != nil {
		return nil, err
	}
	return &DBClient{db.DB}, nil
}

// TokenClaims Token 声明
type TokenClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 Token
func GenerateToken(userID int64, username string) string {
	claims := TokenClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kratos-ecommerce",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("kratos-secret-key"))
	return tokenString
}

// ValidateToken 验证 Token
func ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("kratos-secret-key"), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// JSON 序列化
func JSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// JSONUnmarshal 反序列化
func JSONUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ProviderSet 数据层依赖
var ProviderSet = wire.NewSet(
	NewData,
	NewDBClient,
)

// InitDB 初始化数据库
func InitDB(cfg *config.Config, logger log.Logger) (*database.DB, error) {
	return database.NewDB(&database.Config{
		Driver:  cfg.Database.Driver,
		Source:  cfg.Database.Source,
		MaxOpen: cfg.Database.MaxOpen,
		MaxIdle: cfg.Database.MaxIdle,
		MaxLife: cfg.Database.MaxLife,
	}, logger)
}

// InitNATS 初始化 NATS
func InitNATS(cfg *config.Config, logger log.Logger) (*nats.Client, error) {
	return nats.NewClient(cfg.NATS.URL, logger)
}

// InitRedis 初始化 Redis
func InitRedis(cfg *config.Config, logger log.Logger) (*redis.Client, error) {
	return redis.NewClient(&redis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, logger)
}

// EnsureTables 确保表存在
func EnsureTables(db *database.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}