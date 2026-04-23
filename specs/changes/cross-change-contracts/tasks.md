# Cross-Change Contracts — Tasks

## Phase 1: Review Skill

- [ ] Update `.agents/skills/litespec-review/SKILL.md` setup step: after reading status, check `.litespec.yaml` for `dependsOn` and read dependency artifacts. Also read `specs/glossary.md` if it exists for supplementary terminology context.
- [ ] Update `.agents/skills/litespec-review/SKILL.md` all sections: add "Cross-Change Consistency" dimension — when `dependsOn` exists, cross-reference terms and report name drift as WARNING
- [ ] Manually verify the updated review skill by reviewing the cross-change-contracts change itself — confirm cross-change consistency checks appear in the output
