package chat

import (
	"encoding/json"
	"net/http"
	"yuko_chat/internal/config"
	"yuko_chat/internal/dao"
	"yuko_chat/internal/dto/request"
	"yuko_chat/internal/model"
	"yuko_chat/pkg/constant"
	message_enum "yuko_chat/pkg/enum/message"
	"yuko_chat/pkg/zlog"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	// 检查连接的Origin头
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type MessageBack struct {
	Message []byte
	Uuid    string
}

type Client struct {
	Conn     *websocket.Conn
	Uuid     string
	SendBack chan *MessageBack // 给前端
	// SendTo   chan []byte       // server端的 transmit 的缓冲，用了 kafka 就不用了
}

// 从 ws 连接里读数据再发给 server
// 这里用 kafka 当缓冲，server 从kafka中取数据
func (c *Client) Read() {
	for {
		_, jsonMsg, err := c.Conn.ReadMessage()
		if err != nil {
			zlog.Error("websocket 断开", zap.Error(err), zap.String("clientId", c.Uuid))
			return // 直接断开 ws 连接
		}

		// 反序列化数据，只为了看看他长啥样，方便调试
		var msg request.ChatMessageRequest
		if err := json.Unmarshal(jsonMsg, &msg); err != nil {
			zlog.Error("反序列化消息失败", zap.Error(err))
		}
		zlog.Debug("收到消息", zap.Any("message", msg))

		// 把数据当作消息发给 kafka
		_, _, err = dao.SendMessage(config.Cfg.KafkaConfig.Topic, jsonMsg)
		if err != nil {
			zlog.Error(err.Error())
		}
		zlog.Debug("消息已发送")
	}
}

// // 从 SendBack 通道读取消息发送给websocket
func (c *Client) Write() {
	for msg := range c.SendBack {
		// 写到 ws 连接上
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg.Message); err != nil {
			zlog.Error(err.Error())
			return
		}

		// 修改数据库中 message 的记录为已发送
		err := dao.DB.Model(&model.Message{}).Where("uuid = ?", msg.Uuid).Update("status", message_enum.SENT).Error
		if err != nil {
			zlog.Error(err.Error())
		}
	}
}

func ClientLogin(c *gin.Context, clientId string) {
	// 把 HTTP 连接升级成 web socket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(err.Error())
	}

	client := Client{
		Conn:     conn,
		Uuid:     clientId,
		SendBack: make(chan *MessageBack, constant.CHANNEL_SIZE),
	}

	// 让 server 标记用户上线
	Myserver.SendClientToLogin(&client)

	go client.Read()
	go client.Write()
	zlog.Info("ws 连接成功")
}

// ClientLogout 当接受到前端有登出消息时，会调用该函数
func ClientLogout(clientId string) (string, int) {
	client := Myserver.Clients[clientId]
	if client != nil {
		Myserver.SendClientToLogout(client)
	}

	// 关闭 ws 连接
	if err := client.Conn.Close(); err != nil {
		zlog.Error(err.Error())
		return constant.SYS_ERR_MSG, -1
	}

	// 关闭 client 的 SendBack 通道，同时 client.Write() 协程会结束
	close(client.SendBack)
	return "退出成功", 0
}
