import client from './client.js'
import {sleep} from 'k6'
import { randomSeed } from 'k6';
randomSeed(123456789);

export const options = client.options;

export function setup() {
    // 全局初始化
    console.log('setup');
}

export default function () {
    const openId = '131402'
    const res = client.login(openId);
    const uid = res.uid;
    client.callApi("enter");
    sleep(1);
    // client.notifyApi("logicWatchEvents");
    client.callApi("enterScene");
    sleep(2);
    client.callApi("info", {userId:uid});
    const move_times = 10000;
    const msgJson = client.getInvoKeApiJson("move").msg;
    for (let j = 0; j < move_times; j++) {
        let randomX = Math.random()*20 -10;
        let randomZ = Math.random()*20 -10;
        let location = msgJson.location
        let target_location = msgJson.target_location
        target_location.x = location.x+ randomX;
        target_location.z = location.z +randomZ;
        client.notifyApi("move", msgJson);
        sleep(1);
    }
    // client.notifyApi("leave", {uid});
    // client.notifyApi("logout");
}
function onRec(msg) {
    console.log('onRec:', msg);
    recCounter.add(1);
    epDataRecv.add(JSON.stringify(msg).length);
}
export function teardown() {
    // 全局反初始化
    console.log('teardown');
}