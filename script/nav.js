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
   client.login();
    for (let j = 0; j < 1000; j++) {
        client.invokeApi("navFind");
    }
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}