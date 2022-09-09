import m from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
import {check,fail,sleep} from 'k6';

const epDataSent = new Counter('endpoint_data_sent');
const epDataRecv = new Counter('endpoint_data_recv');
const errCounter = new Counter('errCounter');
const sendCounter = new Counter('sendCounter');
const recCounter = new Counter('recCounter');

let addr = '172.24.140.131:12345';
// addr = '127.0.0.1:12345';
// addr = '10.0.0.3:12345';

const options = {
    // vus: 400,
    // duration: '300s',
    stages: [
        { duration: '10s', target: 10 },
        { duration: '10m', target: 10 },
        { duration: '10s', target: 100 },
        { duration: '10m', target: 100 },
        { duration: '10s', target: 500 },
        { duration: '10m', target: 500 },
        { duration: '10s', target: 10 },
        { duration: '10m', target: 10 },
    ],
    // scenarios: {
    //     // 场景-登录移动
    //     scenario: {
    //         executor: 'shared-iterations',
    //         vus: 1,
    //         iterations: 1,
    //         gracefulStop: '3s',
    //     }
    // }
};

let apiJson = JSON.parse(open('./config/apiData.json'));
let id = 0;
function onRec(msg) {
    console.log('onRec:', msg);
    recCounter.add(1);
    epDataRecv.add(JSON.stringify(msg).length);
}
let getInvoKeApiJson = function (name, msg) {
    const reqJson = apiJson[name];
    if (name !== "event") {
        // 不需要response的不加id
        reqJson.id = ++id;
    }
    reqJson.msg = reqJson.msg || {};
    if (msg) {
        for(const k in msg){
            reqJson.msg[k] = msg[k]
        }
    }

    return reqJson;
}
let invokeApi = function (name, msgJson) {
    let reqJson = getInvoKeApiJson(name, msgJson);
    console.log(new Date(), ':reqJson:', reqJson);
    epDataSent.add(JSON.stringify(reqJson).length);
    let res = m.send(reqJson);
    console.log(new Date(), ':resJson:', res);
    sendCounter.add(1);
    return res;
}
const login = function(){
    sleep(1);
    m.connect(addr);
    const account_id = "" + __VU;
    let loginRes = invokeApi("login", {account_id, 'account_token': '123456'});
    if (!check(loginRes, {
        'logged in successfully': (resp) => (resp && resp.result),
    })) {
        fail('login failed by:', loginRes);
    }
    return loginRes
}

const client = {
    m,
    login,
    invokeApi,
    options,
}
module.exports = client;
