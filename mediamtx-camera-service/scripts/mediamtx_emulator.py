import asyncio
from aiohttp import web

async def handle_get(request):
    return web.json_response({"status": "ok", "version": "emulated", "config": {}})

async def handle_ok(request):
    return web.json_response({"status": "ok"})

app = web.Application()
app.add_routes([
    web.get('/v3/config/global/get', handle_get),
    web.post('/v3/config/paths/add/{name}', handle_ok),
    web.post('/v3/config/paths/delete/{name}', handle_ok),
    web.post('/v3/config/paths/edit/{name}', handle_ok),
    web.get('/v3/paths/list', handle_ok),
    web.get('/v3/paths/get/{name}', handle_ok),
    web.post('/v3/config/global/patch', handle_ok),
])

if __name__ == '__main__':
    web.run_app(app, host='127.0.0.1', port=9997)
