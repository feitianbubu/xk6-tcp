import grpc from 'k6/net/grpc';
import { randomSeed } from 'k6';
import { Counter } from 'k6/metrics';

export const options = {
    vus: 100,
    duration: '1s',
//  thresholds: {
//    grpc_req_duration: ['p(90)<100', 'p(95)<200', 'p(99)<500'],
//  },
}

const grpcReqs = new Counter('grpc_reqs');
const grpcReceiving = new Counter('grpc_req_receiving');

const client = new grpc.Client();
client.load(['/third_party/protobuffer', '../configserver/proto'], 'config/config.proto');
randomSeed(123456789);

export default () => {
    client.connect('127.0.0.1:9013', {
        plaintext: true,
    })

    for (let i=0;i<1;i++){
        const rnd = Math.round(Math.random()*50000)
        const serverType = 'auth';
        const data = {
            serverType,
        }
        grpcReqs.add(1);
        const response = client.invoke('tevat.configserver.config.Config/Info', data, {
            metadata: {
                uid: "" + rnd,
                "X-Forwarded-For": "111",
                "X-Request-ID": "222",
            },
        });

        console.log(JSON.stringify(JSON.parse(response.message.info), null, 4));
        grpcReceiving.add(1);
    }
    client.close();
}