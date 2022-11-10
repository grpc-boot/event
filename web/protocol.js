const LevelJson = 0;
const LevelV0   = LevelJson;
const LevelV1   = 1;
const LevelV2   = 2;

//协议相关
const EventConnectSuccess = 0x0100;
const EventTick           = 0x0101;
const EventClose          = 0x0102;
const EventError          = 0x0103;

// 登录相关
const EventLogin               = 0x0200;
const EventLoginSuccess        = 0x0201;
const EventLoginFailed         = 0x0202;

const EventMessage = 0x0300;

function encrypt(msg, k, v) {
    return CryptoJS.AES.encrypt(CryptoJS.enc.Utf8.parse(msg),
        CryptoJS.enc.Utf8.parse(k), {
            iv: CryptoJS.enc.Utf8.parse(v)
        }
    );
}

function decrypt(base64Data, k, v) {
    return CryptoJS.AES.decrypt(base64Data,
        CryptoJS.enc.Utf8.parse(k), {
            iv: CryptoJS.enc.Utf8.parse(v)
        }
    );
}

function randKey() {
    return Math.ceil(0x1000000000000000 + Math.random() * 0xf000000000000000).toString(16);
}

function int2HexWithPad(val, len) {
    let data = val.toString(16);
    if (data.length >= len) {
        return data;
    }

    let diff = len - data.length;
    for(let i=0; i < diff; i++) {
        data = '0' + data;
    }
    return data;
}

function Package(id, name) {
    this.id   = id;
    this.name = name;
    this.param = {};
}

Package.prototype.setParam = function(param) {
    this.param = param;
}

Package.prototype.pack = function() {
    return JSON.stringify(this);
}

Package.prototype.unpack = function (data) {
    if (!data) {
        return false;
    }

    let obj = JSON.parse(data);
    if (!obj['id'] || !obj['name']) {
        return false;
    }

    this.id   = obj['id'];
    this.name = obj['name'];
    this.param = obj['param'];

    return true;
}

function WsProtocol(uri, level, k, v) {
    this.uri      = uri;
    this.level    = level;
    this.k        = k;
    this.v        = v;
    this.ws       = null;
    this.protocol = null;
    this.events   = {};
    this.retryInterval   = 5000;
    this.timer           = null;
}

WsProtocol.prototype.resetRetryInterval = function (interval) {
    this.retryInterval = interval;
    this.retry();
}

WsProtocol.prototype.retry = function() {
    let self = this;
    clearInterval(this.timer);
    this.timer = setInterval(function (){
        if(self.ws && self.ws.readyState < WebSocket.CLOSING) {
            return;
        }

        if(!self.dial()) {
            console.log('connect failed');
        }
    }, this.retryInterval);
}

WsProtocol.prototype.emit = function (pkg) {
    try {
        let msg = this.protocol.pack(pkg);
        this.ws.send(msg);
    }catch (e) {
        console.log(e);
    }
}

WsProtocol.prototype.buildUri = function () {
    let url = this.uri;

    if(url.indexOf('?') > 0) {
        url += '&';
    }else {
        url += '?';
    }

    url+= ('l='+this.level);
    let key = randKey();

    switch (this.level) {
        case LevelJson:
            this.protocol = new Json();
            break;
        case LevelV1:
            key += randKey();
            this.protocol = new V1(key);
            url+= ('&k='+encrypt(key, this.k, this.v).ciphertext.toString());
            break;
        case LevelV2:
            this.protocol = new V2(key, this.k + this.v);
            url+= ('&k='+encrypt(key, this.k, this.v).ciphertext.toString());
            break;
    }

    return url;
}

WsProtocol.prototype.trigger = function (eventId, data) {
    if(!(this.events[ eventId ])) {
        return;
    }

    let self = this;

    for(let i=0; i<this.events[eventId].length; i++) {
        setTimeout(function (){
            self.events[eventId][i](data);
        }, 0);
    }
}

