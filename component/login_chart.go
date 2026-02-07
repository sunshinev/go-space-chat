package component

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sunshinev/go-space-chat/config"
)

// LoginChart ...
type LoginChart struct {
	today      string          // 日志记录的日期，只保留一天
	seenLocker sync.Mutex      // 去重锁
	seenBotIDs map[string]bool // 当天已记录过的 botId
}

// 十分钟为粒度
var timeSpan float64 = 10

// 入口通道（按 botId 去重）
var entryChannel = make(chan string, 100)

// 数据记录
var records sync.Map

// 点位数据结构
type posData struct {
	TimeSort int64
	Time     string `json:"time"`
	Num      int32  `json:"num"`
}

// 初始化
func InitLoginChart() *LoginChart {
	login := &LoginChart{
		today:      time.Now().Format(config.DateFormatDay),
		seenBotIDs: make(map[string]bool),
	}
	// 开启消费
	go login.consume()

	return login
}

// Entry 记录一次“用户上线”事件，同一个 botId 在同一天内只统计一次
func (s *LoginChart) Entry(botID string) {
	entryChannel <- botID
}

// 消费数据
func (s *LoginChart) consume() {
	// 用chan 主要是为了防止并发add
	for botID := range entryChannel {
		s.add(botID)
	}
}

// 添加数据记录
func (s *LoginChart) add(botID string) {
	// 是否需要重置数据？
	s.isClean()

	// 同一天内，同一个 botId 只统计一次
	if botID != "" {
		s.seenLocker.Lock()
		if s.seenBotIDs[botID] {
			s.seenLocker.Unlock()
			return
		}
		s.seenBotIDs[botID] = true
		s.seenLocker.Unlock()
	}

	now := time.Now()
	min := now.Minute()
	posMin := math.Ceil(float64(min) / timeSpan)

	// 生成key
	key := fmt.Sprintf("%v:%v", now.Hour(), (posMin-1)*timeSpan)
	// 取出当前值
	value, ok := records.Load(key)
	if !ok {
		value = &posData{
			TimeSort: now.Unix(),
			Time:     key,
			Num:      0,
		}
	}

	if pos, ok := value.(*posData); ok {
		// +1 计数
		pos.Num = pos.Num + 1
		records.Store(key, pos)
	}
}

func (s *LoginChart) isClean() {
	today := time.Now().Format(config.DateFormatDay)
	if today != s.today {
		// 复写日期
		s.today = today
		// 清除所有数据
		records.Range(func(key, value interface{}) bool {
			records.Delete(key)
			return true
		})

		// 清空当天已记录过的 botId
		s.seenLocker.Lock()
		s.seenBotIDs = make(map[string]bool)
		s.seenLocker.Unlock()
	}
}

type ChartData struct {
	X string `json:"x"`
	Y int32  `json:"y"`
}

// ChartDataApi 获取所有数据
func (s *LoginChart) FetchAllData() []ChartData {

	xSlice := []string{}
	ySlice := []int32{}

	realData := map[string]int32{}

	records.Range(func(key, value interface{}) bool {
		if pos, ok := value.(*posData); ok {

			xSlice = append(xSlice, pos.Time)
			ySlice = append(ySlice, pos.Num)

			realData[pos.Time] = pos.Num
		}

		return true
	})

	data := []ChartData{}

	for i := 0; i < 24; i++ {
		for j := 0; j < 60; j += 10 {
			newKey := fmt.Sprintf("%v:%v", i, j)
			item := ChartData{
				X: newKey,
				Y: 0,
			}
			if y, ok := realData[newKey]; ok {
				item.Y = y
			}

			data = append(data, item)
		}
	}

	return data
}
