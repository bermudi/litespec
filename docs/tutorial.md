# Tutorial: Your First Feature

A complete feature cycle: fuzzy idea → clear GH issue → durable spec → one unit at a time → review → merged branch → closed issue. We'll add rate limiting to an API.

## Setup

Install litespec ([Getting Started](getting-started.md)), then initialize:

```bash
litespec init
git status --porcelain  # must print nothing
git rev-parse HEAD      # this becomes Base:
git switch -c litespec/add-rate-limiting
```

`litespec-plan` in clear mode performs those checks itself — clean tree, `Base:`, dedicated branch, `Branch:` — and creates the labeled issue. If `gh` is unavailable it writes the same body to `specs/queues/add-rate-limiting.md` instead.

This tutorial follows the new-feature lane. Small fixes skip all of this: read the spec, edit, update the spec if the contract changed, done.

## Plan fuzzy

Invoke `litespec-plan` and describe the idea.

> **You:** Limit each IP to 100 requests per minute, return 429 with a `Retry-After` header.
>
> **Plan (fuzzy):** Reads the code, product, glossary. Asks two or three questions, one at a time. Writes no files.
>
> - Configurable limit, or fixed at 100?
> - In-memory only, or a shared store?
> - Per-IP or per-user?

Fuzzy is ephemeral. When you can answer "what demo proves this?" and "what Verify fails without it?", say ready and move to clear.

## Plan clear

Clear mode writes the issue body: ownership lines first, then proposal, design, and queue. Each unit is one boundary or failure policy with identified clauses, scenario mappings, risk accounting, and one `Verify:` — dry-run on the base tree before filing, so every Verify is known to fail without its outcome.

````markdown
Base: <full commit SHA printed before branch creation>
Branch: litespec/add-rate-limiting

## Proposal

Add rate limiting to prevent abuse and keep usage fair. Out of scope: distributed stores, per-user quotas.

## Design

In-memory sliding-window counter per IP behind a middleware. The middleware extracts the IP, increments the counter, rejects over-limit requests with 429 and `Retry-After`. Limit comes from `RATE_LIMIT_PER_MINUTE`, default 100.

## Sliding window counter
Boundary: process
Done means:
- [window] The counter counts requests in the last 60 seconds
Scenarios:
- [window] TestCounterWindow
Risk cases:
- timeout: N/A — in-memory count, no deadline
- cleanup: N/A — no temp state per request
- non-ENOENT errors: N/A — no filesystem lookup
- concurrency: [window]
- optional configured dependencies: N/A — counter has no optional services
Verify: `go test ./internal/ratelimit -run TestCounterWindow`
- [ ] pending

## Limiter enforces per-IP cap
Depends: Sliding window counter
Done means:
- [cap] 100 requests per minute per IP pass; the 101st returns 429 with `Retry-After`
Scenarios:
- [cap] TestLimiterEnforcement
Verify: `go test ./internal/ratelimit -run TestLimiterEnforcement`
- [ ] pending
````

Create it with `gh issue create --label litespec --body-file issue.md`. Prose-only units are banned — explanation rides with the unit whose behavior it describes.

## Draft the spec

Rate limiting outlives the issue, so it's load-bearing: plan drafts `specs/rate-limit/spec.md` alongside the issue. Contracts, not proposals — `SHALL`/`MUST`, `WHEN`/`THEN`:

```markdown
# rate-limit

## Requirements

### Requirement: Enforce per-IP request limit
The rate limiter MUST enforce 100 requests per minute per IP. Requests over the limit SHALL return HTTP 429 with a `Retry-After` header.

#### Scenario: Request under limit
- **WHEN** an IP makes 50 requests within one minute
- **THEN** all requests succeed with HTTP 200

#### Scenario: Request exceeds limit
- **WHEN** an IP makes 101 requests within one minute
- **THEN** the 101st returns HTTP 429 with `Retry-After: 60`

### Requirement: Configurable limit
The limit MUST be configurable via `RATE_LIMIT_PER_MINUTE`. It SHALL default to 100 when unset.

#### Scenario: Default limit
- **WHEN** `RATE_LIMIT_PER_MINUTE` is unset
- **THEN** the limit is 100 requests per minute
```

Run `litespec validate` — it reports the file and line for a missing `SHALL`/`MUST` or a scenario without `WHEN`/`THEN`.

## Grill me (optional)

Before building:

> **You:** grill-me
>
> **Plan:** What if the counter map grows unbounded? What if `X-Forwarded-For` holds multiple IPs? What if the service restarts?

Grilling may tighten the spec or record a durable decision. Cheap now, expensive later.

## Build one unit at a time

> **You:** build unit 1
>
> **Build:** Reads the issue, the spec, product, glossary. On the issue's branch, from a clean tree, runs the exact Verify — it fails because the counter is absent. That clean commit is pre. (If the unit introduces its own test, build commits just the verifier first and uses that as pre.) Then implements the sliding window, commits without amending, and re-runs the same Verify at the final clean commit — green. That commit is post.

Then the receipt. Build runs `litespec digest` for the unit's contract digest and `litespec receipt` to assemble the evidence — exact command, digest, pre/post SHAs and statuses, both raw outputs unedited, scope lines — into numbered comment files, posts them, and ticks the box with `litespec issue check` (exactly one flip, ownership lines untouched). Then it stops. Re-invoke for unit 2.

Never amend pre or any implementation commit; fixes go in new commits.

## Review

> **You:** review
>
> **Review:** Reads the issue body first, screens every local path before touching it, then fetches comments. Replays the exact Verify in detached throwaway worktrees at pre, post, and `HEAD` — removing each even on failure, never checking out evidence SHAs in your tree. Then the adversarial pass: does the 101st request really return 429? Does the window reset? Concurrent access? Empty headers?

A passing Verify proves only its scope, so review probes beyond the receipt. Findings route in order: suggestions ride the small-fix lane; unit violations rebuild the unit via build (at most twice per contract — the third routes to plan to reshape it); in-scope findings outside units become a direct fix or a new unit on this issue; out-of-scope findings route without blocking.

## Merge, then close

When every box is ticked, every rebuild request and amendment resolved, and review returns `PASS`:

```bash
gh pr create --head litespec/add-rate-limiting
# after merge:
gh issue close 42
```

Merge first, then close — a closed issue leaves no work stranded on a branch. The queue is gone; `specs/rate-limit/spec.md` remains as the durable truth.

## What you did

1. `litespec-plan` grilled the idea, then wrote `Base:`/`Branch:` plus proposal, design, and queue into the issue.
2. `litespec-plan` drafted the load-bearing spec with `SHALL`/`MUST` and `WHEN`/`THEN`.
3. `litespec-build` implemented one unit at a time — red at pre, green at post, receipt, tick, stop.
4. `litespec validate` confirmed structure (it never claims the code is correct).
5. `litespec-review` replayed the evidence and probed the behavior.
6. You merged the branch, then closed the issue.

## What's next

- [Workflow](workflow.md) — unit shape, evidence protocol, and routing in full
- [Concepts](concepts.md) — what makes a good spec
- [CLI Reference](cli-reference.md) — every command and flag
