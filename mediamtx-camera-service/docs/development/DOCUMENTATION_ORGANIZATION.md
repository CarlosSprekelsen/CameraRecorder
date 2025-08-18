# Documentation Organization Guide

**Version:** 1.0  
**Last Updated:** 2025-01-15  
**Purpose:** Maintain proper documentation and evidence organization

---

## Directory Structure

### Root Directory
The root directory should contain only essential project files:
- `README.md` - Project overview
- `CHANGELOG.md` - Version history
- `LICENSE` - Project license
- `requirements.txt` - Python dependencies
- `pyproject.toml` - Project configuration
- `Makefile` - Build automation
- Configuration files (`.flake8`, `mypy.ini`, etc.)

### Documentation (`docs/`)
- `roadmap.md` - Project roadmap and status
- `epics/` - Epic documentation and requirements
- `development/` - Development guidelines and processes
- `deployment/` - Deployment documentation
- `requirements/` - Requirements baseline
- `security/` - Security documentation
- `architecture/` - System architecture
- `api/` - API documentation
- `examples/` - Usage examples

### Evidence (`evidence/`)
Organized by epic and type:
- `e1/` - Epic E1 evidence
- `e2/` - Epic E2 evidence
- `e3/` - Epic E3 evidence
- `e5/` - Epic E5 evidence
- `e6/` - Epic E6 evidence
- `qa/` - Quality assurance reports
- `security-scans/` - Security scan results
- `test-artifacts/` - Test artifacts and logs
- `sdr/` - System Design Review evidence
- `pdr/` - Preliminary Design Review evidence
- `cdr/` - Critical Design Review evidence

---

## File Organization Rules

### Evidence Files
- **QA Reports:** `evidence/qa/`
- **Security Scans:** `evidence/security-scans/`
- **Test Artifacts:** `evidence/test-artifacts/`
- **Epic-Specific Evidence:** `evidence/e{number}/`

### Temporary Files
- **Test Scripts:** `evidence/test-artifacts/`
- **Debug Scripts:** `evidence/test-artifacts/`
- **Installation Logs:** `evidence/qa/`
- **Uninstall Reports:** `evidence/qa/`

### Build Artifacts
- **Coverage Reports:** `evidence/test-artifacts/`
- **Performance Logs:** `evidence/test-artifacts/`
- **Integration Logs:** `evidence/test-artifacts/`

---

## Cleanup Procedures

### Before Commits
1. Move any test artifacts to `evidence/test-artifacts/`
2. Move QA reports to `evidence/qa/`
3. Move security scans to `evidence/security-scans/`
4. Remove temporary directories (`.tmp_*`, `test_env_*`)
5. Ensure `.gitignore` is updated for new file types

### After Epic Completion
1. Organize epic-specific evidence in `evidence/e{number}/`
2. Update roadmap.md with completion status
3. Archive test artifacts to `evidence/test-artifacts/`
4. Review and clean up temporary files

---

## .gitignore Maintenance

The `.gitignore` file should exclude:
- Temporary directories (`.tmp_*`, `test_env_*`)
- Test artifacts and logs
- Security scan results
- Coverage reports
- Virtual environments
- IDE files
- OS-specific files

---

## Quality Gates

### Documentation Review
- [ ] All evidence properly organized
- [ ] No unwanted files in root directory
- [ ] .gitignore updated for new file types
- [ ] Directory structure maintained
- [ ] Documentation links updated

### Before Epic Completion
- [ ] Evidence organized in appropriate directories
- [ ] Test artifacts archived
- [ ] QA reports filed
- [ ] Security scans documented
- [ ] Roadmap updated

---

## Maintenance Schedule

- **Weekly:** Review root directory for unwanted files
- **Per Epic:** Organize evidence and update documentation
- **Per Sprint:** Update roadmap and clean artifacts
- **Per Release:** Archive old evidence and update .gitignore
