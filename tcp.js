import client from 'k6/x/tcp';
import {Counter} from 'k6/metrics';

const errCounter = new Counter('errCounter');
const sendCounter = new Counter('sendCounter');
const recCounter = new Counter('recCounter');
const addr = '127.0.0.1:12345';

export const options = {
    vus: 1000,
    duration: '1000s',
};
export const epDataSent = new Counter('endpoint_data_sent');
export const epDataRecv = new Counter('endpoint_data_recv');

let i = 0;
const apiJson = JSON.parse(open('./config/apiData.json'));
export default function () {
    try {
        let invokeApi = function (name, msg) {
            const reqJson = apiJson[name];
            reqJson.id = ++i;
            if (msg) {
                reqJson.msg = msg;
            }
            // console.log('reqJson:', reqJson);
            epDataSent.add(JSON.stringify(reqJson).length);
            client.send(reqJson);
            sendCounter.add(1);
            const res = client.rec();
            // console.log('res:', res);
            epDataRecv.add(JSON.stringify(res).length);
            recCounter.add(1);
            return res;
        }

        client.connect(addr);
        const msg = {
            account_id: "" + __VU,
            account_token: "123456",
        };

        // let loginRes = invokeApi('login', msg);
        for (let j = 0; j < 100; j++) {
            let infoRes = invokeApi('info');
        }
    } catch (e) {
        console.log('err:', e);
        errCounter.add(1);
    } finally {
        client.close();
    }
}
