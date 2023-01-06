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
    const openId = '';
    const res = client.login(openId);
    const uid = res.uid;
    const single_score = __VU*10;
    const total_score = __VU*100;
    // client.callApi("rankUpdate", {single_score, total_score});
    const msgJson = {"battle_data" : [{userId:uid, damage_score:single_score,kill_score:total_score}]}
    client.callApi("battleRankUpdate", msgJson);
    client.callApi("rankInfo");
    client.callApi("userRankInfo", {uid});
    // invokeApi("event");
    // m.close();
    // sleep(1);
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}