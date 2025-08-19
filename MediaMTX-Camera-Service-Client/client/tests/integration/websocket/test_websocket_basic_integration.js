/**
 * WebSocket Basic Integration Test
 * 
 * Tests basic WebSocket functionality using native browser WebSocket API
 * This test requires a running MediaMTX server for integration testing
 */

function send(ws, method, id, params = undefined) {
  const req = { jsonrpc: '2.0', method, id };
  if (params) req.params = params;
  console.log(`📤 Sending ${method} (#${id})`, params ? JSON.stringify(params) : '');
  ws.send(JSON.stringify(req));
}

describe('WebSocket Basic Integration Tests', () => {
  let ws;
  let firstCamera = null;

  beforeAll(async () => {
    // Skip test if server is not available
    if (process.env.SKIP_INTEGRATION_TESTS === 'true') {
      console.log('Skipping integration tests - server not available');
      return;
    }
  });

  afterAll(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.close();
    }
  });

  test('should establish WebSocket connection and perform basic operations', async () => {
    // Skip test if server is not available
    if (process.env.SKIP_INTEGRATION_TESTS === 'true') {
      console.log('Skipping integration test - server not available');
      return;
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Connection timeout'));
      }, 20000);

      ws = new WebSocket('ws://localhost:8002/ws');

      ws.onopen = () => {
        console.log('✅ WebSocket connection established');
        send(ws, 'ping', 1);
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          console.log('📥', JSON.stringify(msg));

          if (msg.error && (msg.id === undefined || msg.id === null)) {
            throw new Error(`Unexpected error without id: ${msg.error.message}`);
          }

          switch (msg.id) {
            case 1: {
              if (msg.result !== 'pong') throw new Error('Ping failed');
              console.log('✅ ping ok');
              send(ws, 'get_camera_list', 2);
              break;
            }
            case 2: {
              const res = msg.result;
              if (!res || typeof res !== 'object') throw new Error('Invalid camera_list result');
              if (!Array.isArray(res.cameras)) throw new Error('cameras must be array');
              if (typeof res.total !== 'number' && typeof res.connected !== 'number') throw new Error('totals missing');
              console.log(`📊 cameras=${res.cameras.length} total=${res.total ?? res.cameras.length} connected=${res.connected ?? 0}`);
              if (res.cameras.length > 0) {
                const c = res.cameras[0];
                if (!c.device || !c.status || !c.streams) throw new Error('camera fields missing');
                firstCamera = c.device;
              }
              if (firstCamera) {
                send(ws, 'get_camera_status', 3, { device: firstCamera });
              } else {
                send(ws, 'list_recordings', 4, { limit: 1, offset: 0 });
              }
              break;
            }
            case 3: {
              const c = msg.result;
              if (!c || c.device !== firstCamera) throw new Error('get_camera_status mismatch');
              if (!c.streams || typeof c.fps !== 'number') throw new Error('status fields missing');
              console.log('✅ get_camera_status ok');
              send(ws, 'list_recordings', 4, { limit: 1, offset: 0 });
              break;
            }
            case 4: {
              const r = msg.result;
              const total = typeof r.total === 'number' ? r.total : (typeof r.total_count === 'number' ? r.total_count : undefined);
              if (!r || !Array.isArray(r.files) || typeof total !== 'number') throw new Error('list_recordings invalid');
              console.log(`✅ list_recordings ok (files=${r.files.length})`);
              send(ws, 'list_snapshots', 5, { limit: 1, offset: 0 });
              break;
            }
            case 5: {
              const r = msg.result;
              const total = typeof r.total === 'number' ? r.total : (typeof r.total_count === 'number' ? r.total_count : undefined);
              if (!r || !Array.isArray(r.files) || typeof total !== 'number') throw new Error('list_snapshots invalid');
              console.log(`✅ list_snapshots ok (files=${r.files.length})`);
              // Negative test: invalid device
              send(ws, 'get_camera_status', 6, { device: '/dev/invalid' });
              break;
            }
            case 6: {
              if (msg.error) {
                const code = msg.error.code;
                const acceptable = new Set([-32001, -1000, -1001]);
                if (!acceptable.has(code)) throw new Error(`Unexpected error code: ${code}`);
                console.log('✅ error handling ok for invalid device');
              } else if (msg.result) {
                const r = msg.result;
                if (r.device !== '/dev/invalid' || r.status !== 'DISCONNECTED') {
                  throw new Error('Unexpected result for invalid device');
                }
                console.log('✅ DISCONNECTED result handling ok for invalid device');
              } else {
                throw new Error('Expected error or result for invalid device');
              }
              console.log('🎉 All interface contract checks passed');
              clearTimeout(timeout);
              ws.close();
              resolve('All tests passed');
              break;
            }
            default:
              break;
          }
        } catch (err) {
          clearTimeout(timeout);
          ws.close();
          reject(err);
        }
      };

      ws.onerror = (error) => {
        clearTimeout(timeout);
        reject(error);
      };

      ws.onclose = () => {
        console.log('🔌 WebSocket closed');
      };
    });
  }, 30000); // 30 second timeout for integration test
});
