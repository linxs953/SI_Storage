package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// RedisConfig 定义 Redis 连接配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// RedisClient Redis 客户端包装器
type RedisClient struct {
	client *redis.Client
	config RedisConfig
	mutex  sync.Mutex
}

// NewRedisClient 创建新的 Redis 客户端实例
func NewRedisClient(conf RedisConfig) *RedisClient {
	return &RedisClient{
		config: conf,
	}
}

// Connect 建立 Redis 连接
func (r *RedisClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.client != nil {
		return nil
	}
	return r.connect(ctx)
}

func (r *RedisClient) connect(ctx context.Context) error {
	// 记录连接详情
	logx.Infof("开始连接 Redis: %s:%d, 数据库: %d",
		r.config.Host, r.config.Port, r.config.DB)

	// 构建 Redis 连接选项
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password:     r.config.Password,
		DB:           r.config.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	}

	// 创建 Redis 客户端
	logx.Info("正在创建 Redis 客户端...")
	client := redis.NewClient(options)

	// 测试连接
	if _, err := client.Ping(ctx).Result(); err != nil {
		logx.Errorf("Redis 连接测试失败: %v", err)
		return errors.Wrap(err, "无法连接到 Redis")
	}

	r.client = client
	logx.Infof("成功连接到 Redis: %s:%d", r.config.Host, r.config.Port)
	return nil
}

// Disconnect 关闭 Redis 连接
func (r *RedisClient) Disconnect(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.client != nil {
		if err := r.client.Close(); err != nil {
			logx.Errorf("关闭 Redis 连接时出错: %v", err)
			return err
		}
		r.client = nil
		logx.Info("成功关闭 Redis 连接")
	}
	return nil
}

// GetClient 返回 Redis 客户端实例
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// IsConnected 检查 Redis 是否已连接
func (r *RedisClient) IsConnected() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.client.Ping(ctx).Result()
	return err == nil
}

// Set 设置键值对
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.Get(ctx, key).Result()
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if r.client == nil {
		return fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.Del(ctx, key).Err()
}

// 字符串操作

// SetNX 仅在键不存在时设置值
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Increment 自增整数值
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.Incr(ctx, key).Result()
}

// IncrementBy 按指定值自增
func (r *RedisClient) IncrementBy(ctx context.Context, key string, increment int64) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.IncrBy(ctx, key, increment).Result()
}

// 哈希操作

// HSet 设置哈希表中的字段值
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.HSet(ctx, key, field, value).Result()
}

// HGet 获取哈希表中指定字段的值
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希表中的所有字段和值
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希表中的一个或多个字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.HDel(ctx, key, fields...).Result()
}

// HMSet 批量设置哈希表中的多个字段值
func (r *RedisClient) HMSet(ctx context.Context, key string, fields map[string]interface{}) error {
	if r.client == nil {
		return fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.HMSet(ctx, key, fields).Err()
}

// 列表操作

// LPush 从列表左侧插入一个或多个元素
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	if r.client == nil {
		return fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 从列表右侧插入元素
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.RPush(ctx, key, values...).Result()
}

// LPop 从列表左侧弹出元素
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.RPop(ctx, key).Result()
}

// 集合操作

// SAdd 向集合添加一个或多个成员
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if r.client == nil {
		return fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合中的所有成员
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.SMembers(ctx, key).Result()
}

// SRem 从集合中移除一个或多个成员
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.SRem(ctx, key, members...).Result()
}

// 有序集合操作

// ZAdd 向有序集合添加一个或多个成员
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	if r.client == nil {
		return fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合中指定范围内的成员
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 从有序集合中移除一个或多个成员
func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("Redis 客户端未连接")
	}
	return r.client.ZRem(ctx, key, members...).Result()
}

// Close 关闭 Redis 连接
func (r *RedisClient) Close() error {
	return r.Disconnect(context.Background())
}
