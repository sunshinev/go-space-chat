package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	pb "go-space-chat/proto/star"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// 客户端集合
var clients = make(map[*websocket.Conn]*pb.BotStatusRequest)

// 消息缓冲通道
var messages = make(chan *pb.BotStatusRequest, 100)

var addr = flag.String("addr", ":9000", "http service address")
var upgrader = websocket.Upgrader{}

func main() {
	flag.Parse()

	http.HandleFunc("/ws", echo)
	// 广播
	go boardcast()

	// pprof
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	// 这里的ListenAndServe 已经a开启了goroutine协程了
	err := http.ListenAndServe(*addr, nil)
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
		log.Fatalf("升级webcoket %v", err)
	}

	defer func() {
		log.Print("defer 释放连接")
		err := c.Close()
		if err != nil {
			log.Printf(" defer 释放失败 %v", err)
		}

		log.Print("defer 释放成功")
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

			// 清除连接
			delete(clients, c);
			err = c.Close();
			if err != nil {
				log.Printf("连接关闭错误 %v", err)
			}
			log.Print("c 释放成功")

			break
		}

		// 使用protobuf解析
		pbr := &pb.BotStatusRequest{}
		err = proto.Unmarshal(message, pbr)
		if err != nil {
			log.Fatalf("proto 解析失败 %v", err)
			break
		}

		// 初始化链接的id
		if clients[c] == nil {
			clients[c] = &pb.BotStatusRequest{
				BotId:  pbr.GetBotId(),
				Status: pb.BotStatusRequest_connecting,
			}
		}

		messages <- pbr
	}
}

func boardcast() {
	// 始终读取messages
	for msg := range messages {
		// 读取到之后进行广播，启动协程，是为了立即处理下一条msg
		go func() {
			for cli := range clients {
				// protobuf协议
				if clients[cli].BotId == msg.BotId {
					continue
				}

				pbrp := &pb.BotStatusResponse{BotStatus: []*pb.BotStatusRequest{msg}}
				b, err := proto.Marshal(pbrp)
				if err != nil {
					log.Fatalf("proto marshal error %v", err)
				}

				// 二进制发送
				err = cli.WriteMessage(websocket.BinaryMessage, b)
				if err != nil {
					log.Printf("%v", err)
				}
			}
		}()
	}
}
