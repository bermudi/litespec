# Glossary

Project-wide ubiquitous language. Read this before every conversation.

- **Archive**: Promoting a change to implemented — merging its deltas into canon and moving the change directory to `specs/changes/archive/`.
- **Artifact**: One of the four planning documents in a change: proposal.md, specs/, design.md, tasks.md. Created in dependency order during propose.
- **Canon**: The source-of-truth specs in `specs/canon/`. Represents what the system currently IS — accepted capabilities, not proposed changes.
- **Change**: An isolated proposed modification in `specs/changes/<name>/`. Contains planning artifacts (proposal, specs, design, tasks). Tentative until archived.
- **Delta**: A spec describing differences against canon using ADDED/MODIFIED/REMOVED/RENAMED markers. Not a standalone spec — only meaningful relative to a canonical spec.
- **Design**: The architecture artifact of a change — decisions, file changes, and impact analysis. Created after specs so it can reference requirements.
- **Phase**: A group of related tasks in `tasks.md`. One phase = one apply session = one commit. The first phase with unchecked tasks is the current phase.
- **Proposal**: The first artifact of a change — motivation, scope, and non-goals. Sets the contract for everything that follows.
- **Scenario**: A named, concrete example under a requirement using WHEN/THEN format. Every ADDED and MODIFIED requirement must have at least one.
- **Skill**: Generated agent instructions in `.agents/skills/<name>/SKILL.md`. Produced from Go templates via `litespec update`, never written directly.
- **Spec**: A capability document with requirements and scenarios. Exists in two forms: canonical (current truth) and delta (proposed changes).
- **Tasks**: The phased implementation checklist. Organized into phases, applied one phase at a time. Checkbox state drives phase tracking.
