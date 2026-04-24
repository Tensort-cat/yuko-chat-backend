package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/dto/respond"
	"yuko_chat/internal/model"
	"yuko_chat/pkg/constant"
	message_enum "yuko_chat/pkg/enum/message"
	"yuko_chat/pkg/util"
	"yuko_chat/pkg/zlog"

	"github.com/IBM/sarama"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	Clients map[string]*Client
	mu      *sync.Mutex
	Login   chan *Client // 登录通道
	Logout  chan *Client // 退出登录通道
}

var Myserver *Server

func init() {
	if Myserver == nil {
		Myserver = &Server{
			Clients: make(map[string]*Client),
			Login:   make(chan *Client),
			Logout:  make(chan *Client),
			mu:      new(sync.Mutex),
		}
	}
}

func (s *Server) Close() {
	close(s.Login)
	close(s.Logout)
}

func (s *Server) SendClientToLogin(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Login <- client
}

func (s *Server) SendClientToLogout(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Logout <- client
}

func (s *Server) RemoveClient(uuid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Clients, uuid)
}

func (s *Server) Start() {
	defer func() {
		if r := recover(); r != nil {
			zlog.Error(fmt.Sprintf("Server panic: %v", r))
		}
		close(s.Login)
		close(s.Logout)
	}()

	// 开一个 goroutine 处理消息转发
	go func() {
		// 创建消费者
		consumer, err := dao.NewPartitionConsumer(sarama.OffsetNewest)
		if err != nil {
			zlog.Error("Kafka 消费者创建失败", zap.Error(err))
			return
		}
		defer func() {
			consumer.Close()
			zlog.Info("Kafka 消费者已关闭")
		}()

		// 阻塞式消费消息
		for kafkaMsg := range consumer.Messages() {
			defer func() {
				if r := recover(); r != nil {
					zlog.Error(fmt.Sprintf("处理消息时发生panic: %v", r))
				}
			}()
			var chatMessageReq request.ChatMessageRequest
			data := kafkaMsg.Value
			zlog.Debug("接收到 Kafka 消息", zap.String("data", string(data)))
			// 反序列化 Kafka 获取的数据
			if err := json.Unmarshal(data, &chatMessageReq); err != nil {
				zlog.Error("反序列化 json 数据失败", zap.Error(err))
				continue
			}
			zlog.Debug("反序列化后的数据", zap.Any("data", chatMessageReq))

			// 封装 message 的基本信息
			msg := model.Message{
				Uuid:      util.GenUUID("M"),
				SessionId: chatMessageReq.SessionId,
				Type:      chatMessageReq.Type,
				Content:   chatMessageReq.Content,
				Url:       chatMessageReq.Url,
				SendId:    chatMessageReq.SendId,
				SendName:  chatMessageReq.SendName,
				ReceiveId: chatMessageReq.ReceiveId,
				FileType:  chatMessageReq.FileType,
				FileName:  chatMessageReq.FileName,
				FileSize:  chatMessageReq.FileSize,
				Status:    message_enum.UNSENT, // 标记未发送
				CreatedAt: time.Now(),
				AVdata:    chatMessageReq.AVdata,
			}
			// 对SendAvatar去除前面/static之前的所有内容，防止ip前缀引入
			msg.SendAvatar = normalizePath(chatMessageReq.SendAvatar)
			// 存入数据库
			if err := dao.DB.Create(&msg).Error; err != nil {
				zlog.Error(err.Error())
			}

			// 根据 message 的类型设计不同的转发流程
			msgRep := respond.GetMessageListRespond{
				SendId:     msg.SendId,
				SendName:   msg.SendName,
				SendAvatar: msg.SendAvatar,
				ReceiveId:  msg.ReceiveId,
				Type:       msg.Type,
				Content:    msg.Content,
				Url:        msg.Url,
				FileType:   msg.FileType,
				FileName:   msg.FileName,
				FileSize:   msg.FileSize,
				CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
			}

			// 把数据序列化
			jsonMsg, err := json.Marshal(msgRep)
			if err != nil {
				zlog.Error(err.Error())
			}

			// 准备回显的信息
			msgBack := &MessageBack{
				Message: jsonMsg,
				Uuid:    msg.Uuid,
			}

			switch msgRep.Type {
			case message_enum.TEXT: // 文本
				{
					switch msg.ReceiveId[0] {
					case 'U': // 发给用户
						// 先看接收方在不在线，不在线存表就行，在线的话要发到对应client的SendBack
						// 最后要更新 redis
						s.mu.Lock()
						// 接收方在线时要回显
						if client, online := s.Clients[msg.ReceiveId]; online {
							client.SendBack <- msgBack
						}
						// 发送方在线时要回显
						if client, online := s.Clients[msg.SendId]; online {
							client.SendBack <- msgBack
						}
						s.mu.Unlock()

						// 将消息缓存到 redis
						// key 格式: message_list:send_id:receive_id
						redisKey := fmt.Sprintf("message_list:%s:%s", msg.SendId, msg.ReceiveId)
						repString, err := dao.GetKeyNilIsErr(redisKey)
						if err == nil { // redis 里有，要更新
							// 把取出的数据反序列化
							var rep []respond.GetMessageListRespond
							if err := json.Unmarshal([]byte(repString), &rep); err != nil {
								zlog.Error(err.Error())
							}
							rep = append(rep, msgRep)

							// 序列化后存入 redis
							repByte, err := json.Marshal(rep)
							if err != nil {
								zlog.Error(err.Error())
							}
							err = dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT)
							if err != nil {
								zlog.Error(err.Error())
							}
						} else {
							if !errors.Is(err, redis.Nil) {
								zlog.Error(err.Error())
							} else {
								// redis 里没有，直接存入
								rep := []respond.GetMessageListRespond{msgRep}
								repByte, err := json.Marshal(rep)
								if err != nil {
									zlog.Error(err.Error())
								}
								err = dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT)
								if err != nil {
									zlog.Error(err.Error())
								}
							}
						}
					case 'G': // 发给群聊
						// 群里每一个在线用户的前端都要回显
						// 取出群聊信息
						var group model.GroupInfo
						if err := dao.DB.First(&group, "uuid = ?", msg.ReceiveId).Error; err != nil {
							zlog.Error("群聊信息获取失败", zap.Error(err))
							break
						}

						// 取出成员信息并反序列化
						var members []string
						if err := json.Unmarshal(group.Members, &members); err != nil {
							zlog.Error("反序列化失败", zap.Error(err))
						}

						// 把信息发到每一个在线群成员对应的 SendBack 通道里
						s.mu.Lock()
						for _, member := range members {
							if client, online := s.Clients[member]; online {
								client.SendBack <- msgBack
							}
						}
						s.mu.Unlock()

						// redis
						// key 格式: group_message_list:send_id:receive_id
						redisKey := fmt.Sprintf("group_message_list:%s:%s", msg.SendId, msg.ReceiveId)
						repString, err := dao.GetKeyNilIsErr(redisKey)
						if err == nil {
							var rep []respond.GetMessageListRespond
							if err := json.Unmarshal([]byte(repString), &rep); err != nil {
								zlog.Error("反序列化 redis 获取的字符串失败", zap.Error(err))
							}
							rep = append(rep, msgRep)
							repByte, err := json.Marshal(rep)
							if err != nil {
								zlog.Error("序列化 rep 失败", zap.Error(err))
							}

							if err := dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT); err != nil {
								zlog.Error("redis SetKeyEx 失败", zap.Error(err))
							}
						} else {
							if !errors.Is(err, redis.Nil) {
								zlog.Error(err.Error())
							} else {
								// redis 里没有，直接存入
								rep := []respond.GetMessageListRespond{msgRep}
								repByte, err := json.Marshal(rep)
								if err != nil {
									zlog.Error(err.Error())
								}
								err = dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT)
								if err != nil {
									zlog.Error(err.Error())
								}
							}

						}
					}
				}
			case message_enum.FILE: // 文件
				{
					switch msg.ReceiveId[0] {
					case 'U':
						s.mu.Lock()
						if client, online := s.Clients[msg.ReceiveId]; online {
							client.SendBack <- msgBack
						}
						if client, online := s.Clients[msg.SendId]; online {
							client.SendBack <- msgBack
						}
						s.mu.Unlock()

						// redis
						redisKey := fmt.Sprintf("message_list:%s:%s", msg.SendId, msg.ReceiveId)
						repString, err := dao.GetKeyNilIsErr(redisKey)
						if err == nil {
							var rep []respond.GetMessageListRespond
							if err := json.Unmarshal([]byte(repString), &rep); err != nil {
								zlog.Error("反序列化 redis 获取的字符串失败", zap.Error(err))
							}
							rep = append(rep, msgRep)

							repByte, err := json.Marshal(rep)
							if err != nil {
								zlog.Error("序列化 rep 失败", zap.Error(err))
							}
							if err := dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT); err != nil {
								zlog.Error("redis GetKeyNilIsErr 失败", zap.Error(err))
							}
						} else {
							// redis 里没有，直接存入
							rep := []respond.GetMessageListRespond{msgRep}
							repByte, err := json.Marshal(rep)
							if err != nil {
								zlog.Error(err.Error())
							}
							err = dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT)
							if err != nil {
								zlog.Error(err.Error())
							}
						}
					case 'G':
						var group model.GroupInfo
						if err := dao.DB.First(&group, "uuid = ?", msg.ReceiveId).Error; err != nil {
							zlog.Error("群聊信息获取失败", zap.Error(err))
						}
						var members []string
						if err := json.Unmarshal(group.Members, &members); err != nil {
							zlog.Error("反序列化失败", zap.Error(err))
						}

						s.mu.Lock()
						for _, member := range members {
							if client, online := s.Clients[member]; online {
								client.SendBack <- msgBack
							}
						}
						s.mu.Unlock()

						// redis
						// key 格式: group_message_list:send_id:receive_id
						redisKey := fmt.Sprintf("group_message_list:%s:%s", msg.SendId, msg.ReceiveId)
						repString, err := dao.GetKeyNilIsErr(redisKey)
						if err == nil {
							var rep []respond.GetMessageListRespond
							if err := json.Unmarshal([]byte(repString), &rep); err != nil {
								zlog.Error("反序列化 redis 获取的字符串失败", zap.Error(err))
							}
							rep = append(rep, msgRep)

							repByte, err := json.Marshal(rep)
							if err != nil {
								zlog.Error("序列化 rep 失败", zap.Error(err))
							}
							if err := dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT); err != nil {
								zlog.Error("redis GetKeyNilIsErr 失败", zap.Error(err))
							}
						} else {
							// redis 里没有，直接存入
							rep := []respond.GetMessageListRespond{msgRep}
							repByte, err := json.Marshal(rep)
							if err != nil {
								zlog.Error(err.Error())
							}
							err = dao.SetKeyEx(redisKey, string(repByte), constant.REDIS_TIMEOUT)
							if err != nil {
								zlog.Error(err.Error())
							}
						}
					}
				}
			case message_enum.MEDIA: // 通话 (不用 redis)
				{
					var avData request.AVData
					if err := json.Unmarshal([]byte(chatMessageReq.AVdata), &avData); err != nil {
						zlog.Error("反序列化 AVdata 失败", zap.Error(err))
					}
					if avData.MessageId == "PROXY" && (avData.Type == "start_call" || avData.Type == "receive_call" || avData.Type == "reject_call") {
						if err := dao.DB.Create(&msg).Error; err != nil {
							zlog.Error("创建消息失败", zap.Error(err))
						}
					}

					if msg.ReceiveId[0] == 'U' {
						avMsgRep := respond.AVMessageRespond{
							SendId:     msg.SendId,
							SendName:   msg.SendName,
							SendAvatar: msg.SendAvatar,
							ReceiveId:  msg.ReceiveId,
							Type:       msg.Type,
							Content:    msg.Content,
							Url:        msg.Url,
							FileSize:   msg.FileSize,
							FileType:   msg.FileType,
							CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
							AVdata:     msg.AVdata,
						}
						jsonMsg, err := json.Marshal(avMsgRep)
						if err != nil {
							zlog.Error("AV消息序列化失败", zap.Error(err))
						}
						s.mu.Lock()
						if client, online := s.Clients[msg.ReceiveId]; online {
							client.SendBack <- &MessageBack{
								Message: jsonMsg,
								Uuid:    msg.Uuid,
							}
						}
						s.mu.Unlock()
					}
				}
			}
		}
	}()

	// 监控用户的登入登出信息
	for {
		select {
		case client := <-s.Login:
			{
				s.mu.Lock()
				s.Clients[client.Uuid] = client
				s.mu.Unlock()
				zlog.Info(fmt.Sprintf("用户%s登录", client.Uuid))
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到yuko-chat"))
				if err != nil {
					zlog.Error(err.Error())
				}
			}

		case client := <-s.Logout:
			{
				s.mu.Lock()
				delete(s.Clients, client.Uuid)
				s.mu.Unlock()
				zlog.Info(fmt.Sprintf("用户%s退出登录", client.Uuid))
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("已退出登录")); err != nil {
					zlog.Error(err.Error())
				}
			}
		}
	}
}

// 将https://127.0.0.1:8000/static/xxx 转为 /static/xxx
func normalizePath(path string) string {
	// 查找 "/static/" 的位置
	if path == "https://i.bobopic.com/small/69326036.jpg-216" || path == "" {
		return path
	}
	staticIndex := strings.Index(path, "/static/")
	if staticIndex < 0 {
		log.Println(path)
		zlog.Error("路径不合法")
	}
	// 返回从 "/static/" 开始的部分
	return path[staticIndex:]
}
