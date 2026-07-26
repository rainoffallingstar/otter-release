package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
	refpkg "github.com/rainoffallingstar/otter/internal/reference"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var referenceCmd = &cobra.Command{
	Use:   "reference",
	Short: "Manage reference registry assets",
}

var referencePromoteCmd = &cobra.Command{
	Use:   "promote <run-yaml-path>",
	Short: "Promote a run's effective reference selection to the project lock",
	Long: `Read a resolved run snapshot and promote the effective reference selection
into the project's references.lock.yaml.

Promotion is atomic: the lock file is written via a temp file + rename.
The run ID is recorded as promoted_from_run_id for audit.`,
	Args: cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		confirm, _ := command.Flags().GetBool("confirm")
		return runReferencePromote(command, args[0], confirm)
	},
}

func init() {
	referencePromoteCmd.Flags().Bool("confirm", false, "Apply the promotion (default is preview-only)")
	rootCmd.AddCommand(referenceCmd)
	referenceCmd.AddCommand(referencePromoteCmd)
}

func runReferencePromote(command *cobra.Command, runYamlPath string, confirm bool) error {
	runSnapshot, err := loadRunSnapshot(runYamlPath)
	if err != nil {
		return err
	}
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(runYamlPath)))
	lockPath := filepath.Join(projectRoot, "references.lock.yaml")
	currentLock, err := configv1.LoadReferencesLock(lockPath)
	if err != nil {
		return fmt.Errorf("load current lock %q: %w", lockPath, err)
	}

	proposedLock := configv1.ReferencesLock{
		SchemaVersion:     currentLock.SchemaVersion,
		References:        make(map[string]configv1.LockedReference),
		PromotedFromRunID: runSnapshot.Run.ID,
	}
	for role, locked := range currentLock.References {
		proposedLock.References[role] = locked
	}
	for _, role := range []string{"primary", "secondary", "graft", "host"} {
		effectiveSel := selectionByRole(runSnapshot.References.EffectiveSelection, role)
		if effectiveSel == "" {
			continue
		}
		id, release, parseErr := configv1.ParseReferenceSelection(effectiveSel)
		if parseErr != nil {
			continue
		}
		resolvedRef := resolvedReferenceForRole(runSnapshot.References.Resolved, configv1.ReferenceRole(role))
		if resolvedRef == nil {
			continue
		}
		proposedLock.References[role] = configv1.LockedReference{
			ID:             id,
			Release:        release,
			ManifestDigest: resolvedRef.ManifestDigest,
		}
	}

	changes := buildLockDiff(&currentLock, runSnapshot)
	if len(changes) == 0 {
		fmt.Fprintln(command.OutOrStdout(), "No reference changes to promote.")
		return nil
	}

	if err := verifyResolvedReferences(runSnapshot); err != nil {
		return fmt.Errorf("reference verification failed before promotion: %w", err)
	}

	if !confirm {
		fmt.Fprintln(command.OutOrStdout(), "Preview: the following lock changes would be applied:")
		for _, change := range changes {
			fmt.Fprintf(command.OutOrStdout(), "  %s\n", change)
		}
		fmt.Fprintf(command.OutOrStdout(), "\nRun ID: %s\n", runSnapshot.Run.ID)
		fmt.Fprintln(command.OutOrStdout(), "\nUse --confirm to apply.")
		return nil
	}

	lockData, err := yaml.Marshal(proposedLock)
	if err != nil {
		return fmt.Errorf("marshal proposed lock: %w", err)
	}
	tmpPath := lockPath + ".tmp"
	if err := os.WriteFile(tmpPath, lockData, 0o644); err != nil {
		return fmt.Errorf("write temporary lock: %w", err)
	}
	if err := os.Rename(tmpPath, lockPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("atomically update lock %q: %w", lockPath, err)
	}
	fmt.Fprintln(command.OutOrStdout(), "Lock updated successfully.")
	fmt.Fprintf(command.OutOrStdout(), "Promoted from run: %s\n", runSnapshot.Run.ID)
	return nil
}

func loadRunSnapshot(path string) (configv1.RunSnapshot, error) {
	data, err := os.Open(path)
	if err != nil {
		return configv1.RunSnapshot{}, fmt.Errorf("open run snapshot %q: %w", path, err)
	}
	defer data.Close()
	var snapshot configv1.RunSnapshot
	decoder := yaml.NewDecoder(data)
	decoder.KnownFields(true)
	if err := decoder.Decode(&snapshot); err != nil {
		return configv1.RunSnapshot{}, fmt.Errorf("parse run snapshot %q: %w", path, err)
	}
	return snapshot, nil
}

func buildLockDiff(currentLock *configv1.ReferencesLock, snapshot configv1.RunSnapshot) []string {
	var changes []string
	for _, role := range []string{"primary", "secondary", "graft", "host"} {
		effectiveSel := string(selectionByRole(snapshot.References.EffectiveSelection, role))
		projectSel := string(selectionByRole(snapshot.References.ProjectSelection, role))
		if effectiveSel == "" {
			continue
		}
		currentLocked, hasLock := currentLock.References[role]
		currentSel := ""
		if hasLock {
			currentSel = currentLocked.ID + "@" + currentLocked.Release
		}
		if projectSel != "" && effectiveSel != projectSel {
			changes = append(changes, fmt.Sprintf("~ %s: %s -> %s", role, projectSel, effectiveSel))
			continue
		}
		if effectiveSel != currentSel {
			resolvedRef := resolvedReferenceForRole(snapshot.References.Resolved, configv1.ReferenceRole(role))
			newLabel := effectiveSel
			if resolvedRef != nil {
				newLabel = fmt.Sprintf("%s (digest: %s...)", effectiveSel, resolvedRef.ManifestDigest[:16])
			}
			changes = append(changes, fmt.Sprintf("~ %s: %s -> %s", role, currentSel, newLabel))
		}
	}
	sort.Strings(changes)
	return changes
}

func selectionByRole(selections configv1.ReferenceSelections, role string) configv1.ReferenceSelection {
	switch role {
	case "primary":
		return selections.Primary
	case "secondary":
		return selections.Secondary
	case "graft":
		return selections.Graft
	case "host":
		return selections.Host
	}
	return ""
}

func resolvedReferenceForRole(references []configv1.ResolvedReference, role configv1.ReferenceRole) *configv1.ResolvedReference {
	for i := range references {
		if references[i].Role == role {
			return &references[i]
		}
	}
	return nil
}

func verifyResolvedReferences(snapshot configv1.RunSnapshot) error {
	for _, ref := range snapshot.References.Resolved {
		report, err := refpkg.VerifyChecksums(ref.RegistryRoot)
		if err != nil {
			return fmt.Errorf("verify %s (%s@%s): %w", ref.Role, ref.ID, ref.Release, err)
		}
		if !report.Passed {
			for _, issue := range report.Issues {
				fmt.Fprintf(os.Stderr, "  %s: %s (expected: %s, got: %s)\n", issue.Asset, issue.Path, issue.Expected, issue.Actual)
			}
			return fmt.Errorf("checksum verification failed for %s (%s@%s)", ref.Role, ref.ID, ref.Release)
		}
		fmt.Fprintf(os.Stderr, "  verified: %s (%s@%s)\n", ref.Role, ref.ID, ref.Release)
	}
	return nil
}
