##  孤独 Lonely

![d2139b33a9868d1f17a471201d1272371588868902.jpg](https://cdn.jsdelivr.net/gh/sunshinev/remote_pics/d2139b33a9868d1f17a471201d1272371588868902.jpg)


## 特色
1. 支持性别修改、并且有颜色替换
2. 支持敏感词过滤
3. 支持姓名修改

## 介绍

通过canvas 2d来模拟了3D的视觉效果。

并且在该项目中使用了protobuf来进行前端和后端的通讯协议，这一点非常方便！

## 操作

1. 项目使用传统`WASD`按键来控制上下左右
2. 眼睛可以跟随鼠标的位置进行转动
3. 按下`space` 空格可以输入消息，按下回车发送消息
4. 左上角按钮可以输入名称，点击空白处名称生效

## docker（推荐）
最新支持使用docker-compose的方式来启动服务，克隆项目后，直接执行下面命令
```
docker-compose up -d
```

访问`http://localhost:8081`


## 本地运行

建议使用非特权端口启动（例如 8081），避免 macOS/Linux 上 `:80` 需要管理员权限。

```bash
go run main.go -web_addr :8081 -socket_addr :9000
```

- Web 静态服务（默认目录：`web_resource/dist/`）：`http://localhost:8081`
- WebSocket 地址：`ws://localhost:9000/ws`
- 统计接口：`http://localhost:8081/login_charts`

如果你需要修改端口：

```bash
go run main.go -web_addr :8082 -socket_addr :9001
```

注意：如果修改了 WebSocket 端口，需要同步修改前端里建立 WebSocket 连接的端口（`web_resource/src` 相关代码）。

### 依赖下载（Go Modules）

如果本地网络环境导致 `go run/go build` 下载依赖失败，可以临时指定 Go 模块代理：

```bash
GOPROXY=https://proxy.golang.com.cn,direct GOSUMDB=off go run main.go -web_addr :8081 -socket_addr :9000
```


## 技术工具

前端 Vue+canvas+websocket+protobuf

后端 Golang+websocket+protobuf+goroutine

## 有意思的难点
> 这里列举几个在实现过程中，遇到的很有意思的问题

1. 如何实现无限画布？
2. 如何实现游戏状态同步？


## proto 文件生成指令
```
protoc -I ./ *.proto --go_out=.
```

```
protoc --js_out=import_style=commonjs,binary:. *.proto

```


## 相关链接

[Canvas 基本用法](https://developer.mozilla.org/zh-CN/docs/Web/API/Canvas_API/Tutorial/Basic_usage)

[Protobuf Guide](https://developers.google.com/protocol-buffers/docs/proto3)

[Vue.js](https://cn.vuejs.org/index.html)