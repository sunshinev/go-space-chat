package core

import (
	"encoding/json"
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
	loginChart        *component.LoginChart
	IpSearch          *component.IpSearch
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
	// 初始日志记录
	c.loginChart = component.InitLoginChart()
	// 初始化ip转换
	c.IpSearch = component.InitIpSearch()

	// 启动web服务
	SafeGo(func() {
		http.HandleFunc("/login_charts", c.ChartDataApi)
		http.Handle("/", http.FileServer(http.Dir("web_resource/dist/")))

		err := http.ListenAndServe(*c.WebAddr, nil)
		if err != nil {
			log.Fatalf("web 服务启动失败  %v", err)
		} else {
			log.Printf("web 服务启动成功 端口 %s", *c.WebAddr)
		}
	})

	// 广播
	SafeGo(func() {
		c.broadcast()
	})
	// pprof 性能
	SafeGo(func() {
		log.Println(http.ListenAndServe(":6060", nil))
	})

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
		SafeGo(func() {
			c.listenWebsocket(conn)
		})
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
		clientInfo, ok := cInfo.(*pb.BotStatusRequest)
		if !ok {
			log.Printf("assert sync map pb.BotStatusRequest err %v", clientInfo)
			c.Clients.Delete(conn)
			continue
		}
		// 读取消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read message error,client: %v break, ip: %v, err:%v", clientInfo.BotId, conn.RemoteAddr(), err)
			messages <- &pb.BotStatusRequest{
				BotId:   clientInfo.BotId,
				Name:    clientInfo.Name,
				Msg:     "我下线了~拜拜~",
				PosInfo: clientInfo.PosInfo,
			}
			// 广播关闭连接
			messages <- &pb.BotStatusRequest{
				BotId:  clientInfo.BotId,
				Status: pb.BotStatusRequest_close,
			}
			// 清除用户
			c.Clients.Delete(conn)
			// 关闭连接
			err = conn.Close()
			if err != nil {
				log.Printf("close websocket err %v", err)
			}
			break
		}
		// 消息读取成功，解析消息
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

		// 如果是新用户初始化链接的ID
		if clientInfo.BotId == "" {
			// 获取地理位置
			posInfo := pb.PInfo{}
			pinfo, err := c.IpSearch.Search(conn.RemoteAddr().String())
			if err != nil {
				log.Printf("ip search err %v", err)
			} else {
				posInfo = pb.PInfo{
					CityId:   int32(pinfo.CityId),
					Country:  pinfo.Country,
					Region:   pinfo.Region,
					Province: pinfo.Province,
					City:     pinfo.City,
					Isp:      pinfo.ISP,
				}
			}
			c.Clients.Store(conn, &pb.BotStatusRequest{
				BotId:   pbr.GetBotId(),
				Name:    pbr.GetName(),
				Status:  pb.BotStatusRequest_connecting,
				PosInfo: &posInfo,
			})
			// 新用户进行上线提示
			pbr.Msg = "我上线啦~大家好呀"
			pbr.PosInfo = &posInfo
			// 新用户上线，记录次数
			c.loginChart.Entry()
		} else {
			// 老用户直接从clients获取pos信息
			pbr.PosInfo = clientInfo.PosInfo
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

type ChartApiRsp struct {
	X []string `json:"x"`
	Y []int32  `json:"y"`
}

func (c *Core) ChartDataApi(w http.ResponseWriter, r *http.Request) {
	chartData := c.loginChart.FetchAllData()

	xSlice := []string{}
	ySlice := []int32{}

	for _, v := range chartData {
		xSlice = append(xSlice, v.X)
		ySlice = append(ySlice, v.Y)
	}

	data := &ChartApiRsp{
		X: xSlice,
		Y: ySlice,
	}

	d, err := json.Marshal(data)
	if err != nil {
		log.Printf("ChartDataApi marsharl %v", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, err = w.Write(d)
	if err != nil {
		log.Printf("ChartDataApi write %v", err)
	}
}
