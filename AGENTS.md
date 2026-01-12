# Agent Investigation Protocol for Dependency Bumps

This document provides a step-by-step protocol for autonomous agents investigating library dependency updates. Follow this protocol systematically to uncover silent breaking changes that aren't documented in release notes.

## Protocol Overview

This is a **multi-phase investigation** that must be completed in order. Do not skip phases even if they seem unnecessary. Silent breaking changes hide in places you don't expect.

## Phase 1: Gather All Changes

### Step 1.1: Extract Complete Change List

For the PR being investigated:

1. Fetch the complete list of changed files
2. Categorize files into:
   - Non-vendor code changes (application code)
   - Dependency manifest changes (go.mod, go.sum, package.json, etc.)
   - Vendor directory changes
3. Count the changes:
   - How many non-vendor files changed?
   - How many dependencies updated?
   - How large are the vendor changes?

### Step 1.2: Identify Dependency Version Changes

From dependency manifests (go.mod, etc.):

1. List EVERY dependency that changed version
2. For each dependency:
   - Old version
   - New version
   - Version jump type (major, minor, patch)
   - Is it a direct or transitive dependency?

### Step 1.3: Read All Non-Vendor Code Changes

For EACH non-vendor file that changed:

1. Read the complete diff
2. Document what changed:
   - Function calls added/removed/modified
   - Parameters added/removed/modified
   - Imports added/removed
   - New types or interfaces
3. For EACH change, ask: "What forced this change?"

## Phase 2: Analyze Code Change Patterns

### Step 2.1: Pattern Recognition

Look for patterns across all code changes:

1. **Parameter Removal Pattern**
   - Are parameters being removed from function calls?
   - Is the same parameter being removed in multiple places?
   - What type was that parameter?
   - What was that parameter's role?

2. **Parameter Addition Pattern**
   - Are new parameters being added?
   - What are the new parameters?
   - What values are being passed?
   - Where do those values come from?

3. **Import Changes Pattern**
   - Are new packages being imported?
   - Are imports being removed?
   - Are import paths changing?
   - What do the new imports provide?

4. **Type Changes Pattern**
   - Are types being converted?
   - Are new types being constructed?
   - Are interfaces changing?

### Step 2.2: Infer What Changed in Dependencies

Based on the patterns identified:

1. What API contracts changed?
2. What services moved from one place to another?
3. What initialization steps are new?
4. What behavior is being explicitly configured that was implicit before?

## Phase 3: Deep Dive into Function Signature Changes

For EACH function whose signature changed:

### Step 3.1: Document the Change

Create a table:
```
Function: <name>
Old Signature: <full signature>
New Signature: <full signature>
Removed: <list parameters removed>
Added: <list parameters added>
Changed: <list parameters with different types>
```

### Step 3.2: Investigate Removed Parameters

For EACH removed parameter:

1. **Identify the Parameter Type**
   - What was its type?
   - Was it an interface or concrete type?
   - Search documentation for this type

2. **List All Methods/Properties**
   - What methods did this type have?
   - What properties/fields did it expose?
   - What services did it provide access to?

3. **Trace Parameter Usage**
   - In the OLD code, search for ALL uses of this parameter
   - What methods were called on it?
   - What properties were accessed?
   - Was it passed to other functions?

4. **Service Availability Check**
   For each service the parameter provided, verify:
   - Is this service still available another way?
   - How is it accessed now?
   - Is there any code that initializes/configures it?
   - What happens if it's not initialized?

### Step 3.3: Critical Service Checklist

For the removed parameter, check if it provided access to:

- [ ] Logger/Logging system
- [ ] Metrics registry
- [ ] Event recorder
- [ ] Configuration object
- [ ] Client (Kubernetes client, HTTP client, etc.)
- [ ] Scheme (for type registration)
- [ ] Cache
- [ ] REST Mapper
- [ ] Leader election
- [ ] Rate limiter
- [ ] Error handler
- [ ] Context
- [ ] Tracer

If ANY of these boxes are checked, you MUST investigate further.

## Phase 4: Trace Service Initialization

For each critical service identified in Phase 3:

### Step 4.1: Map Old Initialization Path

Using the OLD code (before the PR):

1. Where was the service created?
2. How did it reach the function that uses it?
3. What defaults were set during creation?
4. What configuration was applied?
5. When in the application lifecycle was it initialized?

Create a flow diagram:
```
App Start → Manager Init → Service Creation → Config Applied → Passed to Function → Used
```

### Step 4.2: Map New Initialization Path

Using the NEW code (after the PR):

1. Where is the service created now?
2. How does it reach the function that uses it now?
3. What defaults are set now?
4. What configuration is applied now?
5. When is it initialized now?

Create a flow diagram for the new path.

### Step 4.3: Compare Initialization Paths

Compare the two diagrams:

- [ ] Are there steps in the old path that don't exist in the new path?
- [ ] Are there services initialized in the old path but not the new path?
- [ ] Are there defaults set in the old path but not the new path?
- [ ] Are there configuration steps that are skipped in the new path?

**If ANY of these boxes are checked, you have found a potential silent breaking change.**

