import http from 'k6/http';
const baseUrl = 'http://127.0.0.1:3080'
export const options = {
    vus: 200,
    duration: '10s',
};
export default function () {
    const payload = JSON.stringify({});

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    let apiUrl = baseUrl + '/api/v1/config';
    http.post(apiUrl, payload, params);
}