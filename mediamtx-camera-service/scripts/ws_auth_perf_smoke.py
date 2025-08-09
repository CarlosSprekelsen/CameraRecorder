import asyncio, json, os
import websockets

async def main():
    uri = os.environ.get('WS_URL', 'ws://localhost:8002/ws')
    async with websockets.connect(uri, ping_interval=None) as ws:
        req_id = 1
        async def call(method, params=None):
            nonlocal req_id
            msg = {"jsonrpc":"2.0","id":req_id,"method":method}
            if params is not None:
                msg["params"] = params
            await ws.send(json.dumps(msg))
            while True:
                data = json.loads(await ws.recv())
                if data.get('id') == req_id:
                    req_id += 1
                    return data
        outputs = {}
        outputs['ping'] = await call('ping')
        outputs['get_camera_list'] = await call('get_camera_list')
        outputs['start_no_token'] = await call('start_recording', {"device":"/dev/video0","duration_seconds":1})
        outputs['start_bad_token'] = await call('start_recording', {"device":"/dev/video0","duration_seconds":1, "auth_token":"invalid"})
        outputs['metrics'] = await call('get_metrics')
        print(json.dumps(outputs, indent=2))

if __name__ == '__main__':
    asyncio.run(main())
