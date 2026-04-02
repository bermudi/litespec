# Tutorial: Your First Change

This tutorial walks you through a complete litespec workflow from start to finish. We'll implement a simple feature — adding rate limiting — and see how litespec structures work.

## Setting Up

First, initialize a litespec project:

```bash
$ litespec init
Created specs/ directory structure
Generated .agents/skills/
Project initialized.
```

This creates foundational structure:
- `specs/canon/` — where accepted capabilities live
- `specs/changes/` — where work in progress happens
- `.agents/skills/` — AI skill definitions

## Creating a Change

Start a new change for our feature:

```bash
$ litespec new add-rate-limiting
/home/your/project/specs/changes/add-rate-limiting
```

litespec created a change directory with minimal metadata:

```bash
$ cat specs/changes/add-rate-limiting/.litespec.yaml
schema: spec-driven
created: "2026-04-02T21:21:14Z"
```

Check the status:

```bash
$ litespec status add-rate-limiting
Change: add-rate-limiting
Created: 2026-04-02 21:21:14

  proposal     READY      Why and what — the motivation, scope, and approach for this change
  specs        BLOCKED    Delta specifications — ADDED/MODIFIED/REMOVED/RENAMED requirements
  design       BLOCKED    How — the technical approach, architecture decisions, data flow, file changes
  tasks        BLOCKED    What to do — the phased implementation checklist
```

The proposal is READY (has no dependencies). Everything else is BLOCKED waiting for it.

## The Propose Workflow

Now you invoke your AI agent with the `litespec-propose` skill. The conversation might go like this:

> **You:** I want to add rate limiting to the API. We should limit requests to 100 per minute per IP address.
>
> **AI (propose mode):** I'll create a change proposal for adding rate limiting. Let me check if it exists and create the artifacts.

The AI runs `litespec status add-rate-limiting --json` to see what's needed:

```json
{
  "changeName": "add-rate-limiting",
  "schemaName": "spec-driven",
  "isComplete": false,
  "artifacts": [
    {
      "id": "proposal",
      "outputPath": "proposal.md",
      "status": "ready"
    },
    ...
  ]
}
```

It starts with the proposal artifact. First, it gets instructions:

```bash
$ litespec instructions proposal
```

Then it creates `specs/changes/add-rate-limiting/proposal.md`:

```markdown
# add-rate-limiting

## Motivation

The API currently has no rate limiting, which makes it vulnerable to abuse and DoS attacks. Adding rate limiting protects the service and ensures fair usage.

## Scope

- Add a `rate-limit` capability that limits requests to 100 per minute per IP address
- Track request counts in memory with a sliding window algorithm
- Return HTTP 429 (Too Many Requests) when limit is exceeded
- Include a `Retry-After` header indicating when to retry
- Make rate limit configurable via environment variable

## Non-Goals

- No persistent storage for rate limit counters (memory-only)
- No per-user rate limiting (IP-based only)
- No complex token bucket or leaky bucket algorithms
- No distributed rate limiting across instances
```

The AI then moves on to specs. It reads the proposal for context and gets instructions:

```bash
$ litespec instructions specs
```

Since this is a new capability (not modifying an existing one), it creates a new spec under `specs/changes/add-rate-limiting/specs/rate-limit/spec.md`:

```markdown
# rate-limit

## ADDED Requirements

### Requirement: Rate Limit Enforcement
The rate limiter MUST track incoming requests per IP address and enforce a maximum of 100 requests per minute. Requests exceeding this limit SHALL return HTTP 429 with a `Retry-After` header.

#### Scenario: Request under limit
- **WHEN** an IP address makes 50 requests within one minute
- **THEN** all requests succeed with HTTP 200

#### Scenario: Request exceeds limit
- **WHEN** an IP address makes 101 requests within one minute
- **THEN** 101st request returns HTTP 429 with `Retry-After: 60`

#### Scenario: Sliding window expiration
- **WHEN** 60 seconds pass after the first request
- **THEN** the request count for that IP resets

### Requirement: Configurable Limit
The rate limit MUST be configurable via `RATE_LIMIT_PER_MINUTE` environment variable. The default value SHALL be 100 if not specified.

#### Scenario: Default limit
- **WHEN** no `RATE_LIMIT_PER_MINUTE` environment variable is set
- **THEN** limit defaults to 100 requests per minute

#### Scenario: Custom limit
- **WHEN** `RATE_LIMIT_PER_MINUTE` is set to 200
- **THEN** limit is 200 requests per minute

### Requirement: Retry-After Header
The HTTP 429 response MUST include a `Retry-After` header with the number of seconds until the next request will be allowed.

#### Scenario: Retry-After value
- **WHEN** a request is rate-limited
- **THEN** response includes `Retry-After: <seconds-until-next-window>`
```

