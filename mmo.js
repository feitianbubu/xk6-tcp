import client from 'k6/x/tcp';
import {Counter} from 'k6/metrics';
let addr = '172.24.140.131:12345';
// addr = '10.0.0.3:12345';
addr = '127.0.0.1:12345';

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