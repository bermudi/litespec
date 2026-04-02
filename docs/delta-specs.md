# Delta Specs

Delta specs are the mechanism for proposing changes to the canonical specifications. They live in `specs/changes/<name>/specs/<capability>/spec.md` and define what will change when the change is archived.

Delta specs matter because they allow multiple changes to be developed in parallel without conflict. Each change proposes modifications using declarative markers (`ADDED`, `MODIFIED`, `REMOVED`, `RENAMED`), and these markers are merged into the canonical specs at archive time in a strict, predictable order.

## Delta Markers

There are four delta markers, each expressed as an H2 section:

### ADDED

Adds new requirements to the canonical spec. Content is appended to the end of the `## Requirements` section.

```markdown
## ADDED Requirements

### Requirement: <name>
<body text — must contain SHALL or MUST>

#### Scenario: <short name>
- **WHEN** <condition>
- **THEN** <expected outcome>
```

### MODIFIED

Updates existing requirements in the canonical spec. Requirements are matched by name (the `### Requirement: <name>` header) and replaced entirely.

```markdown
## MODIFIED Requirements

### Requirement: <existing-name>
<updated body text>

#### Scenario: <updated scenario>
- **WHEN** <updated condition>
- **THEN** <updated outcome>
```

### REMOVED

Deletes requirements from the canonical spec. Only the requirement name is needed — no body or scenarios.

```markdown
## REMOVED Requirements

### Requirement: <name-to-remove>
```

### RENAMED

Changes requirement names while preserving content and scenarios. Useful for clarification or reorganization.

```markdown
## RENAMED Requirements

### Requirement: <old-name> → <new-name>
```

## Merge Order

Delta operations are applied in strict order at archive time:

1. **RENAMED** — establishes correct headers before other operations
2. **REMOVED** — eliminates requirements that should not exist
3. **MODIFIED** — updates the remaining requirements
4. **ADDED** — appends new requirements

This order prevents conflicts and ensures each operation sees the correct state. For example, RENAMED runs first so that subsequent MODIFIED operations can reference the new names. REMOVED runs before MODIFIED so you don't waste effort updating a requirement that will be deleted anyway.

## Dangling Delta Detection

`litespec validate` catches **dangling deltas** — MODIFIED or REMOVED operations that reference requirements that don't exist in the target canonical spec. This is an improvement over OpenSpec, which only fails on these at archive time.

If your delta spec tries to modify or remove a requirement that isn't in `specs/canon/<capability>/spec.md`, validation fails immediately with a clear error:

```
error: dangling delta in validate/spec.md: MODIFIED requirement "Nonexistent Requirement" does not exist in canonical spec
```

This catches mistakes early, before you've invested time in implementation.

## Canonical Spec Format

Canonical specs live in `specs/canon/<capability>/spec.md` and follow this structure:

```markdown
# <capability>

## Purpose               ← optional prose

## Requirements          ← required wrapper for all requirements

### Requirement: <name>
<body text — must contain SHALL or MUST>

#### Scenario: <name>
- **WHEN** <condition>
- **THEN** <expected outcome>
```

- `## Purpose` is optional — a section for explaining what this capability is about
- `## Requirements` is required — all `### Requirement:` blocks must live inside it
- No other H2 sections are permitted between the H1 capability heading and `## Requirements`
- Requirement bodies must contain `SHALL` or `MUST` (enforced by validation)
- Scenarios use WHEN/THEN format and describe expected behavior

## Complete Before/After Example

Consider an `auth` capability that evolves to add token-based authentication.

### Before: Canonical Spec

**`specs/canon/auth/spec.md`**

```markdown
# auth

## Purpose

Authentication mechanisms for protecting API endpoints.

## Requirements

### Requirement: Basic Auth
The system SHALL support HTTP Basic Authentication using a username and password. Credentials MUST be validated against the user database.

#### Scenario: Valid credentials
- **WHEN** a request includes `Authorization: Basic base64(username:password)` with valid credentials
- **THEN** the request is authenticated and processed

#### Scenario: Invalid credentials
- **WHEN** a request includes invalid credentials
- **THEN** the system responds with 401 Unauthorized

### Requirement: Password hashing
User passwords MUST be hashed using bcrypt with a work factor of at least 12 before storage in the database.
```

### Delta Spec

**`specs/changes/add-token-auth/specs/auth/spec.md`**