Now, AI creates the design artifact. It reads the proposal and specs for context:

```bash
$ litespec instructions design
```

And creates `specs/changes/add-rate-limiting/design.md`:

```markdown
## Architecture

The rate limiter is an HTTP middleware that sits between the router and request handlers:

```
HTTP Request → RateLimiter Middleware → Handlers → Response
                    ↓
               In-Memory Counter Map
```

A sliding window counter tracks requests per IP. Each request increments the counter for its IP; the counter resets when the window slides.

```
internal/ratelimit/
  ├── limiter.go        — Limiter struct and core logic
  └── counter.go       — Sliding window counter implementation
cmd/api/
  ├── middleware.go    — HTTP middleware wrapper (modified)
  └── main.go          — Wire up middleware (modified)
```

## Decisions

### Sliding window over token bucket
Chosen sliding window because it's simple to implement and provides smooth rate limiting without the complexity of token bucket refill logic. Trade-off: slightly more memory than fixed window, but more accurate enforcement.

### In-memory storage
No persistent storage because rate limiting is a temporary protection mechanism. If the service restarts, counters reset. This is acceptable for our use case and keeps the implementation simple. Trade-off: counters don't survive restarts, but we avoid Redis or database dependency.

### Environment variable configuration
No config file — keeps things simple and follows litespec's convention-over-configuration philosophy. Trade-off: requires restart to change limits, but deployment tooling (k8s, systemd) can handle env var updates.

## File Changes

### `internal/ratelimit/limiter.go` (new)
- `Limiter` struct with `limit int` and `counters map[string]*Counter`
- `Allow(ip string) bool` method — checks if request from IP is allowed
- `NewLimiter(limit int) *Limiter` constructor
- Maps to spec requirement "Rate Limit Enforcement"

### `internal/ratelimit/counter.go` (new)
- `Counter` struct with `timestamps []time.Time`
- `Increment(window time.Duration)` — adds current timestamp, prunes old ones
- `Count(window time.Duration) int` — returns count within window
- Maps to spec requirement "Sliding window expiration"

### `cmd/api/middleware.go` (modified)
- Add `RateLimitMiddleware(limiter *ratelimit.Limiter) func(http.Handler) http.Handler`
- Extract IP from `X-Forwarded-For` or `RemoteAddr`
- Set `Retry-After` header on HTTP 429 responses
- Maps to spec requirement "Retry-After Header"

### `cmd/api/main.go` (modified)
- Parse `RATE_LIMIT_PER_MINUTE` env var, default to 100
- Instantiate `Limiter` with configured limit
- Add middleware to router
- Maps to spec requirement "Configurable Limit"
```

Finally, AI creates the tasks artifact, reading all previous artifacts:

```bash
$ litespec instructions tasks
```

And creates `specs/changes/add-rate-limiting/tasks.md`:

```markdown
## Phase 1: Core Rate Limiting Logic

- [ ] Create `internal/ratelimit` package directory
- [ ] Implement `Counter` struct in `internal/ratelimit/counter.go` with `Increment()` and `Count()` methods using sliding window
- [ ] Implement `Limiter` struct in `internal/ratelimit/limiter.go` with `Allow()` method that tracks per-IP counters
- [ ] Add unit tests for `Counter` sliding window behavior
- [ ] Add unit tests for `Limiter` enforcement under/at/over limit

## Phase 2: HTTP Middleware Integration

- [ ] Add `RateLimitMiddleware()` function in `cmd/api/middleware.go` that wraps handlers
- [ ] Implement IP extraction from `X-Forwarded-For` or `RemoteAddr`
- [ ] Return HTTP 429 with `Retry-After` header when limited
- [ ] Wire up middleware in `cmd/api/main.go` between router and handlers
- [ ] Add `RATE_LIMIT_PER_MINUTE` env var parsing with default of 100
- [ ] Add integration tests for middleware behavior

## Phase 3: Verification and Polish

- [ ] Run full test suite with `go test ./...`
- [ ] Verify `go vet ./...` passes
- [ ] Manual test: verify rate limiting works with `curl`
- [ ] Manual test: verify `Retry-After` header is present on 429 responses
- [ ] Update `DESIGN.md` with rate-limit capability in capabilities table
```

## Validation

Now validate the change:

```bash
$ litespec validate add-rate-limiting
Validation passed.
```

