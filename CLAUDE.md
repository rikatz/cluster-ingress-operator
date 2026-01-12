# Deep Investigation Guide for Dependency Bumps

When analyzing pull requests that bump library versions, especially for core frameworks like controller-runtime, Kubernetes client libraries, or similar foundational dependencies, you must perform a thorough investigation beyond just reading release notes.

## Core Principle

**Release notes only tell you what changed. They don't tell you what broke.**

A seemingly simple API signature change (like removing a parameter) can cascade into multiple silent breaking changes in runtime behavior. Your job is to find these cascading effects.

## Investigation Methodology

### Phase 1: Understand the Full Scope of Changes

1. **Identify ALL Changed Dependencies**
   - List every dependency that changed (direct and transitive)
   - Note the version jump size (patch, minor, major)
   - Larger jumps are higher risk but small patches can still break things

2. **Review ALL Non-Vendor Code Changes**
   - Read every line of code that changed outside of vendor/
   - For EACH code change, ask: "Why was this change necessary?"
   - If code had to change, something in the API or behavior changed

3. **Analyze the Pattern of Changes**
   - Are changes localized or scattered?
   - Do changes follow a pattern (e.g., all adding a parameter)?
   - What does the pattern tell you about what changed in the dependency?

### Phase 2: Deep Dive into Breaking Changes

For each code change identified in Phase 1, perform this analysis:

#### API Signature Changes

When you see a function signature change (added/removed/reordered parameters):

1. **What was removed?**
   - If a parameter was removed, what did that parameter provide?
   - Was it a context? A config object? A manager? A client?

2. **What functionality came from the removed parameter?**
   - List EVERY method/property that was accessed on that parameter
   - For each method, ask: "Is this functionality still available another way?"
   - Check the library documentation for what that object provided

3. **What was the removed parameter's lifecycle?**
   - Was it initialized before being passed?
   - Did it have defaults set?
   - Did it configure the receiving object?

4. **Critical Questions:**
   - Does the receiving function still have access to the same services?
   - Are there initialization steps that no longer happen?
   - Are there default values that are no longer set?

#### Example Investigation Path

If you see: `controller.NewUnmanaged(name, mgr, options)` → `controller.NewUnmanaged(name, options)`

**DO NOT STOP** at "the manager parameter was removed."

**CONTINUE INVESTIGATING:**

1. What did the manager object provide?
   - Client
   - Scheme
   - Logger
   - REST Mapper
   - Event Recorder
   - Metrics Registry
   - Configuration
   - Cache
   - Leader Election
   - Webhooks

2. For EACH of these services:
   - Is it still accessible to the controller?
   - How is it being initialized now?
   - What defaults were being set by the manager?
   - Are those defaults still being set?

3. Check the library source code:
   - Read the OLD implementation of the function
   - Read the NEW implementation of the function
   - What initialization code was removed?
   - What was the removed code doing?

### Phase 3: Trace Service Initialization Paths

For critical services (logging, metrics, configuration, error handling):

1. **Trace the Full Initialization Path**
   ```
   Application Start
   → Manager Creation
   → Controller Creation
   → Service Injection
   → Default Value Setting
   → Runtime Usage
   ```

2. **Verify Each Step Still Happens**
   - Where was the logger initialized in the old code?
   - Where is it initialized in the new code?
   - Are there any steps that were skipped?

3. **Check for Implicit Dependencies**
   - Did the old code implicitly set something?
   - Is that implicit behavior still present?
   - What happens if that initialization doesn't occur?

### Phase 4: Identify Silent Behavioral Changes

