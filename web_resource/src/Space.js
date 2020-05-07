/**
 * author huajie.sun
 * from china
 * 2020年04月21日15:41:35
 * https://github.com/sunshinev/go-space-chat
 * license MIT
 * @type {null}
 */
import "google-protobuf"
import "./proto/star_pb.js"

var ctx = null;
var canvas = null;

var points = [];

// 每个空间产生的star数量
var points_num = 200;

var visual = {
    x: 0,
    y: 0,
    z: 10
};

// 模拟物体真实尺寸
var visual_star_size = 3;

var keys = {
    up: 87, // w
    down: 83, // s
    left: 65, //a
    right: 68, //d
    talk: 32, //space
    send: 13
};

// 矩形坐标
var matrix_poi = {
    x: 0,
    y: 0
};

// 三维粒子分布范围
var points_scope = {
    x: {
        min: 0,
        max: 2000,
    },
    y: {
        min: 0,
        max: 2000,
    },
    z: {
        min: 0,
        max: 10,
    }
};

// 移动速度
var move_speed = 2;

// 移动方向
var move_direction = {
    up: false,
    down: false,
    left: false,
    right: false
};

// star随机移动的方向和距离
var random_points_update = [0.5, 0, 0];
var random_points_update_key = true;
var random_points_update_status = 0;

// 缩放值
var zoom_scope_value = 0;

// 当前画布序号
var current_space = {
    x: 0,
    y: 0
};

var mouse_poi = {x: 0, y: 0};

// 输入框dom
var input = null;
// 如果开启input，禁止移动,默认关闭
var input_deny_move_key = false;
// 输入的消息队列
var input_messages_queue = [];
var show_message_box = null;

// websocket
var ws = null;

var bot_status = {
    x: 0,
    y: 0,
    e_x: 0,
    e_y: 0,
    r_x: 0,
    r_y: 0,
    bot_id: '',
    name: '',
    gender: 0
};

// 状态记忆
var bot_status_old = {};

var is_ws_open = false;

var real_top_left_poi = {
    x: 0,
    y: 0
};

// 维护客户列表
var guest_bots = {};
var guest_show_message_box = {};


function initCtx() {
    canvas = document.getElementById("test");
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    ctx = canvas.getContext('2d');

    ctx.shadowColor = "white";
    ctx.shadowBlur = 10;

    matrix_poi = {
        x: canvas.width / 2,
        y: canvas.height / 2
    }

    // 初始化视点
    visual.x = canvas.width / 2;
    visual.y = canvas.height / 2;

    // 初始化粒子范围
    points_scope.x.max = canvas.width;
    points_scope.y.max = canvas.height;

    // 初始化粒子
    points = randomPoint();

    bot_status.bot_id = Math.random().toString(36).substr(2);

    randomPointsUpdate()
}

function canvasHandle() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    ctx.fillStyle = "rgb(20,7,34)";
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    immediateUpdate();

    points = updatePoints(points, random_points_update);

    drawn(ctx, points);
    matrixMove();

    // 维护状态
    bot_status.x = visual.x;
    bot_status.y = visual.y;

    // 尝试同步消息
    sendStatusByWs();

    createGuestBot();

    window.requestAnimationFrame(canvasHandle);
}

// star 自然振动
function randomPointsUpdate() {

    setInterval(function () {
        if (random_points_update_key) {
            random_points_update = [
                Math.random() - 0.5,
                Math.random() - 0.5,
                0
            ];
        }
    }, 5000)

}

function immediateUpdate() {
    if (random_points_update_status) {
        random_points_update = [
            Math.random() - 0.5,
            Math.random() - 0.5,
            0
        ];
        random_points_update_status = 0;
    }
}

function randomPoint() {

    // 计算每个space的宽高
    var width = points_scope.x.max - points_scope.x.min;
    var height = points_scope.y.max - points_scope.y.min;

    var p = [];
    for (var i = 0; i < points_num; i++) {
        var x = randomValue(current_space.x + width, current_space.x);
        var y = randomValue(current_space.y + height, current_space.y);
        var z = randomValue(points_scope.z.max, points_scope.z.min);
        // 根据z轴 色彩递减
        var c = computeColor(z);
        var s = computeSize(z);

        p.push({x: x, y: y, z: z, c: c, s: s})
    }


    p.sort(function (a, b) {
        return a.z - b.z;
    });

    return p;
}