The AI wrote valid artifacts:
- Delta spec has correct ADDED markers
- Every requirement has scenarios
- Body text contains SHALL/MUST
- All required artifacts exist

Check status again:

```bash
$ litespec status add-rate-limiting
Change: add-rate-limiting
Created: 2026-04-02 21:21:14

  proposal     DONE       Why and what — the motivation, scope, and approach for this change
  specs        DONE       Delta specifications — ADDED/MODIFIED/REMOVED/RENAMED requirements
  design       DONE       How — the technical approach, architecture decisions, data flow, file changes
  tasks        DONE       What to do — the phased implementation checklist
```

All artifacts are DONE. The change is ready for implementation.

## Applying (Implementation)

Invoke your AI agent with the `litespec-apply` skill. The AI reads all artifacts and implements Phase 1:

```bash
# AI applies Phase 1 tasks
# ... creates internal/ratelimit package, implements Counter and Limiter, adds tests ...
```

After Phase 1, AI marks tasks as complete in `tasks.md`:

```markdown
## Phase 1: Core Rate Limiting Logic

- [x] Create `internal/ratelimit` package directory
- [x] Implement `Counter` struct in `internal/ratelimit/counter.go` with `Increment()` and `Count()` methods using sliding window
- [x] Implement `Limiter` struct in `internal/ratelimit/limiter.go` with `Allow()` method that tracks per-IP counters
- [x] Add unit tests for `Counter` sliding window behavior
- [x] Add unit tests for `Limiter` enforcement under/at/over limit

## Phase 2: HTTP Middleware Integration

- [ ] Add `RateLimitMiddleware()` function in `cmd/api/middleware.go` ...
```

It commits with message `phase 1: Core Rate Limiting Logic` and stops. One phase per session. Re-invoke for Phase 2, then Phase 3.

## Verification

After implementation, run verify to check code against specs:

```bash
$ litespec verify add-rate-limiting
```

The AI reads all artifacts and the implemented code, comparing:
- Every spec requirement maps to concrete implementation
- Every scenario is handled
- Design decisions are followed

It reports:
```
### Review Mode
Implementation Review: 10/14 tasks checked

### Scorecard
| Dimension     | Pass | Fail | Not Evaluated |
|---------------|------|------|---------------|
| Completeness  | ✓    |      |              |
| Correctness   | ✓    |      |              |
| Coherence     | ✓    |      |              |
```

No issues. The implementation matches the specs.

## Archiving

When all tasks are done, archive the change:

```bash
$ litespec archive add-rate-limiting
Updated spec: rate-limit
Change "add-rate-limiting" archived successfully.
```

litespec performed these steps:

1. **Validated** the change (all artifacts valid, no errors)
2. **Checked** that all tasks were complete
3. **Merged** delta specs into `specs/canon/rate-limit/spec.md`
4. **Stripped** the `specs/` subtree from the change directory
5. **Moved** the change to `specs/changes/archive/2026-04-02-add-rate-limiting/`

The canonical spec now contains our new capability:

```bash
$ cat specs/canon/rate-limit/spec.md
# rate-limit

## Purpose

Rate limiting protects the API from abuse and ensures fair usage by limiting the number of requests per IP address over a time window.

## Requirements

### Requirement: Rate Limit Enforcement
The rate limiter MUST track incoming requests per IP address and enforce a maximum of 100 requests per minute. Requests exceeding this limit SHALL return HTTP 429 with a `Retry-After` header.

#### Scenario: Request under limit
- **WHEN** an IP address makes 50 requests within one minute
- **THEN** all requests succeed with HTTP 200

...
```

The archived change retains only planning artifacts:

```bash
$ ls specs/changes/archive/2026-04-02-add-rate-limiting/
.litespec.yaml  design.md  proposal.md  tasks.md
```

No `specs/` directory — it's been merged into the source of truth.

## Summary

You've completed your first litespec change:

1. **Init** — Set up the project structure
2. **New** — Created a change directory
3. **Propose** — AI created all planning artifacts (proposal, specs, design, tasks)
4. **Validate** — Confirmed artifacts are well-formed
5. **Apply** — Implemented the change phase by phase
6. **Verify** — Reviewed code against specs
7. **Archive** — Merged delta specs into canon, moved change to archive

The spec is now the single source of truth for the rate limiting capability. Future changes can reference it, modify it, or depend on it.

## What's Next

- Try `litespec explore` to brainstorm your next feature
- Use `litespec adopt` to spec existing code that lacks documentation
- Check `litespec list --specs` to see all capabilities
- Read `concepts.md` for the philosophy behind spec-driven development
