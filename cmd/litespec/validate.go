package main

import (
	"fmt"
	"slices"

	"github.com/bermudi/litespec/internal"
)

func cmdValidate(args []string) error {
	fs := newFlagSet("validate", printValidateHelp)
	var flagAll, flagSpecs, flagDecisions, strict, asJSON, asMinimal bool
	var typeFilter string
	fs.BoolVar(&flagAll, "all", false, "validate all specs and decisions")
	fs.BoolVar(&flagSpecs, "specs", false, "validate all specs only")
	fs.BoolVar(&flagDecisions, "decisions", false, "validate all decisions only")
	fs.BoolVar(&strict, "strict", false, "treat warnings as errors")
	fs.StringVar(&typeFilter, "type", "", "disambiguate name: spec|decision")
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	fsArgs := fs.Args()
	var positional string
	if len(fsArgs) > 0 {
		positional = fsArgs[0]
	}

	hasBulk := flagAll || flagSpecs || flagDecisions

	if positional != "" && hasBulk {
		return fmt.Errorf("positional name and bulk flags (--all, --specs, --decisions) are mutually exclusive")
	}

	if typeFilter != "" && positional == "" {
		return fmt.Errorf("--type requires a positional name")
	}

	if typeFilter != "" && hasBulk {
		return fmt.Errorf("--type cannot be used with bulk flags")
	}

	if typeFilter != "" && typeFilter != "spec" && typeFilter != "decision" {
		return fmt.Errorf("--type must be 'spec' or 'decision', got %q", typeFilter)
	}

	root, err := requireProjectRootWithStaleCheck()
	if err != nil {
		return err
	}

	var result *internal.ValidationResult

	if positional != "" {
		specList, _ := internal.ListSpecs(root)
		specNames := make([]string, len(specList))
		for i, s := range specList {
			specNames[i] = s.Name
		}
		isSpec := slices.Contains(specNames, positional)
		isDecision := false
		decisionMatch, _ := internal.FindDecisionBySlug(root, positional)
		if decisionMatch != nil {
			isDecision = true
		}

		if typeFilter == "spec" {
			isDecision = false
		} else if typeFilter == "decision" {
			isSpec = false
		}

		matches := 0
		if isSpec {
			matches++
		}
		if isDecision {
			matches++
		}

		if matches > 1 {
			return fmt.Errorf("%q is ambiguous — matches multiple artifact types. Use --type spec or --type decision", positional)
		}

		if matches == 0 {
			return fmt.Errorf("%q not found as a spec or decision", positional)
		}

		if isSpec {
			result, err = internal.ValidateSpec(root, positional)
		} else {
			result, err = internal.ValidateDecision(root, positional)
		}
	} else {
		if flagDecisions && flagSpecs {
			return fmt.Errorf("--decisions cannot be combined with --specs (use --all to validate everything)")
		}

		if flagDecisions {
			result, err = internal.ValidateDecisions(root)
			if err != nil {
				return err
			}
			if strict && len(result.Warnings) > 0 {
				result.Valid = false
			}
		} else if flagSpecs {
			result, err = internal.ValidateSpecs(root)
		} else {
			result, err = internal.ValidateAll(root, strict)
		}
	}

	if err != nil {
		return err
	}

	out := internal.BuildValidationResultJSON(result)

	// Build minimal JSON representation
	type validateMinimalJSON struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors,omitempty"`
	}
	minJSON := validateMinimalJSON{Valid: out.Valid}
	for _, e := range out.Errors {
		minJSON.Errors = append(minJSON.Errors, e.Message)
	}

	// Build minimal text representation
	var minimalText string
	if !result.Valid {
		minimalText = fmt.Sprintf("invalid\t%d errors\n", len(result.Errors))
		for _, issue := range result.Errors {
			minimalText += fmt.Sprintf("error\t%s\t%s\n", issue.File, issue.Message)
		}
	} else if strict && len(result.Warnings) > 0 {
		minimalText = fmt.Sprintf("invalid\t%d warnings (strict)\n", len(result.Warnings))
	} else {
		minimalText = fmt.Sprintf("ok\t%d %s, %d %s, %d %s\n",
			result.CapabilitiesCount, pluralize("capability", result.CapabilitiesCount),
			result.RequirementsCount, pluralize("requirement", result.RequirementsCount),
			result.ScenariosCount, pluralize("scenario", result.ScenariosCount))
	}

	// Build text representation
	var text string
	for _, issue := range result.Errors {
		text += fmt.Sprintf("ERROR  %s: %s\n", issue.File, issue.Message)
	}
	for _, issue := range result.Warnings {
		text += fmt.Sprintf("WARN   %s: %s\n", issue.File, issue.Message)
	}
	if result.Valid {
		text += fmt.Sprintf("ok: %d %s, %d %s, %d %s\n",
			result.CapabilitiesCount, pluralize("capability", result.CapabilitiesCount),
			result.RequirementsCount, pluralize("requirement", result.RequirementsCount),
			result.ScenariosCount, pluralize("scenario", result.ScenariosCount))
	}

	if err := Render(Response{
		Full:        out,
		Minimal:     minJSON,
		Text:        text,
		MinimalText: minimalText,
	}, asJSON, asMinimal); err != nil {
		return err
	}

	if !result.Valid || (strict && len(result.Warnings) > 0) {
		return fmt.Errorf("validation failed")
	}
	return nil
}
