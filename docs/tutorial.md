# Tutorial: Your First Feature

This tutorial walks through a complete v2 feature cycle. We'll add rate limiting to an API: plan a fuzzy idea, write a clear GH issue, draft a durable spec, build one unit at a time, review, and close the issue.

## Setting up

You already have litespec installed. Initialize the project:

```bash
$ litespec init --tools claude
Created specs/ directory structure
Generated .agents/skills/
Generated adapter commands for: claude
Project initialized.
```

Start the queue issue from a clean tree and dedicated branch:

```bash
git status --porcelain  # must print nothing
git rev-parse HEAD      # record this as Base:
git switch -c litespec/add-rate-limiting
```

`litespec-plan` in `clear` mode performs those checks, records `Base:` and `Branch:`, and creates the labeled issue. If `gh` is unavailable, it writes the same body to `specs/queues/add-rate-limiting.md`.

## Two lanes, one workflow

v2 has two lanes:

- **Small fix** — zero ceremony. Read `specs/product.md`, the relevant `specs/<feature>/spec.md`, and decisions/glossary; edit code; update the spec if the contract changed. No GH issue required.
- **New feature** — `plan[fuzzy]` → `plan[clear]` (GH issue) → `grill-me` → `build` one unit at a time → `review` → close the issue.

This tutorial follows the new-feature lane.

## Plan (fuzzy)

Invoke the `litespec-plan` skill and describe the idea.

> **You:** I want to add rate limiting to the API. We should limit each IP to 100 requests per minute and return 429 with a `Retry-After` header.
>
> **litespec-plan (fuzzy mode):** Reads the code, `specs/product.md`, and the glossary. Asks 2–3 clarifying questions. Writes no files.
>
> Example questions:
> - Should the limit be configurable?
> - In-memory only, or do we need a shared store?
> - Per-user or IP-based?

Fuzzy mode is ephemeral. When the scope is clear, move to `clear` mode.

## Plan (clear)

`litespec-plan` (clear mode) writes ownership metadata followed by the proposal, design, and queue. Each unit is an `## <outcome>` with `Done means:` and a `Verify:` that must fail without the outcome.

````markdown
Base: <full commit ID printed before branch creation>
Branch: litespec/add-rate-limiting

## Proposal

Add rate limiting to the API to prevent abuse and ensure fair usage.

## Design

Use an in-memory sliding window counter per IP. A middleware extracts the IP, increments the counter, and rejects requests over the limit with HTTP 429 and a `Retry-After` header. The limit is configurable via `RATE_LIMIT_PER_MINUTE` and defaults to 100.

## Unit 1: Sliding window counter
Done means: `internal/ratelimit/counter.go` exists and counts requests in the last 60 seconds.

Verify:
```bash
go test ./internal/ratelimit -run TestCounterWindow
```

- [ ] pending

## Unit 2: Limiter enforces per-IP cap
Done means: `internal/ratelimit/limiter.go` exists, allows 100 requests per minute per IP, and rejects the 101st with 429.

Verify:
```bash
go test ./internal/ratelimit -run TestLimiterEnforcement
```

- [ ] pending

## Unit 3: Middleware integration
Done means: `cmd/api/middleware.go` applies the limiter and sets `Retry-After`.

Verify:
```bash
go test ./cmd/api -run TestMiddleware
```

- [ ] pending
````

Create or update the issue with `gh issue create` or `gh issue edit 42 --body-file issue.md`.

## Draft the spec

For load-bearing features, `litespec-plan` also drafts a durable spec at `specs/<feature>/spec.md`. It is a contract, not a proposal, so it uses `SHALL`/`MUST` and `WHEN`/`THEN`.

`specs/rate-limit/spec.md`:

```markdown
# rate-limit

## Purpose

Rate limiting protects the API from abuse and ensures fair usage by limiting the number of requests per IP address over a time window.

## Requirements

### Requirement: Enforce per-IP request limit
The rate limiter MUST enforce a maximum of 100 requests per minute per IP address. Requests exceeding the limit SHALL return HTTP 429 with a `Retry-After` header.

#### Scenario: Request under limit
- **WHEN** an IP address makes 50 requests within one minute
- **THEN** all requests succeed with HTTP 200

#### Scenario: Request exceeds limit
- **WHEN** an IP address makes 101 requests within one minute
- **THEN** the 101st request returns HTTP 429 with a `Retry-After: 60` header

### Requirement: Configurable limit
The rate limit MUST be configurable via the `RATE_LIMIT_PER_MINUTE` environment variable. The default value SHALL be 100 when the variable is unset.

#### Scenario: Default limit
- **WHEN** `RATE_LIMIT_PER_MINUTE` is unset
- **THEN** the limit defaults to 100 requests per minute
```

