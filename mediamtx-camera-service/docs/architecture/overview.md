# Architecture Overview

## System Design

The MediaMTX Camera Service is built as a lightweight wrapper around MediaMTX that provides:

1. **Real-time USB camera discovery and monitoring**
2. **WebSocket JSON-RPC 2.0 API** for client applications
3. **Dynamic MediaMTX configuration management**
4. **Streaming, recording, and snapshot coordination**

## Component Architecture

`
┌────
                    Client Applications                      
            (Web browsers, mobile apps, etc.)               
─
                       WebSocket JSON-RPC 2.0
                      
─
                Camera Service (Python)                     
  ─    
            WebSocket JSON-RPC Server                      
     Client connection management                         
     JSON-RPC 2.0 protocol handling                      
     Real-time notifications                             
      
      
             Camera Discovery Monitor                      
     USB connect/disconnect detection                    
     v4l2 capability detection                           
     Camera status tracking                              
      
      
              MediaMTX Controller                          
     REST API client                                     
     Dynamic stream management                           
     Recording coordination                              
      

                      HTTP REST API

                   MediaMTX Server (Go)                      
      
                Media Processing                           
     RTSP/WebRTC/HLS streaming                           
     Hardware-accelerated encoding                       
     Multi-protocol support                              
     Recording and snapshot generation                   
      

                      FFmpeg + V4L2

                 USB Cameras                                 
         /dev/video0, /dev/video1, etc.                     

`

## Data Flow

### Camera Discovery Flow
1. **Monitor** detects USB camera connection via udev/polling
2. **Detector** probes camera capabilities using v4l2-ctl
3. **Controller** creates MediaMTX stream configuration
4. **Server** broadcasts camera status notification to clients

### Streaming Flow  
1. **Client** requests stream via JSON-RPC
2. **Controller** configures MediaMTX path with FFmpeg source
3. **MediaMTX** starts camera capture and encoding
4. **Client** accesses stream via RTSP/WebRTC/HLS URL

### Recording Flow
1. **Client** requests recording start via JSON-RPC
2. **Controller** enables recording in MediaMTX configuration
3. **MediaMTX** captures video to file with metadata
4. **Server** notifies client when recording completes

## Technology Stack

- **Camera Service**: Python 3.10+, asyncio, websockets
- **Media Server**: MediaMTX (Go binary)
- **Camera Interface**: V4L2, FFmpeg
- **Protocols**: WebSocket, JSON-RPC 2.0, REST, RTSP, WebRTC, HLS
- **Deployment**: Systemd services, native Linux
