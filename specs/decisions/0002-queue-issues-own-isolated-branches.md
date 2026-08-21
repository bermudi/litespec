# Queue Issues Own Isolated Branches

## Status

accepted

## Context

Review needs an exact boundary around one queue issue. A starting commit alone is not that boundary: unrelated commits and dirty work made after the start become part of the diff and can block the issue. An inferred commit set is also fragile when commits overlap or later commits modify the same lines.

## Decision

Each queue issue SHALL own a dedicated `litespec/<change-name>` branch created from a clean working tree by `litespec-plan` in clear mode. The issue body and local queue mirror SHALL record both `Base: <sha>` and `Branch: <branch>`.

All commits and working-tree changes on that branch SHALL belong to the issue. Unrelated work MUST use another branch or worktree. Build and review SHALL stop when the current branch does not match the recorded branch.

Review scope SHALL contain the tracked diff from `Base:` to the current working tree plus every untracked path reported by NUL-delimited Git status. Before reading contents, review screens every path and file type without following symlinks; secret-like paths, symlinks, and non-regular files stop review without a verdict.

## Consequences

The review boundary is exact without maintaining a separate commit manifest. Queue issues can no longer share a branch, and planning cannot start with dirty work. Concurrent changes require separate branches or worktrees. Review must inspect untracked files explicitly because `git diff` omits them.
