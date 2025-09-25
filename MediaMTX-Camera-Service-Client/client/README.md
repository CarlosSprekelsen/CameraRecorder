# MediaMTX Camera Service Client

A React-based web client for the MediaMTX Camera Service, providing real-time camera management, monitoring, and control capabilities.

## Features

- **Authentication**: Token-based authentication with role-based access control
- **Real-time Communication**: WebSocket connection with automatic reconnection
- **Camera Management**: Discover and monitor camera devices
- **Recording Control**: Start/stop recordings with optional duration
- **File Management**: List, download, and delete recordings and snapshots
- **System Monitoring**: View server status, storage information, and health metrics

## Technology Stack

- **React 19** - UI framework
- **TypeScript** - Type safety
- **Material-UI** - Component library
- **Zustand** - State management
- **Vite** - Build tool
- **WebSocket** - Real-time communication
- **JSON-RPC 2.0** - API protocol

## Getting Started

### Prerequisites

- Node.js 20.19.0 or higher
- npm 10.8.0 or higher
- MediaMTX Camera Service running on the server

### Installation

1. Install dependencies:
```bash
npm install
```

2. Configure environment variables:
```bash
cp env.example .env.local
# Edit .env.local with your server configuration
```

3. Start the development server:
```bash
npm run dev
```

4. Open your browser and navigate to `http://localhost:3000`

### Environment Variables

- `VITE_WS_URL`: WebSocket URL for the MediaMTX Camera Service (default: `ws://localhost:8002/ws`)
- `VITE_API_BASE_URL`: API base URL for file downloads (default: `http://localhost:8002`)

## Development

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

### Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── Layout/         # Layout components (AppLayout, LoadingSpinner)
│   ├── Login/          # Login page components
│   └── About/          # About page components
├── pages/              # Page components
│   ├── Login/          # Login page
│   └── About/          # About page
├── stores/             # Zustand state stores
│   ├── auth/           # Authentication state
│   ├── connection/     # WebSocket connection state
│   └── server/         # Server information state
├── services/           # Business logic services
│   ├── websocket/      # WebSocket service
│   ├── auth/           # Authentication service
│   └── server/         # Server service
├── types/              # TypeScript type definitions
└── utils/              # Utility functions
```

## Architecture

The client follows a layered architecture pattern:

1. **Presentation Layer**: React components and pages
2. **Application Layer**: State management and business logic
3. **Service Layer**: API communication and data services
4. **Infrastructure Layer**: WebSocket transport and storage

### State Management

The application uses Zustand for state management with the following stores:

- **AuthStore**: Authentication state (token, role, permissions)
- **ConnectionStore**: WebSocket connection state
- **ServerStore**: Server information and status

### API Communication

The client communicates with the server using:

- **WebSocket**: Real-time bidirectional communication
- **JSON-RPC 2.0**: Structured message format
- **Automatic Reconnection**: Exponential backoff retry logic

## Authentication

The client supports token-based authentication with three user roles:

- **Viewer**: Read-only access to camera status and file listings
- **Operator**: Viewer permissions + camera control operations
- **Admin**: Full access to all features including system metrics

## Error Handling

The application includes comprehensive error handling:

- **Connection Errors**: Automatic reconnection with exponential backoff
- **Authentication Errors**: Token validation and session management
- **API Errors**: Structured error responses with user-friendly messages
- **Network Errors**: Graceful degradation and retry mechanisms

## Performance

The client is optimized for performance:

- **Lazy Loading**: Components loaded on demand
- **State Optimization**: Minimal re-renders with Zustand
- **Connection Management**: Efficient WebSocket usage
- **Error Boundaries**: Isolated error handling

## Security

Security features include:

- **Token Storage**: Secure session storage
- **Role-based Access**: UI elements hidden based on permissions
- **Input Validation**: Client-side validation for all inputs
- **Secure Communication**: WebSocket over TLS in production

## Browser Support

The client supports modern browsers:

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Contributing

1. Follow the established code style (ESLint + Prettier)
2. Write TypeScript with strict type checking
3. Follow the component architecture patterns
4. Update documentation for new features
5. Test thoroughly before submitting changes

## License

Proprietary - All rights reserved