package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	"github.com/bermudi/litespec/v2/internal"
)

type validateMinimalJSON struct {
	Valid                   bool     `json:"valid"`
	ValidationScope         string   `json:"validationScope"`
	ImplementationSemantics string   `json:"implementationSemantics"`
	Errors                  []string `json:"errors,omitempty"`
}

func cmdValidate(args []string) error {
	fs := newFlagSet("validate", printValidateHelp)
	var flagAll, flagSpecs, flagDecisions, strict, asJSON, asMinimal bool
	var typeFilter string
	var issueNumber int
	var queuePath string
	fs.BoolVar(&flagAll, "all", false, "validate all specs and decisions")
	fs.BoolVar(&flagSpecs, "specs", false, "validate all specs only")
	fs.BoolVar(&flagDecisions, "decisions", false, "validate all decisions only")
	fs.BoolVar(&strict, "strict", false, "treat warnings as errors")
	fs.StringVar(&typeFilter, "type", "", "disambiguate name: spec|decision")
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")
	fs.IntVar(&issueNumber, "issue", 0, "fetch and validate a single GH issue by number")
	fs.StringVar(&queuePath, "queue", "", "validate a single local queue markdown file")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	issueSet := false
	queueSet := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "issue":
			issueSet = true
		case "queue":
			queueSet = true
		}
	})

	if issueSet && issueNumber < 1 {
		return fmt.Errorf("--issue must be a positive integer, got %d", issueNumber)
	}
	if queueSet && queuePath == "" {
		return fmt.Errorf("--queue requires a non-empty path")
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

	if issueSet && queueSet {
		return fmt.Errorf("--issue and --queue are mutually exclusive")
	}

	if issueSet || queueSet {
		if positional != "" {
			return fmt.Errorf("--issue/--queue cannot be combined with a positional name")
		}
		if hasBulk {
			return fmt.Errorf("--issue/--queue cannot be combined with bulk flags")
		}
		if typeFilter != "" {
			return fmt.Errorf("--issue/--queue cannot be combined with --type")
		}

		root, err := requireProjectRootWithStaleCheck()
		if err != nil {
			return err
		}

		var result *internal.ValidationResult
		if issueNumber != 0 {
			result, err = internal.ValidateGHIssueByNumber(root, issueNumber)
		} else {
			result, err = internal.ValidateQueueFile(queuePath)
		}
		if err != nil {
			return err
		}

		return renderQueueResult(result, asJSON, asMinimal, strict)
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
		if !isSpec {
			if _, statErr := os.Stat(internal.FeatureSpecPath(root, positional)); statErr == nil {
				isSpec = true
			}
		}
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
		} else if flagSpecs {
			result, err = internal.ValidateSpecs(root)
		} else {
			result, err = internal.ValidateAll(root, strict)
		}
	}

	if err != nil {
		return err
	}

	okText := fmt.Sprintf("structure ok; implementation semantics not verified: %d %s, %d %s, %d %s, %d %s\n",
		result.CapabilitiesCount, pluralize("capability", result.CapabilitiesCount),
		result.RequirementsCount, pluralize("requirement", result.RequirementsCount),
		result.ScenariosCount, pluralize("scenario", result.ScenariosCount),
		result.UnitsCount, pluralize("unit", result.UnitsCount))
	okMinimalText := fmt.Sprintf("structure-ok\tsemantics-unverified\t%d %s, %d %s, %d %s, %d %s\n",
		result.CapabilitiesCount, pluralize("capability", result.CapabilitiesCount),
		result.RequirementsCount, pluralize("requirement", result.RequirementsCount),
		result.ScenariosCount, pluralize("scenario", result.ScenariosCount),
		result.UnitsCount, pluralize("unit", result.UnitsCount))

	return renderValidationResult(result, asJSON, asMinimal, strict, okText, okMinimalText, "")
}

func renderQueueResult(result *internal.ValidationResult, asJSON, asMinimal, strict bool) error {
	okText := fmt.Sprintf("structure ok; implementation semantics not verified: %d units\n", result.UnitsCount)
	okMinimalText := fmt.Sprintf("structure-ok\tsemantics-unverified\t%d units\n", result.UnitsCount)
	invalidText := fmt.Sprintf("invalid: %d errors\n", len(result.Errors))

	return renderValidationResult(result, asJSON, asMinimal, strict, okText, okMinimalText, invalidText)
}

func renderValidationResult(result *internal.ValidationResult, asJSON, asMinimal, strict bool, okText, okMinimalText, invalidText string) error {
	out := internal.BuildValidationResultJSON(result)

	minJSON := validateMinimalJSON{
		Valid:                   out.Valid,
		ValidationScope:         out.ValidationScope,
		ImplementationSemantics: out.ImplementationSemantics,
	}
	for _, e := range out.Errors {
		minJSON.Errors = append(minJSON.Errors, e.Message)
	}

	var minimalText string
	if !result.Valid {
		minimalText = fmt.Sprintf("invalid\t%d errors\n", len(result.Errors))
		for _, issue := range result.Errors {
			minimalText += fmt.Sprintf("error\t%s\t%s\n", issue.File, issue.Message)
		}
	} else if strict && hasNonExemptWarnings(result.Warnings) {
		minimalText = fmt.Sprintf("invalid\t%d warnings (strict)\n", len(result.Warnings))
	} else {
		minimalText = okMinimalText
	}

	var text string
	if result.Valid {
		text += okText
	}
	for _, issue := range result.Errors {
		text += fmt.Sprintf("ERROR  %s: %s\n", issue.File, issue.Message)
	}
	for _, issue := range result.Warnings {
		text += fmt.Sprintf("WARN   %s: %s\n", issue.File, issue.Message)
	}
	if !result.Valid && invalidText != "" {
		text += invalidText
	}

	if err := Render(Response{
		Full:        out,
		Minimal:     minJSON,
		Text:        text,
		MinimalText: minimalText,
	}, asJSON, asMinimal); err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}
	if strict {
		for _, w := range result.Warnings {
			if !w.StrictExempt {
				return fmt.Errorf("validation failed")
			}
		}
	}
	return nil
}
