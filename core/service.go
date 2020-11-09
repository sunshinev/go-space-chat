package core

import (
	"flag"
	"go-space-chat/component"
	pb "go-space-chat/proto/star"
	"log"
	"net/http"
	"sync"
	"time"

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

var core *Core

func InitCore() {
	core = new(Core)
	core.Run()
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

	c.cronTask()

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

func (c *Core) cronTask() {
	// 统计在线人数
	go func() {
		for {
			time.Sleep(time.Minute)
			count := 0
			c.Clients.Range(func(_, _ interface{}) bool {
				count++
				return true
			})
			log.Println("当前在线连接数: ", count)
		}
	}()
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
		var cli *Client
		v, ok := c.Clients.Load(conn)
		if !ok { // 创建一个新的客户端
			cli = new(Client)
			cli.conn = conn
			cli.dirty = true
			cli.writeChan = make(chan []byte, 64)
			cli.closeChan = make(chan bool)
			c.Clients.Store(conn, cli)
		} else {
			cli, _ = v.(*Client)
		}
		if cli != nil {
			cli.start()
		}
	}
}

// 广播
func (c *Core) broadcast() {
	// 始终读取messages
	// 聊天消息通过这里广播
	go func() {
		for msg := range messages {
			if msg.Msg != "" {
				log.Printf("%s : %s", msg.BotId+":"+msg.Name, msg.Msg)
			}

			// 读取到之后进行广播，启动协程，是为了立即处理下一条msg
			// 遍历所有客户
			c.Clients.Range(func(connKey, value interface{}) bool {

				resp := &pb.BotStatusResponse{
					BotStatus: []*pb.BotStatusRequest{msg},
				}
				b, err := proto.Marshal(resp)
				if err != nil {
					log.Printf("proto marshal error %v %+v", err, resp)
					return true
				}

				cli, ok := value.(*Client)
				if !ok {
					return true
				}
				cli.write(b)
				return true
			})
		}
	}()

	// 广播全服状态, 50ms一次
	go func() {
		for {
			time.Sleep(time.Millisecond * 50)
			resp := &pb.BotStatusResponse{}
			// 遍历所有客户端状态
			c.Clients.Range(func(connKey, value interface{}) bool {
				cli, ok := value.(*Client)
				if !ok {
					return true
				}
				if !cli.dirty {
					return true
				}
				if cli.status != nil {
					resp.BotStatus = append(resp.BotStatus, cli.status)
					cli.dirty = false
				}
				return true
			})

			if len(resp.BotStatus) == 0 {
				continue
			}

			b, _ := proto.Marshal(resp)

			c.Clients.Range(func(connKey, value interface{}) bool {
				cli, ok := value.(*Client)
				if !ok {
					return true
				}
				cli.write(b)
				return true
			})
		}
	}()
}
