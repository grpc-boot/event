function crypt(msg, k, v) {
    k = !k ? "SD#$523asz7*&^df" : k;
    v = !v ? "312c45cDvd$!F~12" : v;

    return CryptoJS.AES.encrypt(CryptoJS.enc.Utf8.parse(msg),
        CryptoJS.enc.Utf8.parse(k), {
            iv: CryptoJS.enc.Utf8.parse(v)
        }
    );
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

function bytes2String(bytes) {
    return String.fromCharCode.apply(String, bytes)
}

function packInt(val) {
    let targets =[];
    targets[0] = val & 0xFF;
    targets[1] = val >> 8 & 0xFF;
    targets[2] = val >> 16 & 0xFF;
    targets[3] = val >> 24 & 0xFF;
    return targets;
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

function WsProtocol(level, k, v) {
    this.level = level;
    this.k     = k;
    this.v     = v;
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
    let data = crypt(pkg.pack(), this.k.substring(0, 16), this.k.substring(16, 32));
    let sign = CryptoJS.MD5(data.ciphertext.words);
    let signVal = parseInt(sign.toString().substring(0, 8), 16);
    let bytes = bytes2String([1, 58])
        + int2HexWithPad(pkg.id, 4)
        + bytes2String([0])
        + bytes2String(packInt(signVal))
        + bytes2String(data.ciphertext.words);
    return bytes;
}

function V2(k) {
    this.k     = k;
    this.level = 2;
}