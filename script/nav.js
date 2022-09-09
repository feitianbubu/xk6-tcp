import client from './client.js'

export const options = client.options;

export function setup() {
    // 全局初始化
    console.log('setup');
}

export default function () {
   client.login();
    client.invokeApi("enter");
    for (let j = 0; j < 1000; j++) {
        client.invokeApi("navFind");
    }
    client.invokeApi("leave");
    client.m.close();
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}