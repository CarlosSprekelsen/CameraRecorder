# Step 1: Define project objectives

Add to the Project Objectives the `project_objectives_footer` Text - Simple checklist to catch over-engineering

# Step 2: Project Instructions
Copy and paste project ground rules (project-ground-rules.md) into Specific project instructions field 

Populate structure with generic cross project documentation:
- Development Principles: docs/development/principles.md
- Documentation Guidelines: docs/development/documentation-guidelines.md
- Roles and Responsibilities: docs/development/roles-responsibilities.md

Add project specifics as they are baselined (eg requirements, architecture, etc, based on standards V model workflow.)

# Step 3: Universal Prompt Template
For AI interfaces without project instruction fields (VS Code Copilot, etc.), use this header in every prompt:

```
Your role: [Role Name]
Ground rules: docs/development/project-ground-rules.md  
Role reference: docs/development/roles-responsibilities.md
Task: [specific request]
```

**Role-based prompt examples:**

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Based on current backlog, generate sprint plan with clear scope definition and acceptance criteria for next sprint.
```

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Implement [specific feature] per sprint requirements. No scope additions, use STOP comments for ambiguities.
```

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Review [developer's implementation] against sprint requirements. Apply quality standards, reject stubs/over-mocking.
```

**The approach:**
- Universal docs: principles.md, documentation-guidelines.md, roles-responsibilities.md
- Project-specific: requirements, architecture, etc.
- Evidence as temporary coordination tool
- Universal prompt header ensures consistent behavior across all AI interfaces