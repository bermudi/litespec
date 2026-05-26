Create the proposal artifact at the specified output path.

This is the first artifact — it has no dependencies. Build it from the conversation context and codebase exploration.

Structure:

## Motivation
Why this change is needed. What problem does it solve or what opportunity does it address?

## Scope
What is included in this change. Be specific about:
- Which capabilities are affected
- What behavior changes
- What new functionality is introduced

## Non-Goals
What is explicitly NOT included. This prevents scope creep and sets clear boundaries.

Rules:
- Be concrete, not aspirational — describe a specific change, not a wishlist
- Keep it focused — if the scope feels too large, suggest splitting into multiple changes
- The proposal informs everything downstream: specs describe the scope in detail, design explains how to implement it, tasks break it into phases