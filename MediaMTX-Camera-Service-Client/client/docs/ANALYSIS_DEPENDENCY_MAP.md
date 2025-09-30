# MediaMTX Camera Service Client - Dependency Map & Analysis

## Executive Summary
**Current State**: Mixed UI systems causing chaos and 81 TypeScript errors
**Objective**: Complete Material-UI to Atomic Design migration
**Status**: 11 files still using Material-UI, 26 atomic components exist

---

## 1. CURRENT STATE ANALYSIS

### Material-UI Usage (11 files remaining)
| File | Material-UI Components Used | Atomic Components Available | Missing Components |
|------|----------------------------|---------------------------|-------------------|
| `src/components/Files/Pagination.tsx` | Box, Pagination, Typography, Select, MenuItem, FormControl, InputLabel | Box, Typography, Select | **Pagination, FormControl, InputLabel** |
| `src/components/Files/ConfirmDialog.tsx` | Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography, Alert | Button, Typography, Alert, Dialog | **DialogTitle, DialogContent, DialogActions** |
| `src/components/Files/FileTable.tsx` | Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, IconButton, Box, Typography, CircularProgress, Checkbox, Tooltip | Box, Typography, CircularProgress, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tooltip | **Paper, IconButton, Checkbox** |
| `src/components/organisms/RecordingController/RecordingController.tsx` | PlayArrow, Stop (icons) | Icon | ✅ **All available** |
| `src/components/organisms/ApplicationShell/ApplicationShell.tsx` | Multiple MUI components + icons | Multiple atomic components | **Complex - needs detailed analysis** |
| `src/components/Loading/LoadingSkeleton.tsx` | Skeleton, Box, Card, CardContent | Box, Card, CardContent, Skeleton | ✅ **All available** |
| `src/components/Accessibility/AccessibilityProvider.tsx` | ThemeProvider, createTheme | ❌ **Missing** | **ThemeProvider** |
| `src/components/Cameras/TimedRecordDialog.tsx` | Multiple MUI components | Multiple atomic components | **Needs detailed analysis** |
| `src/components/Security/ProtectedRoute.tsx` | Box, Typography, Alert | Box, Typography, Alert | ✅ **All available** |
| `src/pages/About/AboutPage.tsx` | Icons only | Icon | ✅ **All available** |
| `src/pages/Files/FilesPage.tsx` | Box, Typography, Container, Alert, CircularProgress | Box, Typography, Container, Alert, CircularProgress | ✅ **All available** |

---

## 2. MISSING ATOMIC COMPONENTS ANALYSIS

### Critical Missing Components
| Component | Usage Count | Priority | Implementation Complexity |
|-----------|-------------|----------|-------------------------|
| **Pagination** | 1 file | HIGH | Medium - Custom pagination logic |
| **FormControl** | 1 file | HIGH | Low - Simple wrapper |
| **InputLabel** | 1 file | HIGH | Low - Simple label |
| **DialogTitle** | 1 file | MEDIUM | Low - Simple title wrapper |
| **DialogContent** | 1 file | MEDIUM | Low - Simple content wrapper |
| **DialogActions** | 1 file | MEDIUM | Low - Simple actions wrapper |
| **Paper** | 1 file | MEDIUM | Low - Simple card-like wrapper |
| **IconButton** | 1 file | MEDIUM | Low - Button with icon |
| **Checkbox** | 1 file | MEDIUM | Medium - Form input component |
| **ThemeProvider** | 1 file | HIGH | Medium - Theme context provider |

### Complex Files Requiring Detailed Analysis
1. **ApplicationShell.tsx** - Multiple MUI components, complex layout
2. **TimedRecordDialog.tsx** - Dialog with form components
3. **FileTable.tsx** - Complex table with multiple interactions

---

## 3. DEPENDENCY TREE ANALYSIS

### Atomic Components Dependencies
```
Box (✅ exists)
├── Container (✅ exists)
├── Typography (✅ exists)
├── Button (✅ exists)
├── Alert (✅ exists)
├── CircularProgress (✅ exists)
├── Table (✅ exists)
├── Dialog (✅ exists)
├── Tooltip (✅ exists)
└── Icon (✅ exists)

Missing Components:
├── Pagination (❌ missing)
├── FormControl (❌ missing)
├── InputLabel (❌ missing)
├── DialogTitle (❌ missing)
├── DialogContent (❌ missing)
├── DialogActions (❌ missing)
├── Paper (❌ missing)
├── IconButton (❌ missing)
├── Checkbox (❌ missing)
└── ThemeProvider (❌ missing)
```

