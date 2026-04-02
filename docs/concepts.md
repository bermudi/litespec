# Concepts

## What a Spec IS (and ISN'T)

A spec in litespec is not a design document. It's not a technical blueprint. It's not an implementation guide.

A spec is a **contract** between what you want and how you'll know when you've got it.

**A spec IS:**

- A list of behavioral requirements written as testable statements
- Each requirement backed by concrete scenarios that describe expected behavior
- The source of truth for what "done" means for a capability
- Something that survives implementation decisions and architectural changes
- Written in SHALL/MUST language for clarity and testability

**A spec ISN'T:**

- Code or pseudocode
- Implementation details (database schema, function signatures, etc.)
- Prose describing how something works internally
- A list of tasks or to-do items
- A user-facing manual or tutorial

Here's a requirement from litespec's own validate command:

```markdown
### Requirement: JSON Output for Validate

The `litespec validate` command MUST support a `--json` flag that returns structured JSON output containing a `valid` boolean, `errors` array, and `warnings` array. Each issue MUST include `severity`, `message`, and `file` fields.

#### Scenario: Validate single change with JSON flag
- **WHEN** `litespec validate <change-name> --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields
```

This spec says nothing about *how* to implement JSON serialization. It doesn't mention structs, marshaling, or stdout writing. It only describes the observable behavior from the outside.

## Why Spec-Driven Development Works

Writing specs before code forces you to think about **what** you're building before worrying about **how**. This ordering matters for three reasons:

### 1. You catch the wrong problem early

If you skip specs and jump straight to code, you might build something beautifully that solves the wrong problem. Specs surface these mismatches before you've committed to an implementation.

A grilling session in litespec's workflow exists specifically for this — relentless Q&A that resolves every branch of the design tree before a single artifact is written to disk.

### 2. You have a clear test for completion

Specs give you an unambiguous answer to "are we done?" When the code satisfies all scenarios in the spec, you're done. No more vague "it feels ready" or "I think that's everything."

### 3. You can change implementation without changing goals

Because specs are behavioral, not structural, you can refactor, optimize, or even completely rewrite the implementation and still know whether the new version satisfies the same contract.

## What Makes a Good Requirement

A good requirement is **testable** and **unambiguous**. Here's how litespec enforces this:

### SHALL and MUST

Every requirement body text must contain SHALL or MUST. This isn't pedantry — it's the difference between a wish and a contract.

**Bad:**
```markdown
### Requirement: The validate command should support JSON
```

"Should" is soft. Does it mean "nice to have" or "non-negotiable"? How would you test failure?

**Good:**
```markdown
### Requirement: JSON Output for Validate

The `litespec validate` command MUST support a `--json` flag that returns structured JSON output containing a `valid` boolean, `errors` array, and `warnings` array.
```

Now you know exactly what success looks like.

### Scenarios describe behavior, not implementation

Scenarios use WHEN/THEN format to describe expected behavior in specific situations.

**Bad:**
```markdown
#### Scenario: JSON output
- **WHEN** json flag is set
- **THEN** serialize the ValidationResult struct to JSON and write to stdout
```

This describes implementation, not behavior. If you change how you represent validation results internally, the spec breaks.

**Good:**
```markdown
#### Scenario: Validate single change with JSON flag
- **WHEN** `litespec validate <change-name> --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields
```

This describes observable behavior from the command line. Implementation can change; the spec remains valid.

### One clear responsibility

Each requirement should address one coherent behavior.

**Bad:**
```markdown
### Requirement: Validate command improvements

The validate command MUST support JSON output, auto-detect change names, accept bulk flags like --all and --changes, and provide disambiguation when names collide.
```

Four distinct behaviors jammed into one requirement. Which part failed a test? How do you track partial progress?

**Good:** Split into four focused requirements:

```markdown
### Requirement: JSON Output for Validate
[...scenarios for JSON output...]

### Requirement: Positional Name Argument
[...scenarios for name auto-detection...]

### Requirement: Bulk Validation Flags
[...scenarios for --all, --changes, --specs...]

### Requirement: Type Disambiguation
[...scenarios for --type flag...]
```