```markdown
## MODIFIED Requirements

### Requirement: Basic Auth
The system SHALL support HTTP Basic Authentication using a username and password. Credentials MUST be validated against the user database. Basic Auth MUST be disabled by default in production configurations.

#### Scenario: Valid credentials
- **WHEN** a request includes `Authorization: Basic base64(username:password)` with valid credentials and Basic Auth is enabled
- **THEN** the request is authenticated and processed

#### Scenario: Basic Auth disabled in production
- **WHEN** a Basic Auth request is received in production configuration
- **THEN** the system responds with 403 Forbidden

## ADDED Requirements

### Requirement: JWT Token Authentication
The system SHALL support JSON Web Token (JWT) authentication for stateless API access. Tokens MUST be signed using RS256 and include `exp` (expiration) and `sub` (subject) claims.

#### Scenario: Generate token
- **WHEN** a user provides valid credentials
- **THEN** the system returns a signed JWT with 1-hour expiration

#### Scenario: Validate token
- **WHEN** a request includes `Authorization: Bearer <jwt>` with a valid, unexpired token
- **THEN** the request is authenticated and the user ID is extracted from the `sub` claim

#### Scenario: Expired token
- **WHEN** a request includes an expired JWT
- **THEN** the system responds with 401 Unauthorized with error code `token_expired`
```

### After: Canonical Spec (merged)

**`specs/canon/auth/spec.md`** after `litespec archive add-token-auth`

```markdown
# auth

## Purpose

Authentication mechanisms for protecting API endpoints.

## Requirements

### Requirement: Basic Auth
The system SHALL support HTTP Basic Authentication using a username and password. Credentials MUST be validated against the user database. Basic Auth MUST be disabled by default in production configurations.

#### Scenario: Valid credentials
- **WHEN** a request includes `Authorization: Basic base64(username:password)` with valid credentials and Basic Auth is enabled
- **THEN** the request is authenticated and processed

#### Scenario: Basic Auth disabled in production
- **WHEN** a Basic Auth request is received in production configuration
- **THEN** the system responds with 403 Forbidden

### Requirement: Password hashing
User passwords MUST be hashed using bcrypt with a work factor of at least 12 before storage in the database.

### Requirement: JWT Token Authentication
The system SHALL support JSON Web Token (JWT) authentication for stateless API access. Tokens MUST be signed using RS256 and include `exp` (expiration) and `sub` (subject) claims.

#### Scenario: Generate token
- **WHEN** a user provides valid credentials
- **THEN** the system returns a signed JWT with 1-hour expiration

#### Scenario: Validate token
- **WHEN** a request includes `Authorization: Bearer <jwt>` with a valid, unexpired token
- **THEN** the request is authenticated and the user ID is extracted from the `sub` claim

#### Scenario: Expired token
- **WHEN** a request includes an expired JWT
- **THEN** the system responds with 401 Unauthorized with error code `token_expired`
```

### What happened?

1. **RENAMED** — none in this example
2. **REMOVED** — none in this example
3. **MODIFIED** — the `Basic Auth` requirement was completely replaced with an updated version (including a new scenario about production configuration)
4. **ADDED** — the `JWT Token Authentication` requirement was appended at the end

The original "Invalid credentials" scenario from `Basic Auth` disappeared because MODIFIED replaces the entire requirement, including its scenarios.

## Rules and Best Practices

### Validation Rules

- **ADDED and MODIFIED requirements must have at least one scenario** — a requirement without a scenario is untestable
- **Requirement bodies must contain `SHALL` or `MUST`** — this enforces precise, verifiable language
- **REMOVED requirements are name-only** — no body text, no scenarios
- **RENAMED requirements preserve content** — scenarios move with the requirement under the new name
- **MODIFIED replaces entire requirement** — all scenarios in the delta spec become the requirement's scenarios

### Best Practices

- **Run `litespec validate` frequently** — catch dangling deltas before you go too far
- **Use MODIFIED for meaningful updates** — don't change just for the sake of it; preserve scenarios when the core behavior is stable
- **Use REMOVED explicitly** — don't rely on MODIFIED to empty a requirement; removing it makes the intent clear
- **Consider the merge order** — if you're renaming and then modifying, both can exist in the same delta spec (RENAMED runs first)
- **Test scenarios should be atomic** — each WHEN/THEN pair should test one behavior
- **Keep requirements focused** — if a requirement is doing too much, split it into multiple requirements with MODIFIED/ADDED

### Common Pitfalls

- **Dangling delta**: Trying to modify a requirement that doesn't exist. Fix by checking the canonical spec first.
- **Missing scenarios**: Forgetting to add scenarios to ADDED or MODIFIED requirements. Validation catches this.
- **SHALL/MUST requirement**: Writing requirement bodies without SHALL/MUST. Validation catches this.
- **Partial modification**: Thinking MODIFIED only changes the body. It replaces the requirement entirely — include all scenarios you want to keep.
