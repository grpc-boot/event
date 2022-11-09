const LevelJson = 0;
const LevelV1   = 1;
const LevelV2   = 2;

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

function intToBytesBigEndian(number, length){
    var bytes = [];
    var i = length;
    do {
        bytes[--i] = number & (255);
        number = number>>8;
    } while (i)
    return bytes;
}

function Package(id, name) {
    this.id   = id;
    this.name = name;
    this.data = {};
}

Package.prototype.setData = function(data) {
    this.data = data;
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
    this.data = obj['data'];

    return true;
}

function WsProtocol(uri, level, k, v) {
    this.uri   = uri;
    this.level = level;
    this.k     = k;
    this.v     = v;
    this.ws    = null;
    this.protocol = null;
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
            this.protocol = new V2(key);
            url+= ('&k='+encrypt(key, this.k, this.v).ciphertext.toString());
            break;
    }

    return url;
}

WsProtocol.prototype.dial = function () {
    let url = this.buildUri();
    console.log(url);
    try {
        this.ws = new WebSocket(url);
        return true;
    } catch (e) {
        return false;
    }
}

WsProtocol.prototype.accept = function () {

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
}

V1.prototype.pack = function (pkg) {
    let data = encrypt(pkg.pack(), this.k.substring(0, 16), this.k.substring(16, 32)).toString();
    let signVal = parseInt(CryptoJS.MD5(data).toString().substring(0, 8), 16);
    let bytes = String.fromCharCode.apply(String, [1, 58])
        + int2HexWithPad(pkg.id, 4)
        + String.fromCharCode.apply(String, intToBytesBigEndian(signVal, 4))
        + data;
    return bytes;
}

V1.prototype.unpack = function (data) {
    if (!data || data.length < 11 || data[0] !== String.fromCharCode(1) || data[1] !== ':'){
        return false;
    }

    let id = parseInt(data.substring(2, 6), 16);
    if(id < 1 || id > 0xffff) {
        return false;
    }

    let signVal = parseInt(CryptoJS.MD5(data.substring(10)).toString().substring(0, 8), 16);
    if(data.substring(6, 10) !== String.fromCharCode.apply(String, intToBytesBigEndian(signVal, 4))) {
        return false;
    }

    let res = decrypt(data.substring(10), this.k.substring(0, 16), this.k.substring(16, 32));
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

function V2(k) {
    this.k     = k;
    this.level = 2;
}

V2.prototype.setV = function(v) {
    this.v = v;
}

V2.prototype.pack = function (pkg) {
    let data = encrypt(pkg.pack(), this.k, this.v).toString();
    let signVal = parseInt(CryptoJS.MD5(data).toString().substring(0, 8), 16);
    let bytes = String.fromCharCode.apply(String, [1, 58])
        + int2HexWithPad(pkg.id, 4)
        + String.fromCharCode.apply(String, intToBytesBigEndian(signVal, 4))
        + data;
    return bytes;
}

V2.prototype.unpack = function (data) {
    if (!data || data.length < 11 || data[0] !== String.fromCharCode(1) || data[1] !== ':'){
        return false;
    }

    let id = parseInt(data.substring(2, 6), 16);
    if(id < 1 || id > 0xffff) {
        return false;
    }

    let signVal = parseInt(CryptoJS.MD5(data.substring(10)).toString().substring(0, 8), 16);
    if(data.substring(6, 10) !== String.fromCharCode.apply(String, intToBytesBigEndian(signVal, 4))) {
        return false;
    }

    let res = decrypt(data.substring(10), this.k, this.v);
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