Now each requirement can be implemented, tested, and tracked independently.

## Progressive Rigor

litespec's workflow acknowledges that not every change needs the same level of planning upfront. That's why we have patterns:

**Quick Feature**: You know exactly what you need. Small scope. Run through explore, grill briefly, propose, apply, done.

**Exploratory**: You're investigating a problem space. The first few iterations might be vague. Use explore and grill heavily to figure out the shape before proposing.

**Adopt**: You have existing code with no spec. Work backward — reverse-engineer specs from the implementation, then use those as baseline for future changes.

The key is that **rigor scales with uncertainty**. If you're adding a simple flag to an existing command, you don't need a week of grilling. If you're designing a new capability from scratch, you might need multiple explore sessions before you're ready to propose.

## When to Use Litespec (and When Not To)

### Use litespec for:

- **Features and capabilities**: New commands, significant behavior changes, capabilities that will live for a while
- **Projects with multiple contributors**: Specs become shared understanding and a contract that outlives any one person's memory
- **Long-lived code**: If you'll be maintaining this code for months or years, invest in specs now to pay dividends later
- **Teams where context transfer matters**: When someone new joins, specs are the fastest way to understand what the system does

### Don't use litespec for:

- **One-off scripts and throwaway code**: If it's running once and deleted, specs are overhead
- **Trivial refactors**: Renaming a variable, extracting a helper — tests are sufficient
- **Experiments and prototypes**: When you don't know what you're building yet, specs will just slow you down. Prototype first, spec later if it sticks
- **Solo projects with short lifespans**: If you're the only person touching the code and it'll be gone in a week, your brain is the spec

The threshold is: **will anyone else need to understand this code in 6 months?** If yes, write a spec.

## Good Specs vs Bad Specs

### Example 1: Adding a completion command

**Bad:**
```markdown
### Requirement: Shell completions

The tool should provide shell completions for bash, zsh, and fish to make it easier to use.
```

Vague. No SHALL/MUST. "Make it easier to use" is subjective. How do you test this?

**Good:**
```markdown
### Requirement: Shell Completion Generation

The `litespec completion <shell>` command MUST print a valid shell completion script to stdout for the specified shell (bash, zsh, or fish). The script MUST provide completions for all commands and their arguments.

#### Scenario: Bash completion script
- **WHEN** `litespec completion bash` is run
- **THEN** a valid bash completion script is printed to stdout

#### Scenario: Invalid shell
- **WHEN** `litespec completion invalid-shell` is run
- **THEN** an error is printed listing supported shells
```

Observable behavior. Testable. Clear success criteria.

### Example 2: Status command changes

**Bad:**
```markdown
### Requirement: Positional arguments

The status command should accept a positional argument instead of the --change flag to be consistent with other commands.
```

"Consistent with other commands" is design rationale, not behavior. What exactly does the command do?

**Good:**
```markdown
### Requirement: Positional Name for Status

The `litespec status` command MUST accept an optional positional `<name>` argument instead of `--change <name>`. When provided, it shows artifact state for that specific change. When omitted, it shows all changes.

#### Scenario: Status for a named change
- **WHEN** `litespec status my-feature` is run
- **THEN** artifact state for `my-feature` is shown

#### Scenario: Status with no arguments
- **WHEN** `litespec status` is run
- **THEN** all active changes are listed

#### Scenario: Status for nonexistent change
- **WHEN** `litespec status nonexistent` is run
- **THEN** an error is printed to stderr indicating the change was not found with exit code 1
```

Every scenario describes concrete input and output. No ambiguity about what happens in each case.

## The Bottom Line

Specs aren't about ceremony. They're about **communication** — between your present self and your future self, between you and your teammates, between what you want and what you build.

Write them like you're writing a contract. Test them like you're verifying that contract. When they're done, they become the foundation for everything that follows.
