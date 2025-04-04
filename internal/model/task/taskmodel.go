package task

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"Storage/internal/errors"

	pkgerr "github.com/pkg/errors"

	"github.com/mr-tron/base58"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TaskCollectionName = "tasks" // 集合名称常量

var _ TaskModel = (*customTaskModel)(nil)

type (
	// 核心接口扩展
	TaskModel interface {
		taskModel // 继承自动生成的基础接口
		// 新增自定义方法
		FindByType(ctx context.Context, taskType string, page, size int, sortBy string) ([]*Task, error)
		FindByCreator(ctx context.Context, creator string, page, size int, sortBy string) ([]*Task, error)
		FindEnabledTasks(ctx context.Context, enabled bool) ([]*Task, error)
		FindByTimeRange(ctx context.Context, start, end time.Time) ([]*Task, error)
		FindByScenario(ctx context.Context, scenarioID string) ([]*Task, error)
		FindByComposite(ctx context.Context, creator, taskType string, enabled bool) ([]*Task, error)
		SearchByName(ctx context.Context, keyword string) ([]*Task, error)
		FindOneByName(ctx context.Context, name string) (*Task, error)
		FindOneByTaskID(ctx context.Context, taskID string) (*Task, error)
		FindAllTask(ctx context.Context) ([]*Task, error)
		InsertTask(ctx context.Context, data *Task) error
		DeleteOneByTaskID(ctx context.Context, taskId string) (int64, error)
		UpdateTask(ctx context.Context, taskId string, data *Task) error
	}

	// 自定义模型（组合基础模型）
	customTaskModel struct {
		*defaultTaskModel
	}
)

// NewTaskModel 初始化 TaskModel
func NewTaskModel(url, db, collection string) TaskModel {

	conn := mon.MustNewModel(url, db, collection)
	// 初始化默认模型
	defaultModel := newDefaultTaskModel(conn)

	// 返回自定义模型
	return &customTaskModel{
		defaultTaskModel: defaultModel,
	}
}

// 实现扩展方法
func (m *customTaskModel) InsertTask(ctx context.Context, data *Task) error {
	if len(data.TaskName) < 3 || len(data.TaskName) > 50 {
		return errors.New(errors.ValidationFailed).WithDetails("任务名称无效", nil)
	}

	existRecord, err := m.FindOneByName(ctx, data.TaskName)
	if existRecord != nil && err == nil {
		// 有存在同名任务, 后面加个error
		return errors.New(errors.DBQueryError).WithDetails("存在同名任务", nil)
	}
	if err != nil && !errors.Is(err, errors.DBNotFound) {
		// 查询出现错误,但不是找不到记录
		return errors.NewWithError(err, errors.DBQueryError).WithDetails("查询数据库失败", err.Error())
	}

	// 生成任务ID
	// taskID, err := generateTaskID(data)
	taskID := GenerateTaskID()
	// if err != nil {
	// 	return errors.NewWithError(err, errors.GenerateTaskIDError).WithDetails("任务ID生成失败", err.Error())
	// 	// return fmt.Errorf("Generate TaskID failed: %w", err)
	// }

	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}
	data.TaskId = taskID
	data.Version = 1
	data.CreateAt = time.Now()
	data.UpdateAt = time.Now()

	_, err = m.conn.InsertOne(ctx, data)
	if mongo.IsDuplicateKeyError(err) {
		return errors.NewWithError(err, errors.DBDuplicateEntry).WithDetails("插入重复记录", err.Error())
	}
	if err != nil {
		return errors.NewWithError(err, errors.DBTxCommitError).WithDetails("插入任务记录失败", nil)
	}
	return nil
}

// func (m *customTaskModel) handleInsertConflict(ctx context.Context, data *Task) error {
// 	// 1. 检查是否真实存在冲突
// 	existing, err := m.FindOneByName(ctx, data.TaskName)
// 	if err != nil {
// 		return err
// 	}

// 	// 2. 处理不同冲突类型
// 	switch {
// 	case existing.TaskId == data.TaskId:
// 		return errors.New("任务ID已存在")
// 	case existing.TaskName == data.TaskName:
// 		return errors.New("任务名称已存在")
// 	default:
// 		return errors.New("未知冲突类型")
// 	}
// }

