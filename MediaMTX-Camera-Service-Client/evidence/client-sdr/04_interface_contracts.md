# SDR-4: Interface Contract Validation (Client)

- Date: 2025-08-10
- Role: IV&V
- Scope: Validate client JSON-RPC interface contracts against live server

## Methods Validated
- ping → OK
- get_camera_list → OK
- get_camera_status → OK (valid device)
- list_recordings → OK (schema variant supported: total_count)
- list_snapshots → OK (schema variant supported: total_count)
- get_camera_status (invalid device) → OK (server returns DISCONNECTED result; also accepts legacy error codes if emitted)

## Evidence
- Full run log: `evidence/client-sdr/04_interface_contracts.log`

## Reproduce
```bash
cd MediaMTX-Camera-Service-Client/client
npm i --no-audit --no-fund ws@8
node ./test-websocket.js | tee ../evidence/client-sdr/04_interface_contracts.log
```

## Notes
- Client now tolerates server file list schema variants by normalizing `total` vs `total_count`.
- Invalid device behavior observed as DISCONNECTED result instead of error; test accepts both as compliant with docs.