function updatePoints(points, values = [0, 0, 0]) {


    for (var i in points) {
        var p = points[i];
        p.x += values[0];
        p.y += values[1];
        p.z += values[2];

        // 重要：实现无限star！这个地方要保证粒子的绘制范围x,y 在canvas之内
        if (p.x > canvas.width) {
            p.x -= canvas.width;
        } else if (p.x < 0) {
            p.x += canvas.width;
        }

        if (p.y > canvas.height) {
            p.y -= canvas.height;
        } else if (p.y < 0) {
            p.y += canvas.height;
        }

        points[i] = p;
    }

    return points;
}

function randomValue(max, min = 0) {
    return Math.floor(Math.random() * (max - min)) + min
}

// 颜色越来越深
function computeColor(z) {
    var v = Math.floor((z * (255 - 100)) / (points_scope.z.max - points_scope.z.min)) + 100
    return "rgb(" + v + "," + v + "," + v + ")";
}

// 尺寸越来越小
function computeSize(z) {
    // 设定视点处，半径为10
    return z * visual_star_size / visual.z;
}

// 绘制star
function drawn(ctx, points = []) {
    for (var i in points) {
        var p = pointConvert(points[i].x, points[i].y, points[i].z, points[i].c, points[i].s);
        // 这里要做一些舍弃动作,视野之外的粒子，不予绘制
        if (p.x < canvas.width || p.y < canvas.height) {
            drawnPoint(ctx, p.x, p.y, p.c, p.s)
        }
    }
}

// 核心转换算法
function pointConvert(x, y, z, c, s) {
    var p = {
        x: (x - visual.x) * visual.z / (visual.z - z) + visual.x,
        y: (y - visual.y) * visual.z / (visual.z - z) + visual.y,
        c: c,
        s: s,
    }

    return p;
}

function drawnPoint(ctx, x, y, c, s) {
    ctx.beginPath();
    ctx.arc(x, y, s, 0, 2 * Math.PI)
    ctx.fillStyle = c;
    ctx.fill()
}

// 绘制矩形
function drawMatrix(x = canvas.width / 2, y = canvas.height / 2) {
    // 矩形
    // ctx.fillStyle = "rgb(200,0,0)"
    // ctx.fillRect(x, y, 50, 30);

    // 眼睛
    ctx.beginPath();
    x = x + 5;
    y = y + 5;
    ctx.arc(x, y, 8, 0, 2 * Math.PI);
    if (bot_status.gender === proto.botStatusRequest.gender_type.WOMAN) {
        ctx.fillStyle = "rgb(255,20,147)";
    } else {
        ctx.fillStyle = "rgb(0,191,255)";
    }

    ctx.fill();

    // 瞳孔
    ctx.beginPath()
    var pupil = moveEye(x, y, 8);

    // 维护botstatus eye
    bot_status.e_x = pupil[0];
    bot_status.e_y = pupil[1];

    ctx.arc(pupil[0], pupil[1], 4, 0, 2 * Math.PI);
    ctx.fillStyle = "rgb(255,255,255)";
    ctx.fill()

    ctx.font = "12px Arial";
    if (bot_status.gender === proto.botStatusRequest.gender_type.WOMAN) {
        ctx.fillStyle = "rgb(255,20,147)";
    } else {
        ctx.fillStyle = "rgb(0,191,255)";
    }
    ctx.fillText(bot_status.name, x - 8, y + 20)

}


// 绘制矩形
function drawMatrixGuest(x = canvas.width / 2, y = canvas.height / 2, e_x, e_y, name, gender) {
    // 矩形
    // ctx.fillStyle = "rgb(200,0,0)"
    // ctx.fillRect(x, y, 50, 30);

    // 眼睛
    ctx.beginPath();
    x = x + 5;
    y = y + 5;
    ctx.arc(x, y, 8, 0, 2 * Math.PI);

    if (gender === proto.botStatusRequest.gender_type.WOMAN) {
        ctx.fillStyle = "rgb(255,20,147)";
    } else {
        ctx.fillStyle = "rgb(0,191,255)";
    }

    ctx.fill();

    // 瞳孔
    ctx.beginPath()

    ctx.arc(e_x, e_y, 4, 0, 2 * Math.PI);
    ctx.fillStyle = "rgb(255,255,255)";
    ctx.fill()

    ctx.font = "12px Arial";
    if (gender === proto.botStatusRequest.gender_type.WOMAN) {
        ctx.fillStyle = "rgb(255,20,147)";
    } else {
        ctx.fillStyle = "rgb(0,191,255)";
    }
    ctx.fillText(name, x - 8, y + 20)

}

