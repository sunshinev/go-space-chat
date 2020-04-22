##  孤独 Lonely

![abc7178898c1ead114f64ec437cb41f81587469886.jpg](https://github.com/sunshinev/remote_pics/raw/master/abc7178898c1ead114f64ec437cb41f81587469886.jpg)

## Demo

http://chat.osinger.com/


## 介绍

项目打造了一个模拟太空的环境，通过canvas 2d来模拟了3D的视觉效果。

并且在该项目中使用了protobuf来进行前端和后端的通讯协议，这一点非常方便！

## 操作

1. 项目使用传统`WASD`按键来控制上下左右
2. 眼睛可以跟随鼠标的位置进行转动
3. 按下`space` 空格可以输入消息，按下回车发送消息
4. 左上角按钮可以输入名称，点击空白处名称生效


## 运行

```$xslt
go run main.go
```

该命令会启动web-server作为静态服务，默认80端口，如果需要修改端口，用下面的命令
```
go run main.go -web_server 8081
```

项目启动默认websocket服务端口为9000端口，如果需要修改
```
go run main.go -socket_server 9001
```
注意：如果修改websocket端口，同时需要修改js里面的socket端口


## 技术工具

前端 Vue+canvas+websocket+protobuf

后端 Golang+websocket+protobuf+goroutine

## 有意思的难点
> 这里列举几个在实现过程中，遇到的很有意思的问题

1. 如何实现无限画布？
2. 如何实现游戏状态同步？



## 相关链接

[Canvas 基本用法](https://developer.mozilla.org/zh-CN/docs/Web/API/Canvas_API/Tutorial/Basic_usage)

[Protobuf Guide](https://developers.google.com/protocol-buffers/docs/proto3)

[Vue.js](https://cn.vuejs.org/index.html)