### File Replacement Dependencies
```
Low Complexity (Direct replacement):
├── RecordingController.tsx (icons only)
├── LoadingSkeleton.tsx (all components exist)
├── ProtectedRoute.tsx (all components exist)
├── AboutPage.tsx (icons only)
└── FilesPage.tsx (all components exist)

Medium Complexity (Need missing components):
├── Pagination.tsx (needs Pagination, FormControl, InputLabel)
├── ConfirmDialog.tsx (needs DialogTitle, DialogContent, DialogActions)
├── FileTable.tsx (needs Paper, IconButton, Checkbox)
└── AccessibilityProvider.tsx (needs ThemeProvider)

High Complexity (Multiple missing components):
├── ApplicationShell.tsx (complex layout)
└── TimedRecordDialog.tsx (complex dialog)
```

---

## 4. IMPLEMENTATION STRATEGY

### Phase 1: Create Missing Atomic Components (Priority Order)
1. **FormControl + InputLabel** (used in Pagination)
2. **Pagination** (used in Pagination)
3. **DialogTitle, DialogContent, DialogActions** (used in ConfirmDialog)
4. **Paper, IconButton, Checkbox** (used in FileTable)
5. **ThemeProvider** (used in AccessibilityProvider)

### Phase 2: Replace Simple Files First
1. **RecordingController.tsx** (icons only)
2. **LoadingSkeleton.tsx** (all components exist)
3. **ProtectedRoute.tsx** (all components exist)
4. **AboutPage.tsx** (icons only)
5. **FilesPage.tsx** (all components exist)

### Phase 3: Replace Medium Complexity Files
1. **Pagination.tsx** (after creating missing components)
2. **ConfirmDialog.tsx** (after creating missing components)
3. **FileTable.tsx** (after creating missing components)
4. **AccessibilityProvider.tsx** (after creating ThemeProvider)

### Phase 4: Replace Complex Files
1. **ApplicationShell.tsx** (detailed analysis required)
2. **TimedRecordDialog.tsx** (detailed analysis required)

---

## 5. RISK ASSESSMENT

### High Risk Areas
1. **ApplicationShell.tsx** - Core navigation, complex layout
2. **FileTable.tsx** - Complex interactions, state management
3. **ThemeProvider** - Global theming system

### Medium Risk Areas
1. **Pagination.tsx** - Custom pagination logic
2. **ConfirmDialog.tsx** - Modal interactions
3. **TimedRecordDialog.tsx** - Form validation

### Low Risk Areas
1. **Icon replacements** - Simple icon swaps
2. **Basic component replacements** - Direct 1:1 replacements

---

## 6. VALIDATION STRATEGY

### After Each Phase
1. **TypeScript compilation** - `npm run type:check`
2. **Component functionality** - Test basic interactions
3. **Visual consistency** - Check styling matches
4. **No regressions** - Ensure existing functionality works

### Final Validation
1. **All Material-UI imports removed** - `grep -r "@mui" src/`
2. **Zero TypeScript errors** - `npm run type:check`
3. **All components functional** - Manual testing
4. **Consistent UI** - Visual audit

---

## 7. ESTIMATED EFFORT

### Missing Components Creation
- **FormControl + InputLabel**: 30 minutes
- **Pagination**: 1 hour
- **Dialog components**: 45 minutes
- **Paper, IconButton, Checkbox**: 1 hour
- **ThemeProvider**: 45 minutes
- **Total**: ~4 hours

### File Replacements
- **Simple files (5)**: 2 hours
- **Medium complexity (4)**: 3 hours
- **Complex files (2)**: 4 hours
- **Total**: ~9 hours

### **TOTAL ESTIMATED EFFORT: 13 hours**

---

## 8. SUCCESS CRITERIA

### Technical Criteria
- [ ] Zero Material-UI imports in codebase
- [ ] Zero TypeScript compilation errors
- [ ] All atomic components follow consistent patterns
- [ ] All components maintain existing functionality

### Quality Criteria
- [ ] Consistent visual design across all components
- [ ] Proper error handling and loading states
- [ ] Accessible components (ARIA labels, keyboard navigation)
- [ ] Responsive design maintained

### Team Criteria
- [ ] Clear documentation for all atomic components
- [ ] Consistent naming conventions
- [ ] Reusable component patterns established
- [ ] Future Material-UI usage prevented

---

## 9. NEXT STEPS

### Immediate Actions
1. **Create missing atomic components** (Phase 1)
2. **Replace simple files** (Phase 2)
3. **Validate after each phase**
4. **Document any issues or deviations**

### Team Coordination
1. **Share this analysis** with all team members
2. **Assign specific components** to developers
3. **Establish review process** for atomic components
4. **Plan testing strategy** for complex replacements

---

## 10. LESSONS LEARNED

### What Went Wrong
1. **Rushed implementation** without proper analysis
2. **Created duplicate components** without checking existing ones
3. **Fragmented approach** instead of systematic replacement
4. **No validation** after each change

### How to Prevent
1. **Always analyze first** - Create dependency maps
2. **Check existing components** before creating new ones
3. **Systematic approach** - Follow dependency order
4. **Continuous validation** - Check errors after each change
5. **Document everything** - Keep team informed of progress

---

**Document Status**: ✅ Complete
**Last Updated**: Current
**Next Review**: After Phase 1 completion
