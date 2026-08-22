You are rebuilding a unit that review reopened — a CRITICAL or WARNING finding showed the unit's `Done means:` or `Verify:` was not satisfied. The previous Verify passed but didn't prove the outcome. Your job is to cure the disease, not patch the symptom.

**Scope expands, does not narrow:**
- Identify the **abstract pattern** behind the finding. Do not fix just the reported `file:line`.
- Search the affected module for the same pattern. Fix all instances, not just the cited one.
- After fixing, re-read the affected module end-to-end. Ask: "Did my changes introduce new surface area? What invariants might now be broken?"
- Run the full test suite, not just tests related to your fix.

**Per-finding loop:**

1. Read the finding and the relevant source. If it references a spec requirement, read that spec section first.
2. Search the module for the same pattern — fix all instances.
3. Make the minimal change that addresses the pattern.
4. Run `go build` and relevant tests. If both pass, move on. If either fails, fix and retry.
5. If the same verification fails twice on the same finding, stop. Re-read the finding and code before retrying.
6. State what was fixed and where.

**Final verification:**
1. `go build`
2. `go test ./...`
3. `go vet ./...`
4. Establish a clean pre commit where the unit's exact `Verify:` fails because the fix is absent. If the old verifier still passes, strengthen only the verifier in one verifier-only commit and use that as pre.
5. Commit the pattern-wide fix as the single implementation commit, then run the same exact Verify and require exit status 0.
6. Record a fresh pre/post receipt. Never amend either evidence commit.
7. `litespec validate` — confirm no structural regressions.

**Escalation:**
If a finding cannot be resolved, state it explicitly: "Finding [X] in `file:line` could not be resolved because [reason]." Never silently skip a finding. Suggest next steps (decision needed, re-plan the unit).

**Guardrails:**
- Do not fix only the cited `file:line` while ignoring structurally identical code nearby.
- Do not declare done after tests pass without re-reading the changed module.
- SUGGESTIONs remain optional. If evidence proves one is a unit-contract violation, it must be reclassified as CRITICAL or WARNING before it can expand rebuild scope.
- Do not modify specs, the GH issue, or decisions — fix implementation code only. If the spec itself is wrong, pause and ask.
- Stay within the unit. Drive-bys outside the unit's scope get noted, not fixed.
