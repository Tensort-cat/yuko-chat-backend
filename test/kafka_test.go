package test

import (
	"encoding/json"
	"testing"
	"yuko_chat/internal/config"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"

	"github.com/IBM/sarama"
)

// 注意：这些测试需要 Kafka 真实运行在 localhost:9092
// 如果你的 Kafka 在其他地址，请修改 config 或这里的 setup 函数

// ============ 测试辅助函数 ============

// setupKafka 初始化 Kafka 客户端
func setupKafka(t *testing.T) {
	// 注意：这里假设你的配置文件已经正确设置
	// 如果没有配置文件，可能需要先初始化 config.Cfg
	config.InitConfig()
	err := dao.InitKafka()
	if err != nil {
		t.Fatalf("failed to init kafka: %v", err)
	}
	t.Logf("kafka client initialized successfully")
}

// teardownKafka 关闭 Kafka 客户端
func teardownKafka(t *testing.T) {
	err := dao.CloseKafka()
	if err != nil {
		t.Logf("warning: failed to close kafka: %v", err)
	}
}

func TestGetMsg(t *testing.T) {
	setupKafka(t)
	defer teardownKafka(t)

	req := request.ChatMessageRequest{
		SessionId: "ssss",
		Content:   "dwqwdada",
		Type:      2,
	}
	// 序列化结构体
	jsonMsg, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	dao.SendMessage("0", jsonMsg)

	consumer, _ := dao.NewPartitionConsumer(sarama.OffsetOldest)
	defer consumer.Close()
	for kafkaMsg := range consumer.Messages() {
		// 反序列化
		var res request.ChatMessageRequest
		if err := json.Unmarshal(kafkaMsg.Value, &res); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		t.Log(res)
	}
}

// 从最老消息开始消费，直到消费完所有消息
func TestConsumeAllMessages(t *testing.T) {
	setupKafka(t)
	defer teardownKafka(t)
	consumer, err := dao.NewPartitionConsumer(sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()
	for kafkaMsg := range consumer.Messages() {
		var res request.ChatMessageRequest
		if err := json.Unmarshal(kafkaMsg.Value, &res); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		t.Log(res)
	}
}

// 发消息
func TestSendMessage(t *testing.T) {
	setupKafka(t)
	defer teardownKafka(t)
	req := request.ChatMessageRequest{
		SessionId: "ssss",
		Content:   "dwqwdada",
		Type:      2,
	}
	jsonMsg, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	partition, offset, err := dao.SendMessage("0", jsonMsg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
	t.Logf("message sent to partition %d at offset %d", partition, offset)
}

func TestCleanKafka(t *testing.T) {
	setupKafka(t)
	defer teardownKafka(t)

	// 删除 topic
	admin, err := sarama.NewClusterAdmin([]string{"localhost:9092"}, nil)
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}
	defer admin.Close()

	// 需要删除的 topic
	topics := []string{"yuko_chat_messages"}

	err = admin.DeleteTopic(topics[0])
	if err != nil {
		t.Fatalf("failed to delete topic: %v", err)
	}

	// 重新创建 topic
	err = admin.CreateTopic(topics[0], &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	t.Log("kafka topic cleaned successfully")
}

// ============ 使用示例（非测试）============

/*
本文件展示的 kafkaClient 的主要使用模式：

1. 初始化（main.go 中）：
   err := dao.InitKafka()
   if err != nil {
       log.Fatal(err)
   }
   defer dao.CloseKafka()

2. 发送消息：
   partition, offset, err := dao.SendMessage("user_123", "hello world")

3. 消费消息（单分区）：
   pc, _ := dao.NewPartitionConsumer(sarama.OffsetNewest)
   defer pc.Close()
   for msg := range pc.Messages() {
       fmt.Printf("key: %s, value: %s\n", msg.Key, msg.Value)
   }

4. 消费消息（指定分区）：
   pc, _ := dao.NewPartitionConsumerByPartition(1, sarama.OffsetOldest)
   defer pc.Close()
   for msg := range pc.Messages() {
       // 处理消息...
   }

5. 创建 topic：
   err := dao.CreateTopic(3, 1)

6. 列出所有 topic：
   topics, _ := dao.KafkaClient.Admin.ListTopics()
   for name := range topics {
       fmt.Println(name)
   }

注意事项：
- 生产者使用同步方式，会阻塞等待响应
- 消费者不使用 Consumer Group，需要手动管理分区和 offset
- 所有错误都应该被正确处理
- 程序退出前必须调用 CloseKafka() 来释放资源
*/
