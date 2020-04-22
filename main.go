package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	pb "go-space-chat/proto/star"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
)

// 客户端集合
var clients = make(map[*websocket.Conn]*pb.BotStatusRequest)

var clients_mutex sync.RWMutex

var conn_mutex sync.RWMutex

// 消息缓冲通道
var messages = make(chan *pb.BotStatusRequest, 100)

var socket_addr = flag.String("socket_addr", ":9000", "socket address")
var web_addr = flag.String("web_addr", ":80", "http service address")

var upgrader = websocket.Upgrader{}

func main() {
	flag.Parse()

	go http.ListenAndServe(*web_addr, http.FileServer(http.Dir("web_resource/dist/")))

	log.Printf("web 服务启动成功 端口 %s", *web_addr)

	http.HandleFunc("/ws", echo)
	// 广播
	go boardcast()

	// pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Printf("socket 服务启动端口 %s", *socket_addr)
	// 这里的ListenAndServe 已经a开启了goroutine协程了
	err := http.ListenAndServe(*socket_addr, nil)
	if err != nil {
		log.Fatalf("create error %v", err)
	}
}

// 这个echo是在serve协程里面运行的
func echo(w http.ResponseWriter, r *http.Request) {
	// 跨域
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	// 升级http为websocket
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("升级webcoket %v", err)
		w.Write([]byte(err.Error()))
	} else {
		go checkConn(c)
	}
}

func checkConn(c *websocket.Conn) {
	defer func() {
		err := c.Close()
		if err != nil {
			log.Printf(" defer 释放失败 %v", err)
		} else {
			log.Print("defer 释放成功")
		}
	}()
	// 监听
	for {
		_, message, err := c.ReadMessage()

		if err != nil {
			log.Printf("read error 读取失败 ", err)
			messages <- &pb.BotStatusRequest{
				BotId: clients[c].BotId,
				// 广播关闭连接
				Status: pb.BotStatusRequest_close,
			}

			clients_mutex.Lock()
			// 清除连接
			delete(clients, c)
			clients_mutex.Unlock()

			break
		}

		// 使用protobuf解析
		pbr := &pb.BotStatusRequest{}
		err = proto.Unmarshal(message, pbr)
		if err != nil {
			log.Printf("proto 解析失败 %v", err)
			break
		}

		// 初始化链接的id
		if clients[c] == nil {

			clients_mutex.Lock()
			clients[c] = &pb.BotStatusRequest{
				BotId:  pbr.GetBotId(),
				Status: pb.BotStatusRequest_connecting,
			}
			clients_mutex.Unlock()
		}

		messages <- pbr
	}
}

func boardcast() {
	// 始终读取messages
	for msg := range messages {
		if msg.Msg != "" {
			log.Printf("%s : %s", msg.BotId+":"+msg.Name, msg.Msg)
		}
		// 读取到之后进行广播，启动协程，是为了立即处理下一条msg
		go func() {
			clients_mutex.RLock()
			defer clients_mutex.RUnlock()
			for cli := range clients {
				// protobuf协议
				if clients[cli].BotId == msg.BotId {
					continue
				}

				pbrp := &pb.BotStatusResponse{BotStatus: []*pb.BotStatusRequest{msg}}
				b, err := proto.Marshal(pbrp)
				if err != nil {
					log.Printf("proto marshal error %v", err)
					continue
				}

				// 二进制发送
				conn_mutex.Lock()
				err = cli.WriteMessage(websocket.BinaryMessage, b)
				conn_mutex.Unlock()
				if err != nil {
					log.Printf("%v", err)
				}
			}
		}()
	}
}
