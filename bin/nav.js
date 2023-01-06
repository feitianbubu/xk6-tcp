import client from './client.js'
import {sleep} from 'k6'

export const options = client.options;

export function setup() {
    // 全局初始化
    console.log('setup');
}

export default function () {
    const openId = '131402';
    client.login(openId);
    client.callApi("enter");
    sleep(1);
    client.notifyApi("logicWatchEvents");
    client.callApi("enterScene");
    sleep(2);
    for (let j = 0; j < 3; j++) {
        client.callApi("navFind");
        sleep(300);
    }
    // client.notifyApi("leave");
    // client.m.close();
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}