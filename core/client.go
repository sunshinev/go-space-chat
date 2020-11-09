package core

import (
	pb "go-space-chat/proto/star"
	"html"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Client struct {
	botId     string
	dirty     bool
	status    *pb.BotStatusRequest
	msgs      []string
	writeChan chan []byte
	closeChan chan bool
	conn      *websocket.Conn
}

func (c *Client) start() {
	go c.waitAndWrite()
	go c.readAndServe()
}

func (c *Client) close() {
	close(c.closeChan)
	c.conn.Close()
}

func (c *Client) readAndServe() {
	defer func() {
		c.close()
	}()
	for {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("read message error,client: %v break, ip: %v, err:%v", c.botId, c.conn.LocalAddr(), err)
			messages <- &pb.BotStatusRequest{
				BotId:  c.botId,
				Status: pb.BotStatusRequest_close,
			}
			// 下线广播
			if c.status != nil {
				messages <- &pb.BotStatusRequest{
					BotId: c.status.BotId,
					Name:  "系统管理员",
					Msg:   "用户@" + c.status.Name + " 下线了",
				}
			}
			core.Clients.Delete(c.conn)
			break
		}
		req := new(pb.BotStatusRequest)
		err = proto.Unmarshal(message, req)
		if err != nil {
			log.Println(err)
			break
		}
		if req.BotId != "" {
			c.botId = req.BotId
		}
		// 文本过滤
		req.Msg = core.TextSafer.Filter(req.Msg)
		req.Name = core.TextSafer.Filter(req.Name)
		req.Msg = html.EscapeString(req.Msg)
		req.Name = html.EscapeString(req.Name)

		if req.Msg != "" {
			messages <- req
			continue
		}
		if c.status == nil {
			// 上线广播
			messages <- &pb.BotStatusRequest{
				BotId: req.BotId,
				Name:  "系统管理员",
				Msg:   "用户@" + req.Name + " 上线啦",
			}
		}
		c.status = req
		c.dirty = true
	}
}

func (c *Client) waitAndWrite() {
	for {
		select {
		case <-c.closeChan:
			time.Sleep(time.Second)
			close(c.writeChan)
			return
		case b, ok := <-c.writeChan:
			if !ok { // channel 关闭了
				return
			}
			err := c.conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				log.Println("write err", err.Error())
			}
		}
	}
}

func (c *Client) write(b []byte) {
	defer func() {
		// 防止写到关闭的writeChan里，但是做了延迟关闭 应该不会发生
		r := recover()
		if r != nil {
			log.Println(r)
		}
	}()
	select {
	case <-c.closeChan:
		return
	case c.writeChan <- b:
	}
}
