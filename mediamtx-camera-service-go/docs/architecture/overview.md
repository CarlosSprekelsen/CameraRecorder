# Architecture Overview - Go Implementation

**Version:** 3.0  
**Date:** 2025-01-15  
**Status:** Go Implementation Architecture - Consolidated  
**Related Epic/Story:** Go Implementation Architecture  

## Document Structure

This architecture documentation has been consolidated into a comprehensive Go implementation guide to eliminate Python references and provide clear implementation guidance.

### Primary Document
- **[Go Architecture Guide](go-architecture-guide.md)** - Complete Go implementation architecture with patterns, guidelines, and code examples

### Archived Documents
The following documents have been archived and superseded by the consolidated guide:
- `archive/codec-compatibility-assessment.md` - H.264 STANAG 4406 compatibility (Go examples in main guide)
- `archive/stream-lifecycle-management.md` - Stream lifecycle patterns (Go examples in main guide)
- `archive/on-demand-stream-activation.md` - On-demand activation patterns (Go examples in main guide)
- `archive/multi-tier-snapshot-capture.md` - Multi-tier snapshot architecture (Go examples in main guide)
- `archive/testing-architecture-decisions.md` - Testing architecture decisions (Go examples in main guide)

## Quick Reference

### Core Architecture Components
- **WebSocket JSON-RPC Server**: gorilla/websocket with 1000+ concurrent connections
- **Camera Discovery Monitor**: goroutines with V4L2 detection (<200ms)
- **MediaMTX Path Manager**: net/http with dynamic path creation
- **Health & Monitoring**: logrus with structured logging

### Key Architecture Patterns
- **Stream Lifecycle Management**: Recording, viewing, and snapshot streams
- **On-Demand Stream Activation**: Power-efficient FFmpeg process management
- **Multi-Tier Snapshot Capture**: Three-tier approach for optimal performance
- **Codec Compatibility**: H.264 STANAG 4406 compliance

### Performance Targets
- **Camera Detection**: <200ms latency
- **WebSocket Response**: <50ms for JSON-RPC methods
- **Memory Usage**: <60MB base, <200MB with 10 cameras
- **Concurrency**: 1000+ WebSocket connections

## Implementation Status

- **Foundation Infrastructure**: ✅ COMPLETED
- **Camera Discovery System**: ✅ COMPLETED
- **WebSocket JSON-RPC Server**: 🔄 IN PROGRESS

## Next Steps

For detailed implementation guidance, architecture patterns, and Go code examples, refer to the **[Go Architecture Guide](go-architecture-guide.md)**.

---

**Document Status**: Architecture overview with consolidated documentation structure  
**Last Updated**: 2025-01-15  
**Next Review**: As needed based on implementation progress
