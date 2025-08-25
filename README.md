# MediaMTX Camera Service - Go Implementation

⚠️ **MIGRATION PROJECT - DO NOT TOUCH PYTHON SERVER**

## Project Status
- **Python server**: `../CameraRecorder/mediamtx-camera-service/` (PROTECTED - DO NOT MODIFY)
- **This directory**: Go migration workspace only
- **Purpose**: Zero-risk migration from Python to Go implementation

## Protection Rules
1. **NEVER** modify files in `../CameraRecorder/mediamtx-camera-service/`
2. **NEVER** delete or rename the original Python project
3. **ONLY** work within this Go project directory
4. **ALWAYS** verify Python server remains untouched after any operations

## Directory Structure
```
~/CameraRecorder/
├── mediamtx-camera-service/          # ← ORIGINAL PYTHON (UNTOUCHED)
└── mediamtx-camera-service-go/       # ← NEW GO PROJECT (this directory)
    ├── cmd/server/                   # Go application entry point
    ├── internal/                     # Internal Go packages
    ├── pkg/                         # Public Go packages
    ├── docs/                        # Migrated documentation
    └── evidence/                    # Clean evidence directory
```

## Development Guidelines
- Copy documentation from Python project and update for Go
- Maintain identical business requirements and API contracts
- Update only technology stack references (Python → Go)
- Preserve all functional specifications

## Next Steps
1. Copy and update documentation
2. Implement Go version with same business logic
3. Maintain API compatibility
4. Validate against original requirements

**Remember: The Python server must remain completely functional throughout this migration.**