// x,y 的坐标是眼眶的坐标,理论上来讲，这个地方应该用角度来计算位置
function moveEye(x, y, r) {

    var r_pupil = 5;

    var e_x = (x - mouse_poi.x) * (r - r_pupil) / Math.sqrt(Math.pow(x - mouse_poi.x, 2) + Math.pow(y - mouse_poi.y, 2));
    var e_y = (y - mouse_poi.y) * (r - r_pupil) / Math.sqrt(Math.pow(x - mouse_poi.x, 2) + Math.pow(y - mouse_poi.y, 2));


    if (move_direction.up) {
        e_x = 0;
        e_y = r - r_pupil;
    } else if (move_direction.down) {
        e_x = 0;
        e_y = -(r - r_pupil);
    }

    if (move_direction.left) {
        e_x = r - r_pupil;
        e_y = 0;
    } else if (move_direction.right) {
        e_x = -(r - r_pupil);
        e_y = 0;
    }

    var dis = Math.sqrt(Math.pow(r - r_pupil, 2) / 2)
    if (move_direction.up && move_direction.left) {
        e_x = dis;
        e_y = dis;
    } else if (move_direction.up && move_direction.right) {
        e_x = -dis;
        e_y = dis;
    } else if (move_direction.down && move_direction.left) {
        e_x = dis;
        e_y = -dis;
    } else if (move_direction.down && move_direction.right) {
        e_x = -dis;
        e_y = -dis;
    }

    return [
        x - e_x,
        y - e_y
    ];
}

// 事件
function bindEvent() {
    window.addEventListener('keydown', (evt) => {
        // 如果input框弹出，那么禁止其他键盘事件
        if (input_deny_move_key && evt.keyCode != keys.send) {
            return false;
        }

        // 缩放
        switch (evt.keyCode) {
            case keys.up:
                move_direction.up = true;
                move_direction.down = false;
                break;
            case keys.down:
                move_direction.up = false;
                move_direction.down = true;
                break;
            case keys.right:
                move_direction.right = true;
                move_direction.left = false;
                break;
            case keys.left:
                move_direction.left = true;
                move_direction.right = false;
                break;
            case keys.talk:
                createInput();
                break;
            case keys.send:
                sendMessage();
                break;
        }
    })

    window.addEventListener('keyup', (evt) => {

        switch (evt.keyCode) {
            case keys.up:
                move_direction.up = false;
                break;
            case keys.down:
                move_direction.down = false;
                break;
            case keys.right:
                move_direction.right = false;
                break;
            case keys.left:
                move_direction.left = false;
                break;
            // talk pass
        }

        if (evt.keyCode == keys.up || evt.keyCode == keys.down || evt.keyCode == keys.left || evt.keyCode == keys.right) {
            random_points_update_key = true;
            random_points_update_status = 1;
        }
    })

    // 获取当前鼠标位置
    window.addEventListener("mousemove", (evt) => {
        mouse_poi = {
            x: evt.x, y: evt.y
        }
    });

    document.body.addEventListener('touchmove', (e) => {
        e.preventDefault();
    });
}


function createInput() {

    if (input != null) {
        return false;
    }

    input_deny_move_key = true;

    input = document.createElement("input");
    input.setAttribute("style", "position:fixed;" +
        "left:" + (visual.x) + "px;" +
        "top:" + (visual.y + 30) + "px;" +
        "background-color:rgba(200,200,200,0.2);" +
        "border:1px solid rgba(200,200,200,0.2);" +
        "border-radius:10px;" +
        "padding:5px;" +
        "outline:none;" +
        "width:150px;" +
        "color:white;" +
        "font-size:12px"
    );

    input.setAttribute('maxlength', 50);

    document.body.appendChild(input)
    input.addEventListener('focus', () => {
    });

    input.addEventListener('blur', () => {
        document.body.removeChild(input);
        input = null;
        input_deny_move_key = false;
    })

    input.focus()
}

// todo
function sendMessage() {
    if (!input || !input.value) {
        return false;
    }

    var value = input.value;
    input.blur();

    // 创建div
    if (show_message_box == null) {
        show_message_box = document.createElement("div");
        show_message_box.setAttribute("style", "position:fixed;" +
            "left:" + (visual.x) + "px;" +
            "bottom:" + (canvas.height - visual.y + 20) + "px;" +
            "color:white;" +
            "font-size:12px"
        );
        document.body.appendChild(show_message_box)
    }

    createMessageBubble(value)

    // 发送文字消息
    sendStatusByWs(value);
}


