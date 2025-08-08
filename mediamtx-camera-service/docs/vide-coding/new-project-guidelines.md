#Step 1: Define project objectives

Add to the Project Objectives the `project_objectives_footer` Text - Simple checklist to catch over-engineering


#Step 2: Project Instructions
Copy and paste prokect ground rules (project-ground-rules.md) into Specific project instructuons field 

Populate structure with generig corss project documentation:
Development Principles: docs/development/principles.md
Documentation Guidelines: docs/development/documentation-guidelines.md
Roles and Responsibilities: docs/development/roles-responsibilities.md

add project speficics as ther are baselined (eg requirements, architechture, etc,, based on standards V model workflow.)

Here are the only relevant artifacts:


**The approach is:**
- Universal docs: principles.md, documentation-guidelines.md, roles-responsibilities.md
- Project-specific: requirements, architecture, etc.
- Evidence as temporary coordination tool

**Role-based prompt examples:**

```
Your role: Project Manager
Follow project ground rules. Based on current backlog, generate sprint plan with clear scope definition and acceptance criteria for next sprint.
```

```
Your role: Developer  
Follow project ground rules. Implement [specific feature] per sprint requirements. No scope additions, use STOP comments for ambiguities.
```

```
Your role: IV&V
Follow project ground rules. Review [developer's implementation] against sprint requirements. Apply quality standards, reject stubs/over-mocking.
```

