import {check, sleep} from 'k6';
import client from 'k6/x/tcp';
import {Counter, Rate} from 'k6/metrics';
// import exec from 'k6/execution';
import {apiJson} from './apiData.js';

const susRate = new Rate('susRate');
const errCounter = new Counter('errCounter');
const addr = '127.0.0.1:12345';

export const options = {
    vus: 100,
    duration: '1s',
};
export const epDataSent = new Counter('endpoint_data_sent');
export const epDataRecv = new Counter('endpoint_data_recv');

// const handler = client.connect(strAddr);

let i = 0;
export default function () {
    // console.log('exec context:', __VU);
    let {method, msg} = apiJson.config;
    const req = {
        id: ++i,
        method: method,
        body: msg,
    }
    // console.log('req:', req.id);
    // epDataSent.add(method.length + msg.length);
    let conn = client.connect(addr);
    client.send(conn, req);
    let res = client.recv(conn);
    // epDataRecv.add(JSON.stringify(res).length);
    // console.log('res:', res.id);
    // sleep(1);
}


