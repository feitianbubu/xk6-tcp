import m from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
import {sleep} from 'k6';

const epDataSent = new Counter('endpoint_data_sent');
const epDataRecv = new Counter('endpoint_data_recv');
const errCounter = new Counter('errCounter');
const sendCounter = new Counter('sendCounter');
const recCounter = new Counter('recCounter');

let addr = '172.24.140.131:12345';
// addr = '10.0.0.3:12345';
// addr = '127.0.0.1:12345';

export const options = {
    // vus: 400,
    // duration: '300s',
    scenarios: {
        example_scenario: {
            executor: 'shared-iterations',
            vus: 1,
            iterations: 1,
            gracefulStop: '3s',
        }
    }
};
let apiJson = JSON.parse(open('./config/apiData.json'));

export function setup() {
    console.log('setup');
}

let i = 0;
export default function () {

    function onRec(msg) {
        console.log('onRec:', msg);
        epDataRecv.add(JSON.stringify(msg).length);
    }

    const move_times = 10;
    // const opts = {
    //     move_times,
    //     account_id: __VU + 10000,
    //     watch_enabled: true
    // }
    // client.start(addr, opts);

    let getInvoKeApiJson = function (name, msg) {
        const reqJson = apiJson[name];
        reqJson.id = ++i;
        if (msg) {
            reqJson.msg = msg;
        }
        reqJson.msg = reqJson.msg || {};
        return reqJson;
    }
    let invokeApi = function (name, msg) {
        let reqJson = getInvoKeApiJson(name, msg);
        console.log(new Date(), ':reqJson:', reqJson);
        epDataSent.add(JSON.stringify(reqJson).length);
        try{
            m.send(reqJson);
        }catch (e){
            console.log('[js] send fail: ', e)
        }
        // sendCounter.add(1);
        // return reqJson.id;
        // if (name === "event"){
        //     return null;
        // }
        // const res = client.getResChan();
        // // console.log('res111:', client.toString(res.method), client.toString(res.msg));
        // // console.log('res222:', client.parse(res.msg));
        // epDataRecv.add(JSON.stringify(res).length);
        // recCounter.add(1);
        // return res;
    }

    sleep(1);
    m.connect(addr);
    // m.init();
    // let uid = m.login(__VU);
    const account_id = "" + __VU;
    invokeApi("login", {account_id, 'account_token': '123456'});
    const loginInfo = m.rec();
    if (!loginInfo || !loginInfo.result) {
        console.log('[js]login fail by loginInfo', __VU, loginInfo);
        // m.close();
        return;
    }
    let uid;
    try {
        uid = m.parse(loginInfo.msg).uid;
    } catch (e) {
        console.log("[js]parse fail by parse", e, loginInfo);
        // m.close();
        return;
    }
    if (!uid || typeof (uid) != 'number') {
        console.log('[js]login fail by uid', __VU, uid);
        // m.close();
        return;
    }
    console.log('login success:', __VU, uid, typeof (uid));
    m.startOnRec(onRec);
    invokeApi("event");
    for (let j = 0; j < move_times; j++) {
        let location = {
            uid,
            "x": 1,
            "y": 1
        }
        // let msg = {
        //     "id": i++,
        //     "method": "/tevat.example.scene.Scene/Move",
        //     "msg": {location}
        // }
        invokeApi("move", {location});
        sleep(1);
    }
    // m.send(m.GetReqObject("leave", {uid}))
    // m.close();
    sleep(2);
}

export function teardown() {
    console.log('teardown');
}