// 创建气泡
function createMessageBubble(value) {
    var bubble = document.createElement('p')
    bubble.innerHTML = "<span style='padding:0 5px;margin:5px 0;display:inline-block;background-color:rgba(200,200,200,0.2);border:1px solid rgba(200,200,200,0.2);border-radius:10px;'>" + value + "</span>";
    show_message_box.appendChild(bubble)

    setTimeout(() => {
        show_message_box.removeChild(bubble)
    }, 1000 * 8);
}

function createGuestBot() {
    for (var i in guest_bots) {
        if (i !== bot_status.bot_id && isShowGuest(guest_bots[i].r_x + guest_bots[i].x - real_top_left_poi.x, guest_bots[i].r_y + guest_bots[i].y - real_top_left_poi.y)) {
            drawMatrixGuest(guest_bots[i].r_x + guest_bots[i].x - real_top_left_poi.x, guest_bots[i].r_y + guest_bots[i].y - real_top_left_poi.y, guest_bots[i].r_x + guest_bots[i].e_x - real_top_left_poi.x, guest_bots[i].r_y + guest_bots[i].e_y - real_top_left_poi.y, guest_bots[i].name, guest_bots[i].gender)
            // console.info(guest_bots[i].show_message_box)
            showGuestMessage(i)
            // console.info(guest_bots[i].show_message_box)
            if (guest_bots[i].msg) {
                createMessageBubbleGuest(i, guest_bots[i].msg)
                guest_bots[i].msg = '';
            }
            moveBubbleGuest(i)
        }
    }
}

/**
 * 如果guest不在视野范围之内，那么不进行绘制，节省绘制资源
 * @param guestX
 * @param guestY
 */
function isShowGuest(guestX, guestY) {
    if (guestX < 0) {
        return false;
    }

    if (guestY < 0) {
        return false;
    }

    if (guestX > canvas.width) {
        return false;
    }

    if (guestY > canvas.height) {
        return false;
    }

    return true;
}


function showGuestMessage(name) {
    // 创建div
    if (!guest_show_message_box[name]) {

        guest_show_message_box[name] = document.createElement("div");

        guest_show_message_box[name].setAttribute("style", "position:fixed;" +
            "left:" + (guest_bots[name].x + guest_bots[name].r_x - real_top_left_poi.x) + "px;" +
            "bottom:" + (canvas.height - (guest_bots[name].y + guest_bots[name].r_y - real_top_left_poi.y) + 20) + "px;" +
            "color:white;" +
            "font-size:12px"
        );
        document.body.appendChild(guest_show_message_box[name])
    }
}

// 创建气泡
function createMessageBubbleGuest(name, value) {
    let bot = guest_bots[name];
    let show_message_box = guest_show_message_box[name];

    var bubble = document.createElement('p')
    bubble.innerHTML = "<span style='padding:0 5px;margin:5px 0;display:inline-block;background-color:rgba(200,200,200,0.2);border:1px solid rgba(200,200,200,0.2);border-radius:10px;'>" + value + "</span>";
    show_message_box.appendChild(bubble)

    setTimeout(() => {
        show_message_box.removeChild(bubble)
    }, 1000 * 15);
}

function moveBubbleGuest(name) {
    let bot = guest_bots[name];
    let show_message_box = guest_show_message_box[name];

    show_message_box.setAttribute("style", "position:fixed;" +
        "left:" + (bot.x + bot.r_x - real_top_left_poi.x) + "px;" +
        "bottom:" + (canvas.height - (bot.y + bot.r_y - real_top_left_poi.y) + 20) + "px;" +
        "color:white;" +
        "font-size:12px"
    );
}

function moveBubble() {
    if (show_message_box != null) {
        show_message_box.setAttribute("style", "position:fixed;" +
            "left:" + (visual.x) + "px;" +
            "bottom:" + (canvas.height - visual.y + 20) + "px;" +
            "color:white;" +
            "font-size:12px"
        );
    }
}