WsProtocol.prototype.dial = function () {
    let url = this.buildUri();
    let self = this;

    this.ws = new WebSocket(url);

    this.ws.onmessage = function (event) {
        let pkg = self.protocol.unpack(event.data);
        if(!pkg) {
            console.log('unpack failed', event);
            return;
        }
        self.trigger(pkg.id, pkg);
    };

    this.ws.onclose = function(event) {
        console.log('connection close', event);
        self.trigger(EventClose, event);
    }

    this.ws.onopen = function (event) {
        console.log('connect success', event);
        self.trigger(EventConnectSuccess, event);
    }

    this.ws.onerror = function (event) {
        console.log('error', event);
        self.trigger(EventError, event);
    }

    this.retry();

    return true;
}

WsProtocol.prototype.on = function(eventName, handler) {
    if(!this.events[eventName]) {
        this.events[eventName] = [];
    }
    this.events[eventName].push(handler);
}

function Json() {
    this.level = 0;
}

Json.prototype.pack = function (pkg) {
    return pkg.pack();
}

Json.prototype.unpack = function (data) {
    let pkg = new Package();
    if(pkg.unpack(data)) {
        return pkg;
    }
    return null;
}

function V1(k) {
    this.k     = k;
    this.level = 1;
    this.headerLength = 16;
    this.startTag  = '013a';
}

V1.prototype.pack = function (pkg) {
    let data = encrypt(pkg.pack(), this.k.substring(0, 16), this.k.substring(16, 32)).toString();
    return this.startTag + int2HexWithPad(pkg.id, 4) + int2HexWithPad(data.length, 8) + data;
}

V1.prototype.unpack = function (data) {
    if (!data || data.length < this.headerLength + 1 || data.substring(0, 4) !== this.startTag){
        return false;
    }

    let id = parseInt(data.substring(4, 8), 16);
    if(id < 1 || id > 0xffff) {
        return false;
    }

    let bodyLength = parseInt(data.substring(8, this.headerLength), 16);
    if(bodyLength + this.headerLength !== data.length) {
        return false;
    }

    let res = decrypt(data.substring(this.headerLength), this.k.substring(0, 16), this.k.substring(16, 32));
    let dataStr = res.toString(CryptoJS.enc.Utf8);
    let pkg = new Package();
    if(pkg.unpack(dataStr)) {
        if(pkg.id !== id) {
            return false;
        }
        return pkg;
    }
    return false;
}

function V2(k, key) {
    this.key   = key;
    this.k     = k;
    this.level = 2;
    this.headerLength = 16;
    this.startTag  = '023a';
    this.v     = '';
}

V2.prototype.pack = function (pkg) {
    let data = encrypt(pkg.pack(), this.k, this.v).toString();
    return this.startTag + int2HexWithPad(pkg.id, 4)+ int2HexWithPad(data.length, 8) + data;
}

V2.prototype.unpack = function (data) {
    if (!data || data.length < this.headerLength + 1) {
        return false;
    }

    if (data.substring(0, 4) !== this.startTag) {
        if(this.v.length < 1 && data.substring(0, 1) === '{') {
            let pkg = new Package();
            if(pkg.unpack(data)) {
                if(pkg.id === EventConnectSuccess && pkg.param['data']) {
                    let v = decrypt(pkg.param['data'], this.key.substring(0, 16), this.key.substring(16));
                    this.v = v.toString(CryptoJS.enc.Utf8);
                }

                return pkg;
            }
        }
        return false;
    }

    let id = parseInt(data.substring(4, 8), 16);
    if(id < 1 || id > 0xffff) {
        return false;
    }

    let bodyLength = parseInt(data.substring(8, this.headerLength), 16);
    if(bodyLength + this.headerLength !== data.length) {
        return false;
    }

    let res = decrypt(data.substring(this.headerLength), this.k, this.v);
    let dataStr = res.toString(CryptoJS.enc.Utf8);
    let pkg = new Package();
    if(pkg.unpack(dataStr)) {
        if(pkg.id !== id) {
            return false;
        }
        return pkg;
    }
    return false;
}