func (m *customTaskModel) UpdateTask(ctx context.Context, taskId string, data *Task) error {
	filter := bson.M{
		"taskId": taskId,
		// "version": data.Version - 1, // 基于旧版本更新
	}

	// if data.ID.IsZero() {
	// 	return errors.New(errors.InvalidMgoObjId)
	// }
	// if len(data.TaskName) < 3 || len(data.TaskName) > 50 {
	// 	return errors.New(errors.InvlidMgoRecordError).WithDetails("TaskName is Invalid", nil)
	// }

	// existing, err := m.FindOneByTaskID(ctx, data.TaskId)
	// if err != nil && !errors.Is(err, errors.DBNotFound) {
	// 	// 有错误，但不是找不到记录
	// 	return err.(*errors.Error).WithDetails("更新任务前查找任务失败", err.Error())
	// }

	// if existing == nil {
	// 	// 找不到记录
	// 	return errors.New(errors.DBNotFound).WithDetails("Find Task By ID not found", nil)
	// }

	// data.UpdateAt = time.Now()
	// data.Version = atomic.AddInt64(&existing.Version, 1)

	updateDoc := bson.M{
		"$set": bson.M{
			"taskName": data.TaskName,
			"taskDesc": data.TaskDesc,
			"enable":   data.Enable,
			"version":  data.Version,
			"apiSpec":  data.APISpec,
			"syncSpec": data.SyncSpec,
			"updateAt": time.Now(),
		},
	}

	res, err := m.conn.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		logx.Error(err)
		return err.(*errors.Error).WithDetails("更新任务失败", err.Error())
	}

	if res.MatchedCount == 0 {
		return errors.New(errors.UpdateMgoRecordError).WithDetails("版本冲突或任务不存在", nil)
		// current, err := m.FindOneByTaskID(ctx, data.TaskId)
		// if err != nil {
		// 	switch {
		// 	case current == nil && errors.Is(err, errors.DBNotFound):
		// 		return err.(*errors.Error).WithDetails("更新任务时任务不存在 ", nil)
		// 	case current.Version > existing.Version:
		// 		return errors.NewWithError(err, errors.InvalidMgoRecordVersionError).WithDetails("数据已被修改，请刷新后重试", nil)
		// 	default:
		// 		return errors.NewWithError(err, errors.UpdateMgoRecordError).WithDetails("更新任务失败", err.Error())
		// 	}
		// }
	}

	return nil
}

func (m *customTaskModel) DeleteTask(ctx context.Context, id string) error {
	existing, err := m.FindOneByTaskID(ctx, id)
	if err != nil && !errors.Is(err, errors.DBNotFound) {
		return err.(*errors.Error).WithDetails("删除任务前查询任务失败", err.Error())
	}

	if errors.Is(err, errors.DBNotFound) {
		return err.(*errors.Error).WithDetails("删除任务前查询任务不存在", err.Error())
		// return errors.New(errors.DBNotFound).WithDetails("Find task by id not found", nil)
	}

	oid, err := primitive.ObjectIDFromHex(existing.ID.Hex())
	if err != nil {
		return errors.New(errors.InvalidMgoObjId).WithDetails("生成objid失败", err.Error())
	}

	_, err = m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		err = errors.NewWithError(err, errors.DeleteMgoRecordError)
	}
	return err
}

func (m *customTaskModel) FindAllTask(ctx context.Context) ([]*Task, error) {
	var result []*Task
	// 使用空filter查询所有文档
	err := m.conn.Find(ctx, &result, bson.M{})

	// 错误处理与其他查询方法保持风格一致
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).
			WithDetails("find all tasks: no records found", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).
			WithDetails("find all tasks failed", err.Error())
		return nil, err
	}

	// 直接返回查询结果（可能为空数组）
	return result, nil
}

// 扩展查询实现
func (m *customTaskModel) FindOneByName(ctx context.Context, name string) (*Task, error) {
	filter := bson.M{"taskName": name}
	var task Task
	err := m.conn.FindOne(ctx, &task, filter)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return &task, nil
}