### Step 4.4: Verify Each Service Endpoint

For EACH service (logger, metrics, events, etc.), verify the complete path:

```
1. Service Construction
   OLD: <how it was constructed>
   NEW: <how it is constructed now>

2. Configuration Applied
   OLD: <what config was applied>
   NEW: <what config is applied now>

3. Defaults Set
   OLD: <what defaults were set>
   NEW: <what defaults are set now>

4. Injection/Passing
   OLD: <how it reached the consumer>
   NEW: <how it reaches the consumer now>

5. Runtime Usage
   OLD: <where/how it was used>
   NEW: <where/how it is used now>
```

If there are ANY differences in steps 1-4, investigate what the runtime impact is.

## Phase 5: Fetch and Analyze Upstream Changes

### Step 5.1: Get Release Notes

For each major dependency that changed:

1. Fetch the official release notes
2. Look for sections titled:
   - "Breaking Changes"
   - "Upgrade Guide"
   - "Migration"
   - "Deprecations"
   - "Behavioral Changes"
3. Document ALL breaking changes mentioned

### Step 5.2: Find the Actual Pull Requests

For each breaking change in release notes:

1. Search the dependency's repository for the PR
2. Read the full PR description
3. Read the code changes
4. Read all comments on the PR
5. Look for mentions of:
   - Side effects
   - Migration requirements
   - Things users need to do manually
   - Known issues

### Step 5.3: Search for Issues

Search the dependency's issue tracker for:

1. **Issues filed AFTER the release**
   - Query: `is:issue created:>YYYY-MM-DD label:bug`
   - Where YYYY-MM-DD is the release date
   - These reveal problems not caught during development

2. **Migration-related issues**
   - Query: `is:issue "migration" OR "upgrade" OR "breaking change"`
   - Read discussions about upgrade problems
   - Note any workarounds or fixes

3. **Specific function/type issues**
   - Query: `is:issue "FunctionName"` (the function that changed)
   - Look for reports of unexpected behavior
   - Check if the issue was introduced by this release

### Step 5.4: Check for Follow-up Releases

1. Were there patch releases after the version being adopted?
2. What did those patches fix?
3. Do the fixes relate to the changes in this PR?
4. Should the PR update to the patched version instead?

## Phase 6: Investigate Default Value Changes

### Step 6.1: Identify Configuration Structures

Find all configuration structures used by the updated dependency:

1. Search for types named *Config, *Options, *Settings
2. For each structure, document:
   - All fields
   - Default value for each field
   - What each field controls

### Step 6.2: Compare Default Values

For EACH field in configuration structures:

1. **OLD behavior**: What was the default in the old version?
2. **NEW behavior**: What is the default in the new version?
3. **Changed?**: Did the default change?

If a default changed, this is a potential silent breaking change.

### Step 6.3: Default Application Check

For configuration that changed:

1. Is this configuration explicitly set in the application code?
2. Or was it relying on the library's default?
3. If relying on defaults, the behavior will change silently.

**Action**: For each default that changed, verify if the application explicitly sets it.

## Phase 7: Runtime Behavior Analysis

### Step 7.1: Observability Impact Assessment

For EACH code change, verify these specific behaviors are still functional:

#### Logging
- [ ] Is a logger initialized?
- [ ] Is the logger passed to all components that need it?
- [ ] Are there code paths that log errors?
- [ ] Can you trace the logger from initialization to usage?
- [ ] Are there any null/noop logger instances?
- [ ] What happens if the logger is not initialized?

#### Metrics
- [ ] Are metrics registered?
- [ ] Are metric collectors initialized?
- [ ] Are metrics emitted in the hot path?
- [ ] Are metrics namespaced correctly?
- [ ] Can you trace metrics from registration to emission?

#### Events
- [ ] Is an event recorder created?
- [ ] Is it passed to controllers/reconcilers?
- [ ] Are events recorded for important state changes?
- [ ] Can you trace the event recorder from creation to usage?

#### Error Handling
- [ ] Are errors still logged?
- [ ] Are errors still wrapped with context?
- [ ] Are errors returned to callers?
- [ ] Are errors converted to events?
- [ ] Is there error suppression anywhere?

### Step 7.2: Performance Impact Assessment

Check these performance-critical settings:

- [ ] Rate limiting: Is it still configured?
- [ ] Concurrency limits: Are they still set?
- [ ] Timeout values: Have they changed?
- [ ] Retry behavior: Is it still the same?
- [ ] Cache settings: Are they still configured?
- [ ] Queue sizes: Have they changed?

### Step 7.3: Control Flow Analysis

For critical control paths:

1. **Startup**
   - Trace from main() to all component initialization
   - Verify all components are initialized
   - Check if initialization errors are logged

2. **Reconciliation Loop**
   - Trace a complete reconciliation
   - Verify logging at key points
   - Verify metrics are emitted
   - Check error handling

3. **Watch Handling**
   - Trace from watch trigger to reconcile
   - Verify errors are logged
   - Check for dropped events

4. **Shutdown**
   - Trace graceful shutdown path
   - Verify cleanup happens
   - Check for resource leaks

