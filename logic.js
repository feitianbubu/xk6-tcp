import {Counter, Rate} from 'k6/metrics';
import client from 'k6/x/tcp';
import { randomSeed, sleep } from 'k6';

const susRate = new Rate('susRate');
const strAddr = '127.0.0.1:12345';
export const options = {
    vus: 200,
    duration: '3s',
};

export const epDataSent = new Counter('endpoint_data_sent');
export const epDataRecv = new Counter('endpoint_data_recv');


const apijson = {
    login: {
        method: '/api/v1/login',
        msg: `{
         "userAddress": "0x279A4C36098c4e76182706511AB0346518ad6049",
         "content": "1659516267",
         "signature": "0xd5f2199e375be586e78e454595047edfe6688989a674eca1dc847658aa387e041ceed79ca17be21bd6c6541822ea87e1bf1f3a73cbdf3c0d76bc176a523584b600"
        }`
    },
    config: {
        method: '/api/v1/config',
        msg: `{}`
    }
}

randomSeed(123456789);

export default function () {
    let {method, msg} = apijson.config;
    req(method, msg);
}
let i = 0;
let req = function (method, strMsg) {
    epDataSent.add(method.length + strMsg.length);
    let res = client.sendDebug(strAddr, method, strMsg);
    epDataRecv.add(JSON.stringify(res).length);
    sleep(1);
}