A silent breaking change means:
- Code compiles successfully
- Tests may pass (if they don't check the specific behavior)
- Runtime behavior changes in a way that affects production
- The change is not obvious from reading the diff

#### Categories of Silent Breaking Changes

1. **Observability Degradation**
   - Logging stops working or becomes less verbose
   - Metrics stop being emitted or change format
   - Traces are incomplete
   - Events are not recorded
   - Errors are swallowed instead of logged

2. **Performance Degradation**
   - Rate limiting is removed/changed
   - Concurrency limits change
   - Timeout values change
   - Cache behavior changes
   - Retry logic changes

3. **Configuration Changes**
   - Default values change
   - Configuration precedence changes
   - Environment variables are ignored
   - Feature flags behave differently

4. **Runtime Behavior Changes**
   - Error handling changes (fail-open vs fail-closed)
   - Reconciliation timing changes
   - Watch behavior changes
   - Cache sync behavior changes
   - Leader election behavior changes
   - Shutdown/cleanup behavior changes

5. **Security Implications**
   - Authentication/authorization checks removed
   - Validation removed
   - Sanitization removed
   - Permission requirements changed
   - TLS/encryption behavior changed

#### Investigation Checklist for EACH Changed Function

For every function call that changed in the PR, verify:

- [ ] Is logging still functional for this code path?
- [ ] Are metrics still being emitted?
- [ ] Are errors still being reported?
- [ ] Are events still being recorded?
- [ ] Is the configuration still being read?
- [ ] Are defaults still being applied?
- [ ] Is rate limiting still applied?
- [ ] Is timeout behavior the same?
- [ ] Is retry behavior the same?
- [ ] Is cleanup/shutdown behavior the same?
- [ ] Are security checks still in place?

### Phase 5: Examine Vendor Changes

Even though vendor changes are typically autogenerated, you must investigate them:

1. **Identify High-Risk Vendor Changes**
   - Changes to core controller/manager/client code
   - Changes to logging/metrics libraries
   - Changes to error handling
   - Changes to initialization code
   - Changes to default values

2. **Look for Deleted Code**
   - Was an entire file or function removed from vendor/?
   - What was that code doing?
   - Is anything in the codebase calling it?
   - What happens now that it's gone?

3. **Look for New Required Patterns**
   - Are there new types that require initialization?
   - Are there new interfaces that must be implemented?
   - Are there new patterns shown in examples?

### Phase 6: Cross-Reference with Upstream

1. **Find the Actual Pull Request**
   - Don't rely on release notes
   - Find the actual PR that made the change
   - Read the full discussion
   - Look for migration guides
   - Check for reported issues

2. **Search for Related Issues**
   - Search the dependency's GitHub issues for:
     - "breaking change"
     - "migration"
     - "upgrade"
     - The specific function/type that changed
   - Look for issues filed AFTER the release
   - These often reveal problems the maintainers didn't anticipate

3. **Check for Follow-up PRs**
   - Look at PRs merged after the one being analyzed
   - Do any of them fix issues introduced by this change?
   - What do those fixes tell you?

### Phase 7: Runtime Impact Analysis

For each change, analyze the runtime impact:

1. **What Happens at Startup?**
   - Does the service start successfully?
   - Are all controllers initialized?
   - Are all watches established?
   - Is logging visible during startup?
   - Are initial metrics emitted?

2. **What Happens During Normal Operation?**
   - Do reconciliation loops still work?
   - Are errors properly logged?
   - Are metrics updated?
   - Is resource usage reasonable?

3. **What Happens During Error Conditions?**
   - Are errors logged?
   - Are errors reported via events?
   - Does the controller recover?
   - Is there exponential backoff?

4. **What Happens During Shutdown?**
   - Are resources cleaned up?
   - Are graceful shutdown signals respected?
   - Are final metrics flushed?

### Phase 8: Test Coverage Analysis

1. **What Isn't Being Tested?**
   - Logging output
   - Metrics emission
   - Event recording
   - Error paths
   - Edge cases

2. **Would Tests Catch This?**
   - Do tests verify logging?
   - Do tests check metrics?
   - Do tests validate events?
   - Do tests cover error scenarios?

## Red Flags

Watch for these warning signs:

- ❌ **"This is just a library update"** - No, it's not. Every change has risk.
- ❌ **"The PR author just followed the compiler errors"** - That fixes compilation, not runtime behavior.
- ❌ **"Release notes say it's backward compatible"** - For the API maybe, but not for behavior.
- ❌ **"Tests pass"** - Tests only check what they're written to check.
- ❌ **"It compiles"** - Compilation proves syntax, not correctness.
- ❌ **"We can fix issues later"** - Some issues are invisible for months.

## Investigation Output Format

Your analysis should include:

### 1. Breaking Changes (Compilation-Breaking)
List all changes that break compilation and what was done to fix them.

### 2. Silent Breaking Changes (Runtime Behavior)
For EACH silent breaking change, document:
- **What changed**: Specific behavior that changed
- **Previous behavior**: How it worked before
- **New behavior**: How it works now (or doesn't work)
- **Impact**: What breaks or degrades at runtime
- **Detection**: How would you discover this in production
- **Risk level**: Critical/High/Medium/Low
- **Duration**: How long could this go unnoticed

### 3. Initialization Path Changes
Document any changes to how services are initialized:
- What was initialized automatically before
- What must be initialized manually now
- What defaults are no longer applied

### 4. Required Follow-up Actions
- Code changes needed
- Testing needed
- Documentation updates
- Monitoring/alerting updates

## Remember

Your job is not to approve or reject changes. Your job is to **uncover what the PR doesn't tell you**.

Every removed parameter was doing something. Every changed signature had a reason. Every vendor update changes behavior. Find those changes. Document them. Make them visible.
