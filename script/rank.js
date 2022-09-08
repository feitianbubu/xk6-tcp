import client from './client.js'

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
    // invokeApi("info");
    const single_score = __VU*10
    const total_score = __VU*100
    client.invokeApi("rankUpdate", {single_score, total_score});
    client.invokeApi("rankInfo");
    console.log('login success:', __VU, uid);
    // invokeApi("event");
    // m.close();
    // sleep(1);
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}