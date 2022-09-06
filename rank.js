import m from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
import {sleep} from 'k6';

const epDataSent = new Counter('endpoint_data_sent');
const epDataRecv = new Counter('endpoint_data_recv');
const errCounter = new Counter('errCounter');
const sendCounter = new Counter('sendCounter');
const recCounter = new Counter('recCounter');

let addr = '172.24.140.131:12345';
addr = '127.0.0.1:12345';
// addr = '10.0.0.3:12345';

export const options = {
    // vus: 400,
    // duration: '300s',
    scenarios: {
        // 场景-登录移动
        scenario: {
            executor: 'shared-iterations',
            vus: 1,
            iterations: 1,
            gracefulStop: '3s',
        }
    }
};
let apiJson = JSON.parse(open('./config/apiData.json'));

export function setup() {
    // 全局初始化
    console.log('setup');
}

let i = 0;
export default function () {

    function onRec(msg) {
        console.log('onRec:', msg);
        recCounter.add(1);
        epDataRecv.add(JSON.stringify(msg).length);
    }

    const move_times = 10;

    let getInvoKeApiJson = function (name, msg) {
        const reqJson = apiJson[name];
        if (name != "event") {
            // 不需要response的不加id
            reqJson.id = ++i;
        } else {
            reqJson.id = 0;
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
        // try {
        let res = m.send(reqJson);
        console.log(new Date(), ':resJson:', res);
        sendCounter.add(1);
        return res;
        // } catch (e) {
        //     console.log('[js] send fail: ', e);
        //     errCounter.add(1);
        // }
    }

    sleep(1);
    m.connect(addr);
    // m.connectOnRec(addr, onRec);
    const account_id = "" + __VU;
    const loginInfo = invokeApi("login", {account_id, 'account_token': '123456'});
    if (!loginInfo || !loginInfo.result) {
        console.log('[js]login fail by loginInfo', __VU, loginInfo);
        // m.close();
        return;
    }
    let uid = loginInfo.msg.uid;

    if (!uid || typeof (uid) != 'number') {
        console.log('[js]login fail by uid', __VU, uid);
        // m.close();
        return;
    }
    // invokeApi("info");
    const single_score = __VU*10
    const total_score = __VU*100
    invokeApi("rankUpdate", {single_score, total_score});
    invokeApi("rankInfo");
    console.log('login success:', __VU, uid);
    // invokeApi("event");
    // m.close();
    // sleep(1);
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}