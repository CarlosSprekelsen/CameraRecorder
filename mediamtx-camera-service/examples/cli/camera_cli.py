#!/usr/bin/env python3
"""
MediaMTX Camera Service CLI Tool

A command-line interface for controlling cameras through the MediaMTX Camera Service.
Supports JWT and API key authentication with comprehensive camera operations.

Usage:
    # Development environment (port 8080)
    python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list
    python camera_cli.py --host localhost --port 8080 --auth-type api_key --key your_api_key snapshot /dev/video0
    python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token record /dev/video0 --duration 30
    
    # Production environment (port 8002)
    python camera_cli.py --host localhost --port 8002 --auth-type jwt --token your_jwt_token list
    python camera_cli.py --host localhost --port 8002 --auth-type api_key --key your_api_key snapshot /dev/video0
    python camera_cli.py --host localhost --port 8002 --auth-type jwt --token your_jwt_token record /dev/video0 --duration 30
"""

import asyncio
import argparse
import json
import sys
import time
from typing import List
from pathlib import Path

# Add the parent directory to the path to import the camera client
sys.path.append(str(Path(__file__).parent.parent.parent))

from examples.python.camera_client import (
    CameraClient,
    CameraServiceError,
    CameraNotFoundError,
    MediaMTXError
)


class CameraCLI:
    """
    Command-line interface for MediaMTX Camera Service.
    
    Provides a comprehensive CLI for camera control operations including
    listing cameras, taking snapshots, recording, and status monitoring.
    """

    def __init__(self):
        self.client = None
        self.parser = self._create_parser()

    def _create_parser(self) -> argparse.ArgumentParser:
        """Create the command-line argument parser."""
        parser = argparse.ArgumentParser(
            description="MediaMTX Camera Service CLI Tool",
            formatter_class=argparse.RawDescriptionHelpFormatter,
            epilog="""
Examples:
  # Development environment (port 8080)
  %(prog)s --host localhost --port 8080 --auth-type jwt --token your_token list
  %(prog)s --host localhost --port 8080 --auth-type jwt --token your_token status /dev/video0
  %(prog)s --host localhost --port 8080 --auth-type jwt --token your_token snapshot /dev/video0
  %(prog)s --host localhost --port 8080 --auth-type jwt --token your_token record /dev/video0 --duration 30
  
  # Production environment (port 8002)
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token list
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token status /dev/video0
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token snapshot /dev/video0
  %(prog)s --host localhost --port 8002 --auth-type jwt --token your_token record /dev/video0 --duration 30
            """
        )

        # Connection arguments
        parser.add_argument('--host', default='localhost', help='Server hostname')
        parser.add_argument('--port', type=int, default=8080, help='Server port')
        parser.add_argument('--ssl', action='store_true', help='Use SSL/TLS')
        parser.add_argument('--auth-type', choices=['jwt', 'api_key'], default='jwt', help='Authentication type')
        parser.add_argument('--token', help='JWT token')
        parser.add_argument('--key', help='API key')
        parser.add_argument('--timeout', type=int, default=30, help='Request timeout in seconds')

        # Command and arguments
        parser.add_argument('command', help='Command to execute')
        parser.add_argument('device_path', nargs='?', help='Camera device path')
        parser.add_argument('--duration', type=int, help='Recording duration in seconds')
        parser.add_argument('--filename', help='Custom filename for snapshot or recording')
        parser.add_argument('--format', choices=['table', 'json', 'csv'], default='table', help='Output format')
        parser.add_argument('--verbose', '-v', action='store_true', help='Verbose output')

        return parser

    async def run(self, args: List[str]) -> int:
        """
        Run the CLI with the given arguments.
        
        Args:
            args: Command-line arguments
            
        Returns:
            Exit code (0 for success, non-zero for error)
        """
        try:
            parsed_args = self.parser.parse_args(args)
            
            # Validate authentication
            if parsed_args.auth_type == 'jwt' and not parsed_args.token:
                print("❌ Error: JWT token required for jwt authentication", file=sys.stderr)
                return 1
            elif parsed_args.auth_type == 'api_key' and not parsed_args.key:
                print("❌ Error: API key required for api_key authentication", file=sys.stderr)
                return 1

            # Create client
            self.client = CameraClient(
                host=parsed_args.host,
                port=parsed_args.port,
                use_ssl=parsed_args.ssl,
                auth_type=parsed_args.auth_type,
                auth_token=parsed_args.token,
                api_key=parsed_args.key
            )

            # Connect to service
            if parsed_args.verbose:
                print(f"🔌 Connecting to {parsed_args.host}:{parsed_args.port}...")
            
            await self.client.connect()
            
            if parsed_args.verbose:
                print("✅ Connected successfully")

            # Execute command
            return await self._execute_command(parsed_args)

        except KeyboardInterrupt:
            print("\n🛑 Operation cancelled by user", file=sys.stderr)
            return 130
        except Exception as e:
            print(f"❌ Error: {e}", file=sys.stderr)
            return 1
        finally:
            if self.client:
                await self.client.disconnect()

    async def _execute_command(self, args) -> int:
        """Execute the specified command."""
        command = args.command.lower()
        
        try:
            if command == 'list':
                return await self._cmd_list(args)
            elif command == 'status':
                return await self._cmd_status(args)
            elif command == 'snapshot':
                return await self._cmd_snapshot(args)
            elif command == 'record':
                return await self._cmd_record(args)
            elif command == 'stop':
                return await self._cmd_stop(args)
            elif command == 'ping':
                return await self._cmd_ping(args)
            elif command == 'monitor':
                return await self._cmd_monitor(args)
            else:
                print(f"❌ Unknown command: {command}", file=sys.stderr)
                self.parser.print_help()
                return 1
                
        except CameraServiceError as e:
            print(f"❌ Camera service error: {e}", file=sys.stderr)
            return 1
        except Exception as e:
            print(f"❌ Unexpected error: {e}", file=sys.stderr)
            return 1

    async def _cmd_list(self, args) -> int:
        """List available cameras."""
        try:
            cameras = await self.client.get_camera_list()
            
            if args.format == 'json':
                print(json.dumps([{
                    'device_path': camera.device_path,
                    'name': camera.name,
                    'status': camera.status,
                    'capabilities': camera.capabilities,
                    'stream_url': camera.stream_url
                } for camera in cameras], indent=2))
            elif args.format == 'csv':
                print("device_path,name,status,capabilities,stream_url")
                for camera in cameras:
                    capabilities = ','.join(camera.capabilities) if camera.capabilities else ''
                    stream_url = camera.stream_url or ''
                    print(f"{camera.device_path},{camera.name},{camera.status},{capabilities},{stream_url}")
            else:
                if not cameras:
                    print("📹 No cameras available")
                    return 0
                
                print(f"📹 Found {len(cameras)} camera(s):")
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
            print(f"❌ Failed to list cameras: {e}", file=sys.stderr)
            return 1

    async def _cmd_status(self, args) -> int:
        """Get status of specific camera."""
        if not args.device_path:
            print("❌ Error: Device path required for status command", file=sys.stderr)
            return 1
        
        try:
            status = await self.client.get_camera_status(args.device_path)
            
            if args.format == 'json':
                print(json.dumps({
                    'device_path': status.device_path,
                    'name': status.name,
                    'status': status.status,
                    'capabilities': status.capabilities,
                    'stream_url': status.stream_url
                }, indent=2))
            else:
                print(f"📹 Camera Status: {status.name}")
                print(f"   Device: {status.device_path}")
                print(f"   Status: {status.status}")
                print(f"   Capabilities: {', '.join(status.capabilities) if status.capabilities else 'None'}")
                if status.stream_url:
                    print(f"   Stream: {status.stream_url}")
            
            return 0
            
        except CameraNotFoundError:
            print(f"❌ Camera not found: {args.device_path}", file=sys.stderr)
            return 1

    async def _cmd_snapshot(self, args) -> int:
        """Take a snapshot from camera."""
        if not args.device_path:
            print("❌ Error: Device path required for snapshot command", file=sys.stderr)
            return 1
        
        try:
            if args.verbose:
                print(f"📸 Taking snapshot from {args.device_path}...")
            
            result = await self.client.take_snapshot(
                args.device_path,
                custom_filename=args.filename
            )
            
            if args.format == 'json':
                print(json.dumps(result, indent=2))
            else:
                print(f"✅ Snapshot saved: {result['filename']}")
                if 'file_size' in result:
                    print(f"   Size: {result['file_size']} bytes")
                if 'duration' in result:
                    print(f"   Duration: {result['duration']:.2f} seconds")
            
            return 0
            
        except CameraNotFoundError:
            print(f"❌ Camera not found: {args.device_path}", file=sys.stderr)
            return 1
        except MediaMTXError as e:
            print(f"❌ Snapshot failed: {e}", file=sys.stderr)
            return 1

    async def _cmd_record(self, args) -> int:
        """Start recording from camera."""
        if not args.device_path:
            print("❌ Error: Device path required for record command", file=sys.stderr)
            return 1
        
        try:
            if args.verbose:
                print(f"🎥 Starting recording from {args.device_path}...")
            
            recording = await self.client.start_recording(
                args.device_path,
                duration=args.duration,
                custom_filename=args.filename
            )
            
            if args.format == 'json':
                print(json.dumps({
                    'device_path': recording.device_path,
                    'recording_id': recording.recording_id,
                    'filename': recording.filename,
                    'start_time': recording.start_time,
                    'status': recording.status
                }, indent=2))
            else:
                print(f"✅ Recording started: {recording.filename}")
                print(f"   Recording ID: {recording.recording_id}")
                print(f"   Start Time: {time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(recording.start_time))}")
                if args.duration:
                    print(f"   Duration: {args.duration} seconds")
            
            return 0
            
        except CameraNotFoundError:
            print(f"❌ Camera not found: {args.device_path}", file=sys.stderr)
            return 1
        except MediaMTXError as e:
            print(f"❌ Recording failed: {e}", file=sys.stderr)
            return 1

    async def _cmd_stop(self, args) -> int:
        """Stop recording from camera."""
        if not args.device_path:
            print("❌ Error: Device path required for stop command", file=sys.stderr)
            return 1
        
        try:
            if args.verbose:
                print(f"⏹️ Stopping recording from {args.device_path}...")
            
            result = await self.client.stop_recording(args.device_path)
            
            if args.format == 'json':
                print(json.dumps(result, indent=2))
            else:
                print(f"✅ Recording stopped: {result['filename']}")
                if 'duration' in result:
                    print(f"   Duration: {result['duration']:.2f} seconds")
                if 'file_size' in result:
                    print(f"   Size: {result['file_size']} bytes")
            
            return 0
            
        except CameraNotFoundError:
            print(f"❌ Camera not found: {args.device_path}", file=sys.stderr)
            return 1
        except MediaMTXError as e:
            print(f"❌ Stop recording failed: {e}", file=sys.stderr)
            return 1

    async def _cmd_ping(self, args) -> int:
        """Test connection with ping."""
        try:
            if args.verbose:
                print("🏓 Testing connection...")
            
            pong = await self.client.ping()
            
            if args.format == 'json':
                print(json.dumps({'ping': pong}, indent=2))
            else:
                print(f"✅ Ping response: {pong}")
            
            return 0
            
        except Exception as e:
            print(f"❌ Ping failed: {e}", file=sys.stderr)
            return 1

    async def _cmd_monitor(self, args) -> int:
        """Monitor cameras in real-time."""
        try:
            print("👀 Monitoring cameras (press Ctrl+C to stop)...")
            print()
            
            # Set up event handlers
            async def on_camera_update(params):
                camera = params.get("camera", {})
                timestamp = time.strftime('%H:%M:%S')
                print(f"[{timestamp}] 📹 Camera {camera.get('name', 'Unknown')}: {camera.get('status', 'Unknown')}")
            
            async def on_recording_update(params):
                recording = params.get("recording", {})
                timestamp = time.strftime('%H:%M:%S')
                print(f"[{timestamp}] 🎥 Recording {recording.get('filename', 'Unknown')}: {recording.get('status', 'Unknown')}")
            
            self.client.set_camera_status_callback(on_camera_update)
            self.client.set_recording_status_callback(on_recording_update)
            
            # Initial camera list
            cameras = await self.client.get_camera_list()
            print(f"📹 Monitoring {len(cameras)} camera(s):")
            for camera in cameras:
                print(f"   - {camera.name} ({camera.device_path}): {camera.status}")
            print()
            
            # Keep monitoring
            while True:
                await asyncio.sleep(1)
            
        except KeyboardInterrupt:
            print("\n🛑 Monitoring stopped")
            return 0
        except Exception as e:
            print(f"❌ Monitoring failed: {e}", file=sys.stderr)
            return 1


def main():
    """Main entry point for the CLI."""
    cli = CameraCLI()
    
    # Run the CLI
    exit_code = asyncio.run(cli.run(sys.argv[1:]))
    sys.exit(exit_code)


if __name__ == "__main__":
    main() 