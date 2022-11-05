function crypt(msg) {
    let encrypted = CryptoJS.AES.encrypt(CryptoJS.enc.Utf8.parse(msg),
        CryptoJS.enc.Utf8.parse("SD#$523asz7*&^df"),
        {
            iv: CryptoJS.enc.Utf8.parse("312c45cDvd$!F~12"),
            mode: CryptoJS.mode.CBC,
            padding: CryptoJS.pad.Pkcs7
        }
    );
    return CryptoJS.enc.Hex.stringify(encrypted.ciphertext)
}

function WsProtocol() {

}

function Json() {
    this.level = 0;
}

function V1() {
    this.level = 1;
}

function V2() {
    this.level = 2;
}