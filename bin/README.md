# xk6-tcp

## 使用方法
  * 解压放到项目根目录下(与proto目录同级)
  * sh命令行下执行: ./k6.exe run mmo.js -u 1 -i 1
  * 更多使用命令,参考k6官方文档:https://k6.io/docs/

## changelog
### v1.1.0
  * 支持proto协议, 支持动态解析proto文件, proto路径在config/config.json中配置
### v1.0.0
  * 支持json协议
  * 支持接口api配置[config/apiData.json]
  * 脚本逻辑参考下方example,或根目录下的*.js

## Example

```javascript
import client from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
let addr = '127.0.0.1:12345';

export const options = {
    vus: 400,
    duration: '300s',
    // scenarios: {
    //     example_scenario: {
    //         executor: 'shared-iterations',
    //         vus: 1,
    //         iterations: 5,
    //         gracefulStop: '3s',
    //     }
    // }
};

export default function () {
  const m = client.m
  client.m.setLocalLogin(true); // 是否本地登录开关
  m.setProto(2);                // 服务端协议1:proto 2:json
  const openId = '126126'
  const res = client.login(openId); // 登录用户
  const account_id = res.account_id;
  client.callApi("changePwd", {account_id}); // 修改密码
  client.callApi("resetPwd", {account_id});  // 重置密码
}

```

Result output:

```
$ ./k6.exe run mmo.js -u 100 -d 100s

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: mmo.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 31s max duration (incl. graceful stop):
           * default: 100 looping VUs for 100s (gracefulStop: 30s)


INFO[0000] bar                                           source=console
INFO[0000] PONG!                                         source=console

running (00m00.0s), 0/100 VUs, 100 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  00m00.0s/10m0s  100/100 iters, 100 per VU

    █ setup

    █ teardown

    data_received........: 0 B 0 B/s
    data_sent............: 0 B 0 B/s
    iteration_duration...: avg=544.06µs min=428.6µs med=597.41µs max=606.18µs p(90)=604.43µs p(95)=605.31µs
    iterations...........: 4600   4600.10603/s
```