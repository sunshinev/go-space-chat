package main

import (
	"bufio"
	"flag"
	pb "go-space-chat/proto/star"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"

	filter "github.com/antlinker/go-dirtyfilter"

	"github.com/antlinker/go-dirtyfilter/store"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
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

var words = []string{}

var filterManage *filter.DirtyManager

func main() {
	flag.Parse()

	filterManage = readWords()

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
		c.Close()
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

		// 敏感词过滤
		pbr.Msg = wordsFilter(pbr.Msg)
		pbr.Name = wordsFilter(pbr.Name)

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

				// 不给自己发消息
				if clients[cli].BotId == msg.BotId {
					continue
				}

				//log.Print(msg)
				//log.Print(clients[cli])
				//// 距离太远的用户就没必要发送消息了
				//if math.Abs(float64(clients[cli].RealX-msg.RealX+clients[cli].X-msg.X)) > 200 {
				//	continue
				//}
				//
				//// 距离太远的用户就没必要发送消息了
				//if math.Abs(float64(clients[cli].RealY-msg.RealY+clients[cli].Y-msg.Y)) > 200 {
				//	continue
				//}

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

func readWords() *filter.DirtyManager {
	fi, err := os.Open("words/gg.txt")
	if err != nil {
		panic(err.Error())
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		words = append(words, string(a))
	}

	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		DataSource: words,
	})
	if err != nil {
		panic(err)
	}
	return filter.NewDirtyManager(memStore)
}

func wordsFilter(filterText string) string {

	result, err := filterManage.Filter().Filter(filterText, '*', '@')
	if err != nil {
		panic(err)
	}

	if result != nil {
		for _, w := range result {
			filterText = strings.ReplaceAll(filterText, w, "*")
		}
	}

	return filterText
}