// 核心移动
function matrixMove() {

    var poi_y = matrix_poi.y;
    var poi_x = matrix_poi.x;
    var x_speed = 0;
    var y_speed = 0;

    if (move_direction.up) {
        poi_y = matrix_poi.y - move_speed;
        y_speed = -move_speed;
    } else if (move_direction.down) {
        poi_y = matrix_poi.y + move_speed;
        y_speed = move_speed;
    }

    if (move_direction.left) {
        poi_x = matrix_poi.x - move_speed;
        x_speed = -move_speed;
    } else if (move_direction.right) {
        poi_x = matrix_poi.x + move_speed;
        x_speed = move_speed;
    }

    if (x_speed || y_speed) {
        zoom('far');
        // 设定martix的移动边界，为半径
        var moveRaidus = 100;

        // 判断如果移动距离超过了canvas的中心moveRadius，那么停止移动，下一步进行star移动
        // 1. 移动star
        if (computeDistance(poi_x, poi_y, canvas.width / 2, canvas.height / 2) >= moveRaidus) {
            // 关闭随机移动
            random_points_update_key = false;
            random_points_update_status = 0;
            random_points_update = [
                -x_speed,
                -y_speed,
                0
            ];

            real_top_left_poi.x += x_speed;
            real_top_left_poi.y += y_speed;

            bot_status.r_x = real_top_left_poi.x;
            bot_status.r_y = real_top_left_poi.y;

        } else {
            // 2. 移动矩形
            matrix_poi.y = poi_y;
            matrix_poi.x = poi_x;
        }
    } else {
        zoom('near');
    }

    // 视点跟随
    visual.x = matrix_poi.x;
    visual.y = matrix_poi.y;

    // 气泡跟随
    moveBubble();

    drawMatrix(matrix_poi.x, matrix_poi.y);
}

// 视角缩放 far near
function zoom(direction = 'far') {
    // 步进灵敏度
    var acc = 0.1;
    // 平滑灵敏度
    var pacc = 5;
    // 变化范围
    var scope = 1;

    if (direction == 'far') {
        if (zoom_scope_value < Math.PI) {
            zoom_scope_value += acc;
            visual.z = visual.z + scope * Math.sin(zoom_scope_value) / pacc
        }
    } else if (direction == 'near') {
        if (zoom_scope_value > 0) {
            zoom_scope_value -= acc;
            visual.z = visual.z - scope * Math.sin(zoom_scope_value) / pacc
        }
    }
}

// 计算两点的距离
function computeDistance(x, y, x1, y1) {
    return Math.sqrt(Math.pow(x - x1, 2) + Math.pow(y - y1, 2))
}

function createWebSocket() {
    ws = new WebSocket("ws://" + location.hostname + ":9000/ws")

    ws.binaryType = 'arraybuffer';

    ws.onopen = function () {
        console.info("ws open")
        is_ws_open = true;
    };

    ws.onmessage = function (evt) {
        var r = proto.botStatusResponse.deserializeBinary(evt.data)
        var bot_list = r.getBotStatusList();

        for (var i in bot_list) {
            // 如果收到广播连接断开，那么删除元素
            if (bot_list[i].getStatus() === proto.botStatusRequest.status_type.CLOSE) {
                delete guest_bots[bot_list[i].getBotId()];
                continue;
            }

            guest_bots[bot_list[i].getBotId()] = {
                x: bot_list[i].getX(),
                y: bot_list[i].getY(),
                e_x: bot_list[i].getEyeX(),
                e_y: bot_list[i].getEyeY(),
                r_x: bot_list[i].getRealX(),
                r_y: bot_list[i].getRealY(),
                msg: bot_list[i].getMsg(),
                name: bot_list[i].getName(),
                gender:bot_list[i].getGender(),
                // show_message_box: !!guest_bots[i].show_message_box ? guest_bots[i].show_message_box:undefined
            };
        }
    };

    ws.onclose = function () {
        console.info("ws close")
    }
}


function sendStatusByWs(msg = '') {

    if (!is_ws_open) {
        return false;
    }

    var is_open = false;

    if (!!bot_status_old) {
        for (var i in bot_status) {
            if (bot_status[i] !== bot_status_old[i]) {
                is_open = true;
            }
        }
    } else {
        is_open = true;
    }

    if (is_open || msg) {

        let chat = new proto.botStatusRequest();

        chat.setBotId(bot_status.bot_id);
        // console.info(visual.x,visual.y,bot_status.e_x,bot_status.e_y)
        chat.setX(visual.x);
        chat.setY(visual.y);
        chat.setEyeX(bot_status.e_x);
        chat.setEyeY(bot_status.e_y);
        chat.setRealX(real_top_left_poi.x);
        chat.setRealY(real_top_left_poi.y);
        chat.setMsg(msg);
        chat.setName(bot_status.name);
        chat.setGender(bot_status.gender);

        ws.send(chat.serializeBinary());

        Object.assign(bot_status_old, bot_status)
    }
}


