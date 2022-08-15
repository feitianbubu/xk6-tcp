export const apiJson = {
    login: {
        method: 'tevat.example.auth.Auth/login',
        msg: {"account_id":"1","account_token":"123456"}
    },
    loginLP: {
        method: '/api/v1/login',
        msg: {
         "userAddress": "0x279A4C36098c4e76182706511AB0346518ad6049",
         "content": "1659516267",
         "signature": "0xd5f2199e375be586e78e454595047edfe6688989a674eca1dc847658aa387e041ceed79ca17be21bd6c6541822ea87e1bf1f3a73cbdf3c0d76bc176a523584b600"
        }
    },
    info: {
        method: 'tevat.example.logic.Logic/Info',
        msg: {},
    },
    config: {
        method: '/api/v1/config',
        msg: {}
    }
};