## Motivation

The archive operation is the most destructive action in litespec — it merges deltas into canon and removes the change directory. Currently, a failure between writing canon specs and renaming the change directory leaves the project in an inconsistent state with no recovery path. Additionally, delta file ordering is non-deterministic, which means the same change can produce different merge results on different machines.

## Scope

- Make archive writes atomic: write to temp location, verify, then swap
- Sort delta files deterministically before merging
- Verify canon spec is parseable after archive
- Clean up partial state on failure

## Non-Goals

- Transactional filesystem operations (not portable)
- Git integration (branch per change, commits)
- Changing the delta merge algorithm itself
