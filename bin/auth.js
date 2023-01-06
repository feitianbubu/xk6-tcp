import client from './client.js'
import {sleep} from 'k6'

export const options = client.options;

export function setup() {
    // 全局初始化
    console.log('setup');
}

export default function () {
    const m = client.m
    const openId = '131402'
    // const res = client.connect(); // 登录用户
    const res = client.login(openId); // 登录用户
    let account_id = res.accountId;
    // client.callApi("changePwd", {account_id}); // 修改密码
    // client.callApi("resetPwd", {account_id:openId});  // 重置密码
    // client.callApi("enter");
    // sleep(1);
    // client.callApi("sceneEnter");
    // sleep(3);
    // client.callApi("unReg",{account_id});
    // client.callApi("serverInfo",{account_id});
    // client.callApi("userAction",{account_id});
    client.callApi("airdrop", {account_id});
    // m.close();
}

export function teardown() {
    // 全局反初始化
    console.log('teardown');
}