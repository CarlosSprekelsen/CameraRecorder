# MediaMTX Camera Service - Go Implementation

⚠️ **MIGRATION PROJECT - DO NOT TOUCH PYTHON SERVER**

## Project Status
- **Python server**: `../mediamtx-camera-service/` (PROTECTED - DO NOT MODIFY)
- **This directory**: Go migration workspace only
- **Purpose**: Zero-risk migration from Python to Go implementation

## Protection Rules
1. **NEVER** modify files in `../mediamtx-camera-service/`
2. **NEVER** run commands that could affect the Python server
3. **ALWAYS** verify Python server integrity before and after operations
4. **ONLY** work within this Go project directory

## Directory Structure
```
../mediamtx-camera-service/          # ← ORIGINAL PYTHON (UNTOUCHED)
../mediamtx-camera-service-go/       # ← NEW GO PROJECT (this directory)
├── README.md                        # This file
├── go.mod                           # Go dependencies
├── cmd/server/                      # Go application entry point
├── internal/                        # Go internal packages
├── pkg/                            # Go public packages
├── docs/                           # Migrated documentation
└── evidence/                        # Clean evidence directory
```

## Development Guidelines
- All Go development happens in this directory only
- Documentation is migrated from Python project with Go-specific updates
- Business requirements remain the same, only technology stack changes
- Performance targets are enhanced for Go implementation

## Verification Commands
```bash
# Verify Python server is untouched
ls ../mediamtx-camera-service/src/  # Should see Python files

# Verify Go project structure
ls cmd/ internal/ pkg/              # Should see Go directories
```

**Next Steps**: Begin Go implementation in this isolated environment with zero risk to current Python server.