function initLocalStorage() {
    var name = localStorage.getItem('star_name');
    if (name !== null && name !== "") {
        bot_status.name = name;
    } else {
        bot_status.name = 'guest' + Math.random().toString(36)
    }

    var gender = localStorage.getItem('star_gender');
    if (gender !== null) {
        bot_status.gender = parseInt(gender);
    } else {
        bot_status.gender = proto.botStatusRequest.gender_type.MAN
    }
}

function initTools() {
    var tool_box = document.createElement('div');
    tool_box.setAttribute("style", "" +
        "position:fixed;" +
        "text-align:center;" +
        "left:5px;" +
        "top:50px;" +
        "width:30px;" +
        "height:200px;" +
        "background-color:rgba(0,0,0,0.5);" +
        "border:1px solid rgba(0,0,0,0.5);" +
        "border-radius:5px;");
    document.body.appendChild(tool_box);

    let name = createBtn(tool_box, 'image/human.png', '点我修改昵称');

    var input = null;

    name.addEventListener('click', (evt) => {

        if (input) {
            return false;
        }

        input = document.createElement('input');

        input.setAttribute("style", "position:fixed;" +
            "left:50px;" +
            "top:50px;" +
            "background-color:white;" +
            "border:1px solid white;" +
            "border-radius:5px;" +
            "padding:5px;" +
            "outline:none;" +
            "width:150px;" +
            "font-size:12px"
        );

        input.setAttribute('placeholder', '请输入昵称，长度10')
        input.setAttribute('maxlength', 10);

        document.body.appendChild(input)
        input.focus();

        input.addEventListener('blur', () => {
            // 设置名称
            if (input.value !== "") {
                bot_status.name = input.value;
                localStorage.setItem('star_name', bot_status.name);
            }
            // 移除节点
            document.body.removeChild(input)
            input = null;
        })
    })

    let genderMan = createBtn(tool_box, 'image/m.png', '男生');

    genderMan.addEventListener('click', (evt) => {
        bot_status.gender = proto.botStatusRequest.gender_type.MAN
        localStorage.setItem('star_gender', bot_status.gender);
    });

    let genderWoman = createBtn(tool_box, 'image/w.png', '女生');
    genderWoman.addEventListener('click', (evt) => {
        bot_status.gender = proto.botStatusRequest.gender_type.WOMAN
        localStorage.setItem('star_gender', bot_status.gender);
    });


}

function createBtn(tool_box, src, title = '') {
    var button = document.createElement("img")
    button.setAttribute("style", "" +
        // "display:inline-block;" +
        "width:25px;" +
        "height:25px;" +
        // "background-color:rgba(200,200,0,0.5);" +
        "border:1px solid rgba(200,200,0,0.5);" +
        "color:white;" +
        "cursor:default;" +
        "border-radius:5px;");

    button.setAttribute('src', src);
    button.setAttribute('title', title);

    tool_box.appendChild(button);

    return button;
}

function createReadme() {
    var readme = document.createElement("div")
    readme.setAttribute("style", "" +
        "position:fixed;" +
        "left:5px;" +
        "bottom:0px;" +
        "width:500px;" +
        "height:50;" +
        // "background-color:rgba(200,200,0,0.5);" +
        // "border:1px solid rgba(200,200,0,0.5);" +
        "color:rgba(200,200,200,0.8);" +
        "cursor:default;" +
        "border-radius:5px;");
    readme.innerHTML = "" +
        "<p>欢迎进入游戏</p>" +
        "<p>概念来自EVE游戏，以及蝌蚪聊天室，不过该游戏代码都是全新实现的</p>" +
        "<p>操作方式：</p>" +
        "<p>1. W A S D进行上下左右</p>" +
        "<p>2. 空格开启聊天框，回车发送消息</p>" +
        "<p>3. 左上角修改昵称，点击空白修改成功</p>" +
        "<p>作者GIT：<a href='https://github.com/sunshinev/go-space-chat' style='color:rgba(200,200,200,0.8)'>https://github.com/sunshinev/go-space-chat</a></p>" +
        "<p>前端 Vue+canvas+websocket+protobuf</p>" +
        "<p>后端 Golang+websocket+protobuf+goroutine</p>";

    document.body.appendChild(readme)
}


function createDirectionSign() {
    // 根据两点

}

export default function () {
    initCtx();
    bindEvent();
    initTools();
    initLocalStorage();
    createWebSocket();
    createReadme();
    window.requestAnimationFrame(canvasHandle);
};