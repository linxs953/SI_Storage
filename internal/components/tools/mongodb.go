package tools

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient represents a MongoDB client wrapper
type MongoConfig struct {
	MongoHost   string
	MongoPort   int
	MongoUser   string
	MongoPasswd string
	UseDb       string
}

type MongoClient struct {
	client *mongo.Client
	config MongoConfig
	mutex  sync.Mutex
}

// NewMongoClient creates a new MongoDB client instance
func NewMongoClient(conf MongoConfig) *MongoClient {
	return &MongoClient{
		config: conf,
	}
}

// Connect establishes a connection to MongoDB
func (m *MongoClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.client != nil {
		return nil
	}
	return m.connect(ctx)
}

func (m *MongoClient) connect(ctx context.Context) error {
	// 记录连接详情
	logx.Infof("开始连接 MongoDB: %s:%d, 用户名: %s, 数据库: %s",
		m.config.MongoHost, m.config.MongoPort, m.config.MongoUser, m.config.UseDb)

	// 构建 MongoDB URI
	// uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
	// 	m.config.MongoUser,
	// 	m.config.MongoPasswd,
	// 	m.config.MongoHost,
	// 	m.config.MongoPort,
	// )
	uri := "mongodb://root:8767gbp7@dbconn.sealosbja.site:41038/?authSource=admin&directConnection=true"
	// 记录 URI
	logx.Infof("连接 URI: %s", strings.Replace(uri, m.config.MongoPasswd, "****", 1))

	// 配置 MongoDB 客户端选项
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second).
		SetSocketTimeout(10 * time.Second).
		SetMaxConnecting(2)

	// 创建客户端
	logx.Info("正在创建 MongoDB 客户端...")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logx.Errorf("创建 MongoDB 客户端失败: %v", err)
		return errors.Wrap(err, "创建 MongoDB 客户端失败")
	}

	// 测试连接
	if err = client.Ping(ctx, nil); err != nil {
		logx.Errorf("Ping MongoDB 失败: %v", err)
		return errors.Wrap(err, "无法连接到 MongoDB")
	}

	m.client = client
	logx.Infof("成功连接到 MongoDB: %s:%d", m.config.MongoHost, m.config.MongoPort)
	return nil
}

// Disconnect closes the MongoDB connection
func (m *MongoClient) Disconnect(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.client != nil {
		m.client = nil
		logx.Info("Successfully disconnected from MongoDB")
	}
	return nil
}

// GetDatabase returns the MongoDB database instance

// GetCollection returns a MongoDB collection
func (m *MongoClient) GetCollection(name string) *mongo.Collection {
	if m.client == nil {
		return nil
	}
	return m.client.Database(m.config.UseDb).Collection(name)
}

// IsConnected checks if the client is connected to MongoDB
func (m *MongoClient) IsConnected() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查客户端状态
	logx.Infof("检查连接状态: 当前会话数=%d", m.client.NumberSessionsInProgress())

	err := m.client.Ping(ctx, nil)
	if err != nil {
		logx.Errorf("连接检查失败: %+v", err)
		return false
	}

	return true
}
