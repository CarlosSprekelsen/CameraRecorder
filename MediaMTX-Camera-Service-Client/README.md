# MediaMTX Camera Service Client

A React/TypeScript Progressive Web App (PWA) client for the MediaMTX Camera Service, providing real-time camera management, snapshot capture, and recording controls.

## MVP Scope

### Phase 1 (S1): Architecture & Scaffolding
- Initialize project with Vite, TypeScript, ESLint/Prettier.
- Configure Material-UI theme and PWA manifest.
- Implement WebSocket JSON-RPC client:
  - Connection management with auto-reconnect and error handling.
  - JSON-RPC 2.0 method calls.
- Set up Zustand stores for camera, UI, and connection state.
- Scaffold core React components:
  - `AppShell` (header, sidebar, main content).
  - `Dashboard` and `CameraDetail`.
- Define TypeScript types for camera data, RPC requests/responses, and notifications.

### Phase 2 (S2): Core Implementation
- **Dashboard**: Real-time camera grid with status cards and quick actions.
- **Camera Detail**: Controls for snapshot and recording, status display.
- **Snapshot Functionality**: Take snapshots (format/quality options) with history.
- **Recording Functionality**: Start/stop recording (duration/format options) and progress display.
- **Real-time Updates**: WebSocket notifications with polling fallback.
- **Responsive Design**: Mobile-first layout and PWA installability.

## Planning

All upcoming tasks and timelines are tracked exclusively in the roadmap document:  
- See `docs/requirements/client-roadmap.md` for detailed next steps.
