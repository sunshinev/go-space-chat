package core

import (
	"flag"
	"go-space-chat/component"
	pb "go-space-chat/proto/star"
	"html"
	"log"
	"net/http"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Core struct {
	SocketAddr        *string
	WebAddr           *string
	WebsocketUpgrader websocket.Upgrader
	ConnMutex         sync.RWMutex
	Clients           sync.Map // 客户端集合
	TextSafer         component.TextSafe
}

// 广播消息缓冲通道
var messages = make(chan *pb.BotStatusRequest, 1000)

func (c *Core) Run() {
	// 启动参数
	c.SocketAddr = flag.String("socket_addr", ":9000", "socket address")
	c.WebAddr = flag.String("web_addr", ":80", "http service address")

	flag.Parse()

	log.Printf("socket port %s", *c.SocketAddr)
	log.Printf("web port %s", *c.WebAddr)

	// 敏感词初始化
	err := c.TextSafer.NewFilter()
	if err != nil {
		log.Fatalf("text safe new err %v", err)
	}

	// 启动web服务
	go func() {
		err := http.ListenAndServe(*c.WebAddr, http.FileServer(http.Dir("web_resource/dist/")))
		if err != nil {
			log.Fatalf("web 服务启动失败  %v", err)
		} else {
			log.Printf("web 服务启动成功 端口 %s", *c.WebAddr)
		}
	}()

	// 广播
	go c.broadcast()
	// pprof 性能
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// 监听websocket
	http.HandleFunc("/ws", c.websocketUpgrade)

	err = http.ListenAndServe(*c.SocketAddr, nil)
	if err != nil {
		log.Fatalf("create error %v", err)
	}
}

// 升级http为websocket协议
func (c *Core) websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	// 跨域
	c.WebsocketUpgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	// 升级http为websocket
	conn, err := c.WebsocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("http upgrade webcoket err %v", err)
	} else {
		go c.listenWebsocket(conn)
	}
}

// 监听message消息
func (c *Core) listenWebsocket(conn *websocket.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("close websocket err %v", err)
		}
	}()
	// 监听
	for {
		// 尝试查询当前连接
		cInfo, ok := c.Clients.Load(conn)
		if !ok {
			// 写入空
			cInfo = &pb.BotStatusRequest{}
		}

		// 类型断言
		clientInfo, ok := cInfo.(*pb.BotStatusRequest)
		if !ok {
			log.Printf("assert sync map pb.BotStatusRequest err %v", clientInfo)
			c.Clients.Delete(conn)
			continue
		}
		// 读取消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read message error,client: %v break, ip: %v, err:%v", clientInfo.BotId, conn.LocalAddr(), err)
			// 广播关闭连接
			messages <- &pb.BotStatusRequest{
				BotId:  clientInfo.BotId,
				Status: pb.BotStatusRequest_close,
			}
			// 清除连接
			c.Clients.Delete(conn)
			err = conn.Close()
			if err != nil {
				log.Printf("close websocket err %v", err)
			}
			break
		}
		// 使用protobuf解析
		pbr := &pb.BotStatusRequest{}
		err = proto.Unmarshal(message, pbr)
		if err != nil {
			log.Printf("proto parse message %v err %v", message, err)
			continue
		}
		// 敏感词过滤
		pbr.Msg = c.TextSafer.Filter(pbr.Msg)
		pbr.Name = c.TextSafer.Filter(pbr.Name)
		// 过滤html 标签
		pbr.Msg = html.EscapeString(pbr.Msg)
		pbr.Name = html.EscapeString(pbr.Name)

		// 初始化链接的ID
		if clientInfo.BotId == "" {
			c.Clients.Store(conn, &pb.BotStatusRequest{
				BotId:  pbr.GetBotId(),
				Status: pb.BotStatusRequest_connecting,
			})
		}
		// 广播队列
		messages <- pbr
	}
}

// 广播
func (c *Core) broadcast() {
	// 始终读取messages
	for msg := range messages {
		if msg.Msg != "" {
			log.Printf("%s : %s", msg.BotId+":"+msg.Name, msg.Msg)
		}

		// 读取到之后进行广播，启动协程，是为了立即处理下一条msg
		go func(m pb.BotStatusRequest) {
			// 遍历所有客户
			c.Clients.Range(func(connKey, bs interface{}) bool {

				//bot, ok := bs.(*pb.BotStatusRequest)
				//if !ok {
				//	return true
				//}
				// 不给自己发消息
				//if bot.BotId == m.BotId {
				//	return true
				//}

				resp := &pb.BotStatusResponse{
					BotStatus: []*pb.BotStatusRequest{&m},
				}
				b, err := proto.Marshal(resp)
				if err != nil {
					log.Printf("proto marshal error %v %+v", err, resp)
					return true
				}

				// 二进制发送
				conn, ok := connKey.(*websocket.Conn)
				if !ok {
					log.Printf("assert connkey websocket.Conn err %v", conn)
					return true
				}
				// 防止并发写
				c.ConnMutex.Lock()
				err = conn.WriteMessage(websocket.BinaryMessage, b)
				c.ConnMutex.Unlock()
				if err != nil {
					log.Printf("conn write message err %v", err)
				}
				return true
			})
		}(*msg)
	}
}
