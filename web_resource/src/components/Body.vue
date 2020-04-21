<template>
    <Content>
        <!--
        <br>
        <Row style="text-align: left">
            <Col span="12" offset="6">
                <Input type="email" maxlength="10" style="width: 200px; margin:20px 0" placeholder="Email" v-model="email"></Input>
                <Input type="textarea" :autosize="{minRows: 2,maxRows: 5}" placeholder="Enter something..." style="margin-bottom: 20px" v-model="content"></Input>
                <Button type="primary" @click="wsSend"> Send</Button>
            </Col>
        </Row>
        <Row style="margin-top: 20px">

            <Col span="12" offset="6">
                <Card style="text-align: left;margin-top: 20px" v-for="item in message_list" v-bind:key="item.email">
                    <p slot="title">{{ item.email }}</p>
                    <p>{{ item.content }}</p>
                    <p>{{item.create_at}}</p>
                </Card>
            </Col>
        </Row>
        -->
        <canvas id="test">

        </canvas>
    </Content>
</template>
<style>
    canvas {
        border: 1px solid green;
    }
</style>
<script>
    import "google-protobuf"
    import proto from  "../proto/chat_pb.js"


    export default {
        data () {
            return {
                username:'',
                email:'',
                content:'',
                message_list:[
                    {
                        email:'',
                        content:'',
                    }
                ],
                ws:null,
                visual:{
                    x:0,
                    y:0,
                    z:100
                },
                w:1000,
                h:500,
                points:this.randomPoint(500)
            }
        },
        methods: {
            handleRecv:function(data) {
                // var jsonData = JSON.parse(data)
                var rep = proto.ChatResponse.deserializeBinary(data)

                console.info(rep.getData())

                console.info(rep.getCode())


                this.message_list.unshift({email:rep.getData().getEmail(),content:rep.getData().getContent()})
            },
            wsOpen: function () {
                var that = this
                var ws = new WebSocket("ws://localhost:9000/ws")

                ws.binaryType = 'arraybuffer';

                ws.onopen = function () {
                    console.info("ws open")
                }
                
                ws.onmessage = function (evt) {
                    console.info(evt)
                    console.info("Received message:"+evt.data)
                    that.handleRecv(evt.data)
                }

                ws.onclose  = function () {
                    console.info("ws close")
                }

                this.ws = ws
            },
            wsSend: function() {
                if(this.ws == null) {
                    console.info("连接尚未打开")
                }

                var chat = new proto.ChatRequest()
                chat.setEmail(this.email)
                chat.setContent(this.content)

                this.ws.send(chat.serializeBinary())
            },

            canvasHandle: function() {
                var canvas = document.getElementById("test");
                var ctx = canvas.getContext('2d');
                // console.info(this.w,this.h)

                canvas.width = 1000;
                canvas.height = 500;

                ctx.fillStyle = "rgb(0,0,0)"
                ctx.fillRect(0,0,canvas.width,canvas.height)
                ctx.save();

                this.visual.x += 1;

                if(this.visual.x == 1000) {
                    this.visual.x = 0;
                }

                ctx.restore()
                this.drawn(ctx, this.points);

                window.requestAnimationFrame(this.canvasHandle);
            },

            drawn:function(ctx, points = []) {
                for (var i in points) {
                    var p = this.pointConvert(points[i].x,points[i].y,points[i].z,points[i].c,points[i].s);
                    this.drawnPoint(ctx,p.x,p.y,p.c,p.s)
                }
            },
            drawnPoint:function(ctx,x,y,c,s) {
                ctx.beginPath();
                ctx.arc(x,y,s,0,2 * Math.PI)
                ctx.fillStyle = c;
                ctx.fill()
            },
            // 核心转换算法
            pointConvert:function (x, y, z,c,s, offsetX =0 , offsetY = 0) {
                return {
                    x: (x - this.visual.x) * this.visual.z / (this.visual.z - z) + offsetX,
                    y: (y - this.visual.y) * this.visual.z / (this.visual.z - z) + offsetY,
                    c:c,
                    s:s,
                }
            },
            randomPoint:function(count = 100) {
                var points =[];
                for(var i =0;i<count;i++) {
                    var x = this.randomValue(1000);
                    var y = this.randomValue(500);
                    var z = this.randomValue(50);
                    // 根据z轴 色彩递减
                    var c = this.computeColor(z);
                    var s = this.computeSize(z);

                    points.push({x:x,y:y,z:z,c:c,s:s})
                }

                points.sort(function (a,b) {
                    return a.z - b.z;
                })

                return points;
            },
            randomValue:function(max) {
                return Math.floor(Math.random()*max)
            },
            // 颜色越来越深
            computeColor: function(z) {
                var v = Math.floor((z * (255-100))/50)+100
                return "rgb("+v+","+v+","+v+")";
            },
            // 尺寸越来越小
            computeSize:function(z) {
                // 设定视点处，半径为10
                return z*10/300;
            }
        },
        mounted(){
            this.wsOpen();
            window.requestAnimationFrame(this.canvasHandle);
        }
    }
</script>