func (m *customTaskModel) FindOneByTaskID(ctx context.Context, taskID string) (*Task, error) {
	var task Task

	err := m.conn.FindOne(ctx, &task, bson.M{"taskId": taskID})

	logx.Error(taskID)

	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return &task, nil
}

func (m *customTaskModel) FindByType(ctx context.Context, taskType string, page, size int, sortBy string) ([]*Task, error) {
	filter := bson.M{"type": taskType}
	opts := options.Find().
		SetSkip(int64((page - 1) * size)).
		SetLimit(int64(size)).
		SetSort(parseSort(sortBy))

	var result []*Task
	err := m.conn.Find(ctx, &result, filter, opts)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
	}
	return result, nil
}

func (m *customTaskModel) FindByCreator(ctx context.Context, creator string, page, size int, sortBy string) ([]*Task, error) {
	filter := bson.M{"creator": creator}
	opts := options.Find().
		SetSkip(int64((page - 1) * size)).
		SetLimit(int64(size)).
		SetSort(parseSort(sortBy))

	var result []*Task
	err := m.conn.Find(ctx, &result, filter, opts)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return result, nil
}

func (m *customTaskModel) FindEnabledTasks(ctx context.Context, enabled bool) ([]*Task, error) {
	filter := bson.M{"enabled": enabled}
	var result []*Task
	err := m.conn.Find(ctx, &result, filter)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err

	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return result, nil
}

func (m *customTaskModel) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*Task, error) {
	filter := bson.M{
		"createAt": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	var result []*Task
	err := m.conn.Find(ctx, &result, filter)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return result, nil
}

func (m *customTaskModel) FindByScenario(ctx context.Context, scenarioID string) ([]*Task, error) {
	filter := bson.M{
		"scenarios": bson.M{
			"$in": []string{scenarioID},
		},
	}
	var result []*Task
	err := m.conn.Find(ctx, &result, filter)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return result, nil
}

func (m *customTaskModel) FindByComposite(ctx context.Context, creator, taskType string, enabled bool) ([]*Task, error) {
	filter := bson.M{
		"creator": creator,
		"type":    taskType,
		"enabled": enabled,
	}

	var result []*Task
	err := m.conn.Find(ctx, &result, filter)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err
	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return result, nil
}

func (m *customTaskModel) SearchByName(ctx context.Context, keyword string) ([]*Task, error) {
	filter := bson.M{
		"name": bson.M{
			"$regex":   keyword,
			"$options": "i",
		},
	}

	var result []*Task
	err := m.conn.Find(ctx, &result, filter)
	if pkgerr.Is(err, mon.ErrNotFound) {
		err = errors.NewWithError(err, errors.DBNotFound).WithDetails("find task by name, not found record", err.Error())
		return nil, err

	} else if err != nil {
		err = errors.NewWithError(err, errors.DBQueryError).WithDetails("Find task by name occur error", err.Error())
		return nil, err
	}
	return result, nil
}

// 辅助方法：解析排序参数
func parseSort(sortBy string) bson.D {
	switch sortBy {
	case "newest":
		return bson.D{{Key: "createAt", Value: -1}}
	case "oldest":
		return bson.D{{Key: "createAt", Value: 1}}
	default:
		return bson.D{{Key: "_id", Value: 1}}
	}
}

// 生成机器标识（简化版）
func getMachineID() []byte {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				return []byte{ip[2], ip[3]} // 取IP后两段（如192.168.1.1 → 0x01, 0x01）
			}
		}
	}
	return []byte{0x00, 0x00}
}

