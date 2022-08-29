import m from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
import {sleep} from 'k6'

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
    const move_times = 100
    const opts = {
        move_times,
        account_id: __VU + 10000,
        watch_enabled: true
    }
    // client.start(addr, opts);

        let getInvoKeApiJson = function(name ,msg){
        const reqJson = apiJson[name];
        reqJson.id = ++i;
        if (msg) {
            reqJson.msg = msg;
        }
        reqJson.msg = reqJson.msg || {};
        return reqJson
    }
    let invokeApi = function (name, msg) {
        let reqJson = getInvoKeApiJson(name, msg)
        console.log(new Date(), ':reqJson:', reqJson);
        // epDataSent.add(JSON.stringify(reqJson).length);
        m.send(reqJson);
        // sendCounter.add(1);
        return reqJson.id;
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


    m.connect(addr)
    // m.init()
    // let uid = m.login(__VU)
    const account_id = ""+__VU
    invokeApi("login",{account_id, 'account_token':'123456'})
    const loginInfo = m.rec();
    if(!loginInfo.result){
        console.log('login fail', __VU, loginInfo)
        return;
    }
    const {uid} = m.parse(loginInfo.msg);
    console.log('login success:',__VU, uid);
    m.startOnRec(onRec)
    invokeApi("event")
    for (let i = 0; i < move_times; i++) {
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
        invokeApi("move", {location})
        sleep(1)
    }
    // m.send(m.GetReqObject("leave", {uid}))
    sleep(1)
}

export function onRec(msg) {
    console.log('onRec:', msg);
}

export function teardown() {
    console.log('teardown');
}