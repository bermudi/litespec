# Review Instructions Are Trusted Bootstrap

## Status

accepted

## Context

The review safety policy is delivered through local instruction files. A coding-agent harness normally reads system instructions, `AGENTS.md`, and the selected review `SKILL.md` before the skill can apply any path-screening rule. Litespec CLI code also runs after that bootstrap, so moving screening into the CLI cannot protect instruction files that the harness already loaded.

## Decision

Harness/system instructions and every repository instruction file auto-loaded by the harness to activate `litespec-review` SHALL form a trusted bootstrap boundary. This includes applicable `AGENTS.md` files and the selected review `SKILL.md`.

Litespec's local-path screening guarantee SHALL begin only after the review skill is active. From that point, the remote GH issue body is the only additional content that may be read before screening; every other local path MUST pass the review safety process first.

Litespec SHALL NOT claim to screen or secure bootstrap inputs. Users who cannot trust auto-loaded repository instructions MUST use a harness-level sandbox or pre-load policy before invoking litespec review.

## Consequences

The safety contract is honest about what the skill can enforce. Auto-loaded instructions remain a trusted computing base and can still expose or contain unsafe content before review begins. Litespec cannot remove that risk portably across agent harnesses; stronger isolation belongs in the harness.