// 生成全局唯一的TaskID（格式：时间戳(4B) + 机器标识(2B) + 进程ID(2B) + 随机数(4B) + 序列号(2B)）
func GenerateTaskID() string {
	var (
		machineID []byte     // 机器标识（如IP哈希）
		pid       uint16     // 进程ID
		counter   uint32 = 0 // 自增序列号（原子操作）
	)

	// 初始化机器标识（基于本机IP的低16位）
	machineID = getMachineID()
	// 初始化进程ID（示例，实际可读取系统PID）
	pid = uint16(os.Getpid() % 0xFFFF)

	buf := make([]byte, 14)

	// 时间戳（4字节，精确到秒，支持到2038年）
	now := uint32(time.Now().Unix())
	binary.BigEndian.PutUint32(buf[0:4], now)

	// 机器标识（2字节）
	buf[4], buf[5] = machineID[0], machineID[1]

	// 进程ID（2字节）
	binary.BigEndian.PutUint16(buf[6:8], pid)

	// 随机数（4字节）
	rand.Read(buf[8:12])

	// 自增序列号（2字节，原子递增）
	c := atomic.AddUint32(&counter, 1)
	binary.BigEndian.PutUint16(buf[12:14], uint16(c%0xFFFF))

	// 转为Base64 URL安全编码（长度固定为20字符）
	return base64.URLEncoding.EncodeToString(buf)
}

func generateTaskID(task *Task) (string, error) {
	// 参数基础校验
	if task == nil {
		return "", errors.New(errors.InvalidParameter).WithDetails("generate task id, task params nil", nil)
	}

	// 1. 根据任务类型获取对应的场景列表
	var scenarios []ScenarioRef
	switch {
	case task.APISpec != nil:
		scenarios = task.APISpec.Scenarios
	default:
		return "", errors.New(errors.InvalidParameter).WithDetails("任务缺少有效的规格配置", nil)
	}

	// 2. 校验场景列表
	if len(scenarios) == 0 {
		return "", errors.New(errors.ValidationFailed).WithDetails("scenarios emppty", nil)
		// return "", errors.New("场景列表不能为空")
	}

	// 3. 提取场景ID
	sceneIDs := make([]string, 0, len(scenarios))
	for i, scenario := range scenarios {
		if scenario.ID == "" {
			return "", fmt.Errorf("场景[%d]的ID不能为空", i)
		}
		sceneIDs = append(sceneIDs, scenario.ID)
	}

	// 4. 排序生成哈希基值
	sort.Strings(sceneIDs)
	joined := strings.Join(sceneIDs, "|")

	// 5. 计算哈希值
	hasher := sha256.New()
	hasher.Write([]byte(joined))
	hash := hasher.Sum(nil)

	// 6. Base58编码（取前8字节）
	encoded := base58.Encode(hash[:8])

	// 7. 生成随机后缀（2字节）
	randSuffix := make([]byte, 2)
	if _, err := rand.Read(randSuffix); err != nil {
		return "", errors.NewWithError(err, errors.GenerateTaskIDError).WithDetails("生成随机后缀失败", err.Error())
		// return "", fmt.Errorf("生成随机后缀失败: %w", err)
	}

	// 8. 组合最终ID
	return fmt.Sprintf("TASK-%s-%02X", encoded, randSuffix), nil
}

func (m *customTaskModel) DeleteOneByTaskID(ctx context.Context, taskId string) (int64, error) {
	// 参数校验
	if taskId == "" {
		return 0, errors.New(errors.InvalidParameter).
			WithDetails("DeleteOneByTaskID: empty taskId", nil)
	}

	// 1. 查询任务是否存在
	existing, err := m.FindOneByTaskID(ctx, taskId)
	if err != nil {
		if errors.Is(err, errors.DBNotFound) {
			return 0, errors.New(errors.DBNotFound).
				WithDetails(fmt.Sprintf("taskId=%s 不存在", taskId), nil)
		}
		return 0, errors.NewWithError(err, errors.DBQueryError).
			WithDetails("删除前查询任务失败", err.Error())
	}

	// 2. 转换为 ObjectID
	oid, err := primitive.ObjectIDFromHex(existing.ID.Hex())
	if err != nil {
		return 0, errors.New(errors.InvalidMgoObjId).
			WithDetails("ObjectID 转换失败", err.Error())
	}

	// 3. 执行删除操作
	res, err := m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return 0, errors.NewWithError(err, errors.DeleteMgoRecordError).
			WithDetails("数据库删除操作失败", err.Error())
	}

	// 4. 验证删除结果
	if res == 0 {
		return 0, errors.New(errors.DBNotFound).
			WithDetails(fmt.Sprintf("taskId=%s 记录不存在", taskId), nil)
	}

	return res, nil
}
