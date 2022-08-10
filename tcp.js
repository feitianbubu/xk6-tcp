import {check, sleep} from 'k6';
import tcp from 'k6/x/tcp';
import {Counter, Rate} from 'k6/metrics';

const susRate = new Rate('susRate');
const errCounter = new Counter('errCounter');
const strAddr = '127.0.0.1:12345';
let client = new tcp.Client();

//
//
// import fs from 'fs';
// import util from 'util';
// let logPath = 'tcp.log';
// let logFile = fs.createWriteStream(logPath, {flags: 'a'});
// console.log = function () {
//     logFile.write(util.format.apply(null, arguments) + '\n')
// }


export function OnRevMsg(msg) {
    let strMsg = msg.toString();
    console.info('OnRevMsg:', strMsg);
    susRate.add(true)
}

export default function () {
    console.log('init test tcp');
    let err = client.connect(strAddr, OnRevMsg)
    if (err) {
        console.log('connect fail: ', err);
    }
    try {
        req('hello');
    } catch (e) {
        console.log('req fail: ', e);
        errCounter.add(1);
    }
    // sleep(1);
}
let req = function (strMsg) {
    console.log('req:', strMsg);
    client.writeStrLn(strMsg);
}

