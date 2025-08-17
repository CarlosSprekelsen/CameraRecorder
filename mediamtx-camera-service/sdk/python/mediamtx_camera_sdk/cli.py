"""
MediaMTX Camera Service Python SDK CLI.

This module provides a simple command-line interface for the MediaMTX Camera Service.
"""

import asyncio
import argparse
import json
import sys
from typing import List

from .client import CameraClient
from .exceptions import CameraServiceError, AuthenticationError, ConnectionError


async def main():
    """Main CLI entry point."""
    parser = argparse.ArgumentParser(
        description="MediaMTX Camera Service CLI",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token list
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token status /dev/video0
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token snapshot /dev/video0
  %(prog)s --host host localhost --port 8002 --auth-type jwt --token your_token record /dev/video0
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token stop /dev/video0
        """
    )

    # Connection arguments
    parser.add_argument('--host', default='localhost', help='Server hostname')
    parser.add_argument('--port', type=int, default=8002, help='Server port')
    parser.add_argument('--ssl', action='store_true', help='Use SSL/TLS')
    parser.add_argument('--auth-type', choices=['jwt', 'api_key'], default='jwt', help='Authentication type')
    parser.add_argument('--token', help='JWT token')
    parser.add_argument('--key', help='API key')
    parser.add_argument('--timeout', type=int, default=30, help='Request timeout in seconds')

    # Command and arguments
    parser.add_argument('command', help='Command to execute')
    parser.add_argument('device_path', nargs='?', help='Camera device path')
    parser.add_argument('--format', choices=['table', 'json'], default='table', help='Output format')

    args = parser.parse_args()

    # Validate authentication
    if args.auth_type == 'jwt' and not args.token:
        print("Error: JWT token required for jwt authentication", file=sys.stderr)
        return 1
    elif args.auth_type == 'api_key' and not args.key:
        print("Error: API key required for api_key authentication", file=sys.stderr)
        return 1

    # Create client
    client = CameraClient(
        host=args.host,
        port=args.port,
        use_ssl=args.ssl,
        auth_type=args.auth_type,
        auth_token=args.token,
        api_key=args.key
    )

    try:
        # Connect to service
        await client.connect()

        # Execute command
        command = args.command.lower()
        
        if command == 'list':
            return await cmd_list(client, args)
        elif command == 'status':
            return await cmd_status(client, args)
        elif command == 'snapshot':
            return await cmd_snapshot(client, args)
        elif command == 'record':
            return await cmd_record(client, args)
        elif command == 'stop':
            return await cmd_stop(client, args)
        elif command == 'ping':
            return await cmd_ping(client, args)
        else:
            print(f"Unknown command: {command}", file=sys.stderr)
            parser.print_help()
            return 1

    except KeyboardInterrupt:
        print("\nOperation cancelled by user", file=sys.stderr)
        return 130
    except AuthenticationError as e:
        print(f"Authentication error: {e}", file=sys.stderr)
        return 1
    except ConnectionError as e:
        print(f"Connection error: {e}", file=sys.stderr)
        return 1
    except CameraServiceError as e:
        print(f"Camera service error: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        return 1
    finally:
        await client.disconnect()


async def cmd_list(client: CameraClient, args) -> int:
    """List available cameras."""
    try:
        cameras = await client.get_camera_list()
        
        if args.format == 'json':
            print(json.dumps([{
                'device_path': camera.device_path,
                'name': camera.name,
                'status': camera.status,
                'capabilities': camera.capabilities,
                'stream_url': camera.stream_url
            } for camera in cameras], indent=2))
        else:
            if not cameras:
                print("No cameras available")
                return 0
            
            print(f"Found {len(cameras)} camera(s):")
            print()
            
            for i, camera in enumerate(cameras, 1):
                print(f"{i}. {camera.name}")
                print(f"   Device: {camera.device_path}")
                print(f"   Status: {camera.status}")
                print(f"   Capabilities: {', '.join(camera.capabilities) if camera.capabilities else 'None'}")
                if camera.stream_url:
                    print(f"   Stream: {camera.stream_url}")
                print()
        
        return 0
        
    except Exception as e:
        print(f"Error listing cameras: {e}", file=sys.stderr)
        return 1


async def cmd_status(client: CameraClient, args) -> int:
    """Get camera status."""
    if not args.device_path:
        print("Error: Device path required", file=sys.stderr)
        return 1
    
    try:
        camera = await client.get_camera_status(args.device_path)
        
        if args.format == 'json':
            print(json.dumps({
                'device_path': camera.device_path,
                'name': camera.name,
                'status': camera.status,
                'capabilities': camera.capabilities,
                'stream_url': camera.stream_url
            }, indent=2))
        else:
            print(f"Camera: {camera.name}")
            print(f"Device: {camera.device_path}")
            print(f"Status: {camera.status}")
            print(f"Capabilities: {', '.join(camera.capabilities) if camera.capabilities else 'None'}")
            if camera.stream_url:
                print(f"Stream: {camera.stream_url}")
        
        return 0
        
    except Exception as e:
        print(f"Error getting camera status: {e}", file=sys.stderr)
        return 1


async def cmd_snapshot(client: CameraClient, args) -> int:
    """Take camera snapshot."""
    if not args.device_path:
        print("Error: Device path required", file=sys.stderr)
        return 1
    
    try:
        snapshot = await client.take_snapshot(args.device_path)
        
        if args.format == 'json':
            print(json.dumps({
                'device_path': snapshot.device_path,
                'filename': snapshot.filename,
                'timestamp': snapshot.timestamp,
                'size_bytes': snapshot.size_bytes
            }, indent=2))
        else:
            print(f"Snapshot taken: {snapshot.filename}")
            print(f"Device: {snapshot.device_path}")
            print(f"Timestamp: {snapshot.timestamp}")
            if snapshot.size_bytes:
                print(f"Size: {snapshot.size_bytes} bytes")
        
        return 0
        
    except Exception as e:
        print(f"Error taking snapshot: {e}", file=sys.stderr)
        return 1


async def cmd_record(client: CameraClient, args) -> int:
    """Start camera recording."""
    if not args.device_path:
        print("Error: Device path required", file=sys.stderr)
        return 1
    
    try:
        recording = await client.start_recording(args.device_path)
        
        if args.format == 'json':
            print(json.dumps({
                'device_path': recording.device_path,
                'recording_id': recording.recording_id,
                'filename': recording.filename,
                'start_time': recording.start_time,
                'status': recording.status
            }, indent=2))
        else:
            print(f"Recording started: {recording.filename}")
            print(f"Device: {recording.device_path}")
            print(f"Recording ID: {recording.recording_id}")
            print(f"Start time: {recording.start_time}")
            print(f"Status: {recording.status}")
        
        return 0
        
    except Exception as e:
        print(f"Error starting recording: {e}", file=sys.stderr)
        return 1


async def cmd_stop(client: CameraClient, args) -> int:
    """Stop camera recording."""
    if not args.device_path:
        print("Error: Device path required", file=sys.stderr)
        return 1
    
    try:
        recording = await client.stop_recording(args.device_path)
        
        if args.format == 'json':
            print(json.dumps({
                'device_path': recording.device_path,
                'recording_id': recording.recording_id,
                'filename': recording.filename,
                'start_time': recording.start_time,
                'duration': recording.duration,
                'status': recording.status
            }, indent=2))
        else:
            print(f"Recording stopped: {recording.filename}")
            print(f"Device: {recording.device_path}")
            print(f"Recording ID: {recording.recording_id}")
            print(f"Duration: {recording.duration or 'Unknown'}")
            print(f"Status: {recording.status}")
        
        return 0
        
    except Exception as e:
        print(f"Error stopping recording: {e}", file=sys.stderr)
        return 1


async def cmd_ping(client: CameraClient, args) -> int:
    """Test connection with ping."""
    try:
        result = await client.ping()
        
        if args.format == 'json':
            print(json.dumps({'result': result}, indent=2))
        else:
            print(f"Ping result: {result}")
        
        return 0
        
    except Exception as e:
        print(f"Error pinging service: {e}", file=sys.stderr)
        return 1


def cli_main():
    """Synchronous wrapper for the async main function."""
    return asyncio.run(main())


if __name__ == "__main__":
    sys.exit(cli_main())