Every requirement body contains `SHALL` or `MUST`. Each load-bearing requirement has at least one named scenario with `WHEN` and `THEN`.

## Grill me (optional)

Before building, ask `litespec-plan` to stress-test the plan.

> **You:** grill-me
>
> **litespec-plan:** Questions the design. What if the counter map grows unbounded? What if `X-Forwarded-For` contains multiple IPs? What if the service restarts? It may add a durable decision to `specs/decisions/` or tighten the spec.

This is optional, but it catches gaps before code.

## Build one unit at a time

Invoke `litespec-build` for **Unit 1**.

> **You:** build unit 1
>
> **litespec-build:** Reads the GH issue, `specs/rate-limit/spec.md`, `specs/product.md`, and `specs/glossary.md`. Creates `internal/ratelimit/counter.go`, implements the sliding window, and adds a test. Runs the Verify:
>
> ```bash
> go test ./internal/ratelimit -run TestCounterWindow
> ```
>
> When it passes, it updates the GH issue checkbox from `- [ ] pending` to `- [x] done` and commits:
>
> ```bash
> git commit -m "ratelimit: add sliding window counter"
> ```
>
> It then stops. One unit per session.

Repeat for Unit 2 and Unit 3. Each unit is demo-able and has a Verify that fails without it.

## Validate

Run `litespec validate` as you go:

```bash
$ litespec validate
ok: 1 capability, 2 requirements, 3 scenarios
```

If a requirement is missing `SHALL`/`MUST` or a scenario is missing `WHEN`/`THEN`, `validate` reports the file and line.

## View the dashboard

`litespec view` shows the spec and the open issue:

```text

Litespec Dashboard

════════════════════════════════════════════════════════════
Product:
  specs/product.md — # Product
  product: mental models + flows

Summary:
  ● Specifications: 1 specs, 2 requirements
  ● GH Issues: 1 open

Specifications
────────────────────────────────────────────────────────────
  ▪ rate-limit                     2 requirements  (specs/rate-limit/spec.md)

GH Issues (open)
────────────────────────────────────────────────────────────
  #42     Add rate limiting                        https://github.com/your-org/your-project/issues/42

════════════════════════════════════════════════════════════

```

## Review

When all units are done, invoke `litespec-review`. It reads only the remote issue first, then screens every local queue, contract, implementation, and reference path before reading it. Unsafe paths or parent components stop review.

> **You:** review
>
> **litespec-review:** Adversarial review of the GH issue + `specs/rate-limit/spec.md` vs the implementation. It checks:
> - Does the code enforce 100 requests per minute per IP?
> - Does the 101st request return 429 with `Retry-After`?
> - Does `RATE_LIMIT_PER_MINUTE` default to 100?
> - Are edge cases (empty header, window reset, concurrent access) tested?

Routing is ordered:

1. Suggestions are non-blocking.
2. Unit violations, including WARNINGs, reopen that unit for `litespec-build`.
3. CRITICAL/WARNING findings inside issue scope but outside units block as a direct fix or a new unit on this issue.
4. Findings outside issue scope route without blocking; non-trivial work gets its own later `plan[clear]` and branch.

## Close the issue

When every unit is checked and review returns `PASS`, close the GH issue:

```bash
gh issue close 42
```

The queue is closed. The durable spec remains in `specs/rate-limit/spec.md`.

## Summary

You completed a v2 feature cycle:

1. `litespec-plan` clarified the idea, started from a clean tree, created the isolated branch, and wrote `Base:`/`Branch:` into the queue issue.
2. `litespec-plan` drafted `specs/rate-limit/spec.md` with `SHALL`/`MUST` and `WHEN`/`THEN`.
3. `litespec-build` implemented one unit at a time, satisfying `Done means:` and `Verify` for each.
4. `litespec validate` confirmed the spec format.
5. `litespec-review` checked the implementation against the issue and spec.
6. `gh issue close 42` closed the queue.

The spec is the durable source of truth. The issue is disposable.

## What's next

- [Workflow](workflow.md) — the two lanes in detail
- [Concepts](concepts.md) — what makes a good spec
- [CLI Reference](cli-reference.md) — command reference
