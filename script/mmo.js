import client from './client.js'
import {sleep} from 'k6';

export const options = {
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

export function setup() {
    // 全局初始化
    console.log('setup');
}

export default function () {
    const loginInfo = client.login();
    let uid = loginInfo.msg.uid;

    if (!uid || typeof (uid) != 'number') {
        console.log('[js]login fail by uid', __VU, uid);
        // m.close();
        return;
    }
    console.log('login success:', __VU, uid);
    if (__VU==1){
        client.invokeApi("event");
    }
    const move_times = 10;
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
        client.invokeApi("move", {location});
        sleep(1);
    }
    client.invokeApi("leave", {uid});
    // m.close();
    // sleep(1);
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}