package dao

import (
	"errors"
	"strings"
	"time"
	"yuko_chat/internal/config"
	"yuko_chat/pkg/zlog"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// kafkaClient 用来集中管理当前项目里对 Kafka 的访问对象。
// 这里把 Producer、Consumer、Admin 放在一起，
// 是为了让初始化、关闭、发送消息、创建 topic 这些动作都围绕同一个全局入口展开，
// 用起来会比较像你前面的 RedisClient。
type kafkaClient struct {
	// Producer: 同步生产者。
	// 调用 SendMessage 后会等待 Kafka 返回结果，
	// 成功时能拿到消息最终落到哪个 partition、对应 offset 是多少。
	Producer sarama.SyncProducer

	// Consumer: 普通消费者。
	// 这里使用的是 sarama 的“简单消费者”能力，
	// 它适合先学习 Kafka 基本模型：topic / partition / offset。
	// 注意它不是 Consumer Group 模式。
	Consumer sarama.Consumer

	// Admin: 管理端客户端。
	// 一般用于创建 topic、查看 topic、删除 topic 等管理型操作。
	Admin sarama.ClusterAdmin

	// Brokers: Kafka broker 地址列表。
	// 例如配置里如果写 "localhost:9092,localhost:9093"，
	// 最终会被拆成一个字符串切片存到这里。
	Brokers []string

	// Topic: 当前项目默认使用的 topic。
	// 为了让发送消息时不用每次手动传 topic，
	// 我们在初始化时把配置里的默认 topic 缓存下来。
	Topic string
}

// KafkaClient 是 dao 层统一暴露出去的 Kafka 访问入口。
// 后续业务层一般只需要 dao.InitKafka() + dao.KafkaClient.xxx 或包装函数即可。
var KafkaClient = new(kafkaClient)

// InitKafka 初始化 Kafka 的三个核心能力：
// 1. Producer: 发消息
// 2. Consumer: 读消息
// 3. Admin: 管理 topic
func InitKafka() error {
	// 先把配置取出来，后面就不用一直写很长的 config.Cfg.KafkaConfig。
	kafkaCfg := config.Cfg.KafkaConfig

	// 把配置文件里的 hostPort 拆成 broker 列表。
	// 这样既兼容单节点 "localhost:9092"，
	// 也兼容多节点 "host1:9092,host2:9092"。
	brokers := parseBrokers(kafkaCfg.HostPort)
	if len(brokers) == 0 {
		err := errors.New("kafka brokers are empty")
		zlog.Error("failed to init kafka", zap.Error(err))
		return err
	}
	if kafkaCfg.Topic == "" {
		err := errors.New("kafka topic is empty")
		zlog.Error("failed to init kafka", zap.Error(err))
		return err
	}

	// sarama.NewConfig() 是 Sarama 的核心配置对象。
	// Producer / Consumer / 网络超时等参数都在这里调整。
	saramaCfg := sarama.NewConfig()

	// Producer.RequiredAcks 表示生产者发送消息后，需要等多少确认才算成功。
	// WaitForAll 是最稳妥的一种：
	// 只有 leader 和所有 ISR 副本都确认后，才认为发送成功。
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll

	// SyncProducer 必须把 Return.Successes 设为 true，
	// 否则 SendMessage 时拿不到成功结果，Sarama 也会报配置错误。
	saramaCfg.Producer.Return.Successes = true

	// Return.Errors 设为 true 后，Producer 内部会把错误返回出来，
	// 这样我们在 SendMessage 时能拿到详细错误。
	saramaCfg.Producer.Return.Errors = true

	// NewManualPartitioner returns a Partitioner
	// which uses the partition manually set in the provided ProducerMessage's
	// Partition field as the partition to produce to.
	saramaCfg.Producer.Partitioner = sarama.NewManualPartitioner

	// Producer.Timeout 是生产者等待 Kafka 确认的超时时间。
	// 你的 config.toml 里 timeout 配的是整数，这里按“秒”解释。
	saramaCfg.Producer.Timeout = kafkaCfg.Timeout * time.Second

	// 下面这三个是网络层超时：
	// DialTimeout: 建立 TCP 连接超时
	// ReadTimeout: 读超时
	// WriteTimeout: 写超时
	saramaCfg.Net.DialTimeout = kafkaCfg.Timeout * time.Second
	saramaCfg.Net.ReadTimeout = kafkaCfg.Timeout * time.Second
	saramaCfg.Net.WriteTimeout = kafkaCfg.Timeout * time.Second

	// 打开 Consumer 错误返回，便于后续消费阶段排查问题。
	saramaCfg.Consumer.Return.Errors = true

	// 创建同步生产者。
	// 如果初始化 Producer 都失败了，后面的 Consumer / Admin 就没必要继续做了。
	producer, err := sarama.NewSyncProducer(brokers, saramaCfg)
	if err != nil {
		zlog.Error("failed to create kafka producer", zap.Strings("brokers", brokers), zap.Error(err))
		return err
	}

	// 创建普通消费者。
	// 这里不是 Consumer Group，所以后面消费时需要你明确指定 partition 和 offset。
	consumer, err := sarama.NewConsumer(brokers, saramaCfg)
	if err != nil {
		// 前面已经成功创建了 producer，
		// 所以这里如果 consumer 创建失败，要记得把 producer 关掉，避免资源泄漏。
		_ = producer.Close()
		zlog.Error("failed to create kafka consumer", zap.Strings("brokers", brokers), zap.Error(err))
		return err
	}

	// 创建 Admin 客户端，用于 topic 管理。
	// 如果这里失败，同样需要把前面已经创建成功的资源关闭掉。
	admin, err := sarama.NewClusterAdmin(brokers, saramaCfg)
	if err != nil {
		_ = consumer.Close()
		_ = producer.Close()
		zlog.Error("failed to create kafka admin", zap.Strings("brokers", brokers), zap.Error(err))
		return err
	}

	// 只有三个对象都创建成功，才整体写入全局 KafkaClient。
	// 这样可以避免“初始化到一半”却留下部分可用、部分不可用状态。
	KafkaClient.Producer = producer
	KafkaClient.Consumer = consumer
	KafkaClient.Admin = admin
	KafkaClient.Brokers = brokers
	KafkaClient.Topic = kafkaCfg.Topic

	// 自动创建默认 topic（单分区，单副本）
	// 这样就不需要手动调用 CreateTopic 了
	err = CreateTopic(1, 1)
	if err != nil {
		// topic 创建失败时，需要清理已创建的资源
		_ = admin.Close()
		_ = consumer.Close()
		_ = producer.Close()
		zlog.Error("failed to create default kafka topic", zap.Error(err))
		return err
	}

	return nil
}

// CloseKafka 统一关闭 Kafka 相关资源。
// 关闭顺序上先关 Admin，再关 Consumer，最后关 Producer。
// 这里把多个关闭错误用 errors.Join 合并起来返回，
// 这样不会因为前一个 Close 失败，就丢掉后一个 Close 的报错信息。
func CloseKafka() error {
	var closeErr error

	if KafkaClient.Admin != nil {
		if err := KafkaClient.Admin.Close(); err != nil {
			zlog.Error("failed to close kafka admin", zap.Error(err))
			closeErr = errors.Join(closeErr, err)
		}
		KafkaClient.Admin = nil
	}

	if KafkaClient.Consumer != nil {
		if err := KafkaClient.Consumer.Close(); err != nil {
			zlog.Error("failed to close kafka consumer", zap.Error(err))
			closeErr = errors.Join(closeErr, err)
		}
		KafkaClient.Consumer = nil
	}

	if KafkaClient.Producer != nil {
		if err := KafkaClient.Producer.Close(); err != nil {
			zlog.Error("failed to close kafka producer", zap.Error(err))
			closeErr = errors.Join(closeErr, err)
		}
		KafkaClient.Producer = nil
	}

	KafkaClient.Brokers = nil
	KafkaClient.Topic = ""
	return closeErr
}

// CreateKafkaTopic 创建当前默认 topic。
//
// 这里把 topic 名固定用 KafkaClient.Topic，
// 因为这个值来自配置文件，比较符合你现在项目“单默认 topic”的设计。
//
// numPartitions: 要创建几个分区
// replicationFactor: 副本因子
func CreateTopic(numPartitions int32, replicationFactor int16) error {
	// CreateTopic 属于管理操作，所以必须依赖 Admin 客户端。
	if KafkaClient.Admin == nil {
		err := errors.New("kafka admin is not initialized")
		zlog.Error("failed to create kafka topic", zap.Error(err))
		return err
	}
	if numPartitions <= 0 {
		err := errors.New("numPartitions must be greater than 0")
		zlog.Error("failed to create kafka topic", zap.Error(err))
		return err
	}
	if replicationFactor <= 0 {
		err := errors.New("replicationFactor must be greater than 0")
		zlog.Error("failed to create kafka topic", zap.Error(err))
		return err
	}

	// 先列出已有 topic，目的是避免重复创建时直接报错。
	// 如果 topic 已经存在，这里直接返回 nil，把这个操作做成“幂等”的。
	topics, err := KafkaClient.Admin.ListTopics()
	if err != nil {
		zlog.Error("failed to list kafka topics", zap.Error(err))
		return err
	}
	if _, exists := topics[KafkaClient.Topic]; exists {
		return nil
	}

	// TopicDetail 里最关键的就是分区数和副本数。
	// 这里没有额外配置保留策略、压缩策略等高级参数，
	// 先保持简单，后面你需要时再往这里扩。
	err = KafkaClient.Admin.CreateTopic(KafkaClient.Topic, &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}, false)
	if err != nil {
		zlog.Error(
			"failed to create kafka topic",
			zap.String("topic", KafkaClient.Topic),
			zap.Int32("numPartitions", numPartitions),
			zap.Int16("replicationFactor", replicationFactor),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// SendMessage 往默认 topic 发送一条消息。
//
// key 的作用：可以作为业务上的消息键
//
// value 就是消息体内容。
//
// 返回值：
// 1. partition: 消息最终写入的分区
// 2. offset: 消息在该分区内的偏移量
func SendMessage(key string, value []byte) (int32, int64, error) {
	if KafkaClient.Producer == nil {
		err := errors.New("kafka producer is not initialized")
		zlog.Error("failed to send kafka message", zap.Error(err))
		return 0, 0, err
	}

	// ProducerMessage 是 Sarama 对“待发送消息”的封装。
	// 这里默认发送到 KafkaClient.Topic。
	message := &sarama.ProducerMessage{
		Topic:     KafkaClient.Topic,
		Partition: int32(config.Cfg.KafkaConfig.Partition),
		Value:     sarama.ByteEncoder(value),
	}

	// key 不是必填项。
	// 如果业务上不需要 key，可以直接传空字符串。
	if key != "" {
		message.Key = sarama.StringEncoder(key)
	}

	// SendMessage 是同步发送：
	// 调用方会阻塞等待 Kafka 返回结果。
	partition, offset, err := KafkaClient.Producer.SendMessage(message)
	if err != nil {
		zlog.Error(
			"failed to send kafka message",
			zap.String("topic", KafkaClient.Topic),
			zap.String("key", key),
			zap.Error(err),
		)
		return 0, 0, err
	}
	return partition, offset, nil
}

// NewPartitionConsumer 按配置文件里的 partition 创建消费者。
//
// 这里之所以只传 offset，
// 是因为 partition 默认从 config.toml 里读取，
// 这样你在大多数情况下只需要关心“从哪里开始消费”。
//
// offset 常见取值：
// sarama.OffsetNewest: 从最新消息开始
// sarama.OffsetOldest: 从最老消息开始
func NewPartitionConsumer(offset int64) (sarama.PartitionConsumer, error) {
	return NewPartitionConsumerByPartition(int32(config.Cfg.KafkaConfig.Partition), offset)
}

// NewPartitionConsumerByPartition 按指定分区创建消费者。
// 当你不想用配置文件里的默认 partition 时，可以直接调用这个函数。
func NewPartitionConsumerByPartition(partition int32, offset int64) (sarama.PartitionConsumer, error) {
	if KafkaClient.Consumer == nil {
		err := errors.New("kafka consumer is not initialized")
		zlog.Error("failed to create kafka partition consumer", zap.Error(err))
		return nil, err
	}

	// ConsumePartition 会返回一个 PartitionConsumer。
	// 你后面可以通过它的 Messages() channel 持续读取消息：
	//
	// pc, _ := dao.NewKafkaPartitionConsumer(sarama.OffsetNewest)
	// for msg := range pc.Messages() {
	//     fmt.Println(string(msg.Value))
	// }
	partitionConsumer, err := KafkaClient.Consumer.ConsumePartition(KafkaClient.Topic, partition, offset)
	if err != nil {
		zlog.Error(
			"failed to create kafka partition consumer",
			zap.String("topic", KafkaClient.Topic),
			zap.Int32("partition", partition),
			zap.Int64("offset", offset),
			zap.Error(err),
		)
		return nil, err
	}
	return partitionConsumer, nil
}

// parseKafkaBrokers 把配置中的 broker 字符串拆成切片。
//
// 例如：
// "localhost:9092" -> []string{"localhost:9092"}
// "host1:9092, host2:9092" -> []string{"host1:9092", "host2:9092"}
//
// TrimSpace 是为了兼容你在配置里写逗号后带空格的情况。
func parseBrokers(hostPort string) []string {
	parts := strings.Split(hostPort, ",")
	brokers := make([]string, 0, len(parts))
	for _, broker := range parts {
		broker = strings.TrimSpace(broker)
		if broker != "" {
			brokers = append(brokers, broker)
		}
	}
	return brokers
}
