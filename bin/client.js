import m from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
import {check,fail,sleep} from 'k6';

const epDataSent = new Counter('endpoint_data_sent');
const epDataRecv = new Counter('endpoint_data_recv');
const errCounter = new Counter('errCounter');
const sendCounter = new Counter('sendCounter');
const recCounter = new Counter('recCounter');

let addrs = ['172.24.140.131:12345'];
addrs = ['127.0.0.1:12345'];
// addrs = ['220.162.240.50:5816'];
addrs = ['172.24.135.32:6010'];
// addrs = ['10.35.11.162:5816'];
// addrs = ['10.0.0.3:12345', '10.0.0.3:12346'];

const options = {
    // vus: 400,
    // duration: '300s',
    stages: [
        { duration: '10s', target: 10 },
        { duration: '1m', target: 10 },
        { duration: '10s', target: 1000 },
        { duration: '1m', target: 1000 },
        { duration: '3m', target: 2000 },
        { duration: '30m', target: 3000 },
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
let startOnRec = false;
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
let invokeApi = function (name, msgJson, needRec) {
    let reqJson = getInvoKeApiJson(name, msgJson);
    // console.log(new Date(), ':reqJson:', reqJson);
    epDataSent.add(JSON.stringify(reqJson).length);
    let res;
    if (needRec) {
        if(!startOnRec){
            startOnRec = true;
            m.startOnRec();
        }
        res = m.sendAndRec(reqJson);
        console.log(new Date(), ':resJson:', res);
    }else{
        res = m.send(reqJson);
    }
    sendCounter.add(1);
    return res;
}
let notifyApi = function (name, msgJson) {
    return invokeApi(name, msgJson, false);
}
let callApi = function (name, msgJson) {
    return invokeApi(name, msgJson, true);
}

const connect = function(onRec){
    let addr = addrs[__VU%addrs.length]
    m.connectOnRec(addr,onRec);
}

const login = function(account_id){
    if(!account_id) {
        account_id = 'k6_'+ __VU + '_' + __ITER
    }
    sleep(1);
    startOnRec = false;
    connect(onRec);
    // let loginRes = invokeApi("login", {account_id});
    // if (!check(loginRes, {
    //     'logged in successfully': (resp) => (resp && resp.result),
    // })) {
    //     fail('login failed by:', loginRes);
    // }
    // return loginRes
    let res = m.login(account_id)
    console.log('login success:', account_id, res);
    return res;
}

const client = {
    m,
    login,
    notifyApi,
    callApi,
    options,
    getInvoKeApiJson,
    connect,
}
module.exports = client;