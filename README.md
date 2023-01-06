# xk6-tcp

This is a [k6](https://go.k6.io/k6) extension using the [xk6](https://github.com/grafana/xk6) system.

| :exclamation: This is a proof of concept, isn't supported by the k6 team, and may break in the future. USE AT YOUR OWN RISK! |
|------|

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Then:

1. Install `xk6`:
  ```shell
  $ go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
  ```shell
  $ xk6 build --with github.com/feitianbubu/xk6-tcp@latest
  ```
3. Use custom version and path Build the binary:
  ```shell
  $ xk6 build v0.39.0 --with github.com/feitianbubu/xk6-tcp="../mypay/xk6-tcp"   
  ```

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
    const move_times = 100
    const opts = {
        move_times,
        account_id:__VU+10000,
        watch_enabled: true
    }
    client.start(addr, opts);
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
