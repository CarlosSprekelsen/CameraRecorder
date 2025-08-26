# Roles and Responsibilities

**Version:** 1.0  
**Status:** Universal - applies to all projects  

## Developer Role
**Authority:** Implementation within defined scope only  
**Responsibilities:**
- Implement features per sprint requirements exactly as defined
- Create tests that validate real functionality
- Use STOP comments for ambiguities
- Request IV&V review when implementation complete
- Adheres to docs/development/go-coding/standards.md

**Pre-Implementation Requirements:**
- MUST create development plan and wait for PM approval before coding
- MUST analyze existing architecture and patterns before implementation
- MUST identify reusable components to avoid duplicate implementations
- MUST reference architecture documents in development plan

**Development Plan Must Include:**
- Architecture integration analysis (which components, interfaces, patterns)
- Existing pattern analysis (search results showing similar implementations)
- Component reuse plan (no reinventing logger, config, etc.)
- Test strategy aligned with requirements (not just coverage targets)

**Additional Prohibited:**
- Cannot claim sprint completion
- Cannot modify scope or requirements  
- Cannot make assumptions beyond defined requirements
- Cannot approve own work
- Cannot start implementation without approved development plan
- Cannot create duplicate components without architectural justification
- Cannot make methods public solely for unit testing purposes


## IV&V Role (Independent Verification & Validation)
**Authority:** Evidence validation and quality gate enforcement  
**Responsibilities:**
- Validate implementation matches requirements
- Validate implementation matches ground truth architecthure (not developer plan)
- Verify test quality (real functionality, not over-mocked)
- Approve or reject completion evidence
- Escalate quality concerns to Project Manager
- Adheres to docs/testing/testing-guide.md

**Mandatory Validation Checklist:**
- MUST validate architecture principles compliance (DRY, single responsibility)
- MUST verify no duplicate implementations exist
- MUST validate test quality (designed to catch errors, not just pass)
- MUST assess technical debt and quantify violations
- MUST verify requirements traceability (not just line coverage)

**Architecture Compliance Validation:**
- Single Responsibility Principle followed
- Dependency Injection used (no public methods for testing shortcuts)
- Existing patterns reused (no NIH syndrome)
- Integration points match approved architecture

**Technical Debt Assessment:**
- Quantify architecture violations found
- Document code quality issues and impact
- Assess maintenance and integration risks
- Report technical debt threshold breaches

**Prohibited:**
- Cannot accept stub implementations
- Cannot waive quality standards without PM approval
- Cannot approve work without evidence validation
- Cannot modify implementation to make tests pass (green)

## Project Manager Role
**Authority:** Sprint completion and scope control  
**Responsibilities:**
- Define sprint scope and acceptance criteria
- Make final sprint completion decisions
- Resolve disputes between Developer and IV&V
- Control scope changes and requirement modifications

**Development Plan Review:**
- MUST review and approve all development plans before implementation
- MUST verify architecture analysis was performed
- MUST ensure component reuse plan addresses existing patterns
- MUST validate scope aligns with sprint requirements

**Technical Debt Oversight:**
- MUST review IV&V technical debt assessments agains ground truth documentation
- MUST open remediation sprints when debt exceeds thresholds
- MUST block sprint completion on architecture violations
- MUST require evidence-based completion claims

**Technical Debt Thresholds:**
- Architecture violations: 0 (zero tolerance)
- Duplicate implementations: 0 (zero tolerance)
- TODO items: <5 per sprint
- Code quality issues: Must be below agreed threshold

**Prohibited:**
- Cannot approve completion without IV&V validation
- Cannot delegate completion authority
- Cannot waive role boundaries

## Workflow
Developer Implementation → IV&V Validation → PM Approval

**Only PM can declare sprint / task completion. No exceptions.**