## Phase 8: Vendor Code Investigation

Even though vendor changes are auto-generated, investigate them:

### Step 8.1: Find High-Risk Changes

Search vendor changes for:

1. Deleted files
   - Any files removed entirely?
   - What functionality did they provide?
   - Is that functionality used anywhere?

2. Changed initialization code
   - Search for: `func New`, `func Default`, `func Init`
   - Compare old vs new implementations
   - What changed in the initialization logic?

3. Changed default values
   - Search for: `= Default`, `defaultValue`, `const default`
   - Compare old vs new values
   - Document all changes

### Step 8.2: Check Internal Implementation Changes

Look for changes in:

1. **Controller implementation**
   - How are controllers started?
   - How are watches registered?
   - How are reconciliation requests queued?

2. **Manager implementation**
   - How are runnables started?
   - How are clients initialized?
   - How is configuration propagated?

3. **Logging setup**
   - How is the logger initialized?
   - How is it propagated to controllers?
   - What is the null logger behavior?

## Phase 9: Create Investigation Report

### Required Sections

#### 1. Breaking Changes (Compilation)

For each breaking change:
```
### Change: <description>
- **Location**: file:line
- **Old**: <old code>
- **New**: <new code>
- **Fix Applied**: <what the PR did to fix it>
- **Complete**: Yes/No (is the fix complete?)
```

#### 2. Silent Breaking Changes (Runtime)

For EACH silent breaking change found:
```
### Silent Change: <description>
- **Category**: Observability/Performance/Security/Configuration/Control Flow
- **Previous Behavior**: <how it worked before>
- **New Behavior**: <how it works now>
- **Impact**: <what breaks or degrades>
- **Detection Method**: <how would you discover this>
- **Risk Level**: Critical/High/Medium/Low
- **Affected Components**: <list of components>
- **Root Cause**: <what change caused this>
- **Recommendation**: <what should be done>
```

#### 3. Initialization Path Changes

For each service that changed initialization:
```
### Service: <service name>
- **Old Initialization Path**: <diagram/description>
- **New Initialization Path**: <diagram/description>
- **Missing Steps**: <what steps were removed>
- **Impact**: <what happens without these steps>
```

#### 4. Configuration Default Changes

```
### Configuration: <config name>
- **Field**: <field name>
- **Old Default**: <value>
- **New Default**: <value>
- **Explicitly Set**: Yes/No
- **Impact**: <what changes if relying on default>
```

#### 5. Required Actions

List specific actions needed:
```
- [ ] Action item 1
- [ ] Action item 2
- [ ] Action item 3
```

## Investigation Checklist

Before concluding your investigation, verify you completed ALL steps:

### Phase Completion
- [ ] Phase 1: Gathered all changes
- [ ] Phase 2: Analyzed patterns
- [ ] Phase 3: Deep dived into signature changes
- [ ] Phase 4: Traced service initialization
- [ ] Phase 5: Analyzed upstream changes
- [ ] Phase 6: Investigated default values
- [ ] Phase 7: Analyzed runtime behavior
- [ ] Phase 8: Investigated vendor code
- [ ] Phase 9: Created investigation report

### Service Verification (for EACH changed function)
- [ ] Logger initialization path verified
- [ ] Metrics registration path verified
- [ ] Event recorder path verified
- [ ] Error handling path verified
- [ ] Configuration loading verified
- [ ] Rate limiting verified
- [ ] Timeout configuration verified
- [ ] Retry behavior verified

### Documentation Review
- [ ] Release notes reviewed
- [ ] Breaking changes documented
- [ ] Migration guide reviewed
- [ ] Related PRs reviewed
- [ ] Related issues reviewed
- [ ] Follow-up releases checked

### Runtime Impact
- [ ] Startup behavior verified
- [ ] Reconciliation behavior verified
- [ ] Error path behavior verified
- [ ] Shutdown behavior verified
- [ ] Performance impact assessed
- [ ] Security impact assessed

## Key Principles

1. **Never assume compatibility**: Even patch versions can break behavior
2. **Trust code, not docs**: Documentation lags reality
3. **Follow the data flow**: Trace initialization paths completely
4. **Question every change**: If code changed, behavior changed
5. **Check implicit dependencies**: Removed code was doing something
6. **Verify observability**: If you can't see it, you can't debug it

## Common Pitfalls to Avoid

- ❌ Stopping at release notes
- ❌ Assuming tests catch everything
- ❌ Trusting "backward compatible" claims
- ❌ Skipping vendor code review
- ❌ Not tracing initialization paths
- ❌ Ignoring default value changes
- ❌ Not checking for follow-up fixes
- ❌ Missing cascading effects

## Success Criteria

Your investigation is complete when you can answer ALL these questions:

1. What services were previously auto-initialized that might not be now?
2. What default values changed and what is the impact?
3. Are all observability systems (logging, metrics, events) still functional?
4. Are all performance controls (rate limiting, timeouts) still in place?
5. Are all security controls still active?
6. What could break at runtime that wouldn't break at compile time?
7. What would be invisible in production without proper monitoring?

If you cannot answer any of these questions with confidence, continue investigating.
