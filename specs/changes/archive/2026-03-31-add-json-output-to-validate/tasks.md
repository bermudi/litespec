## Phase 1: Types
- [ ] Add ValidationResultJSON type to internal/json.go
- [ ] Add BuildValidationResultJSON builder function

## Phase 2: CLI Wiring
- [ ] Add --json flag parsing to cmdValidate
- [ ] Output JSON when flag is present

## Phase 3: Verification
- [x] Build and vet
- [x] Smoke test validate --json with a change
