package site

import (
	"fmt"
	"os/exec"
	"strings"

	configv1 "github.com/rainoffallingstar/otter/internal/config/v1"
)

type toolChecker func(tool string) bool

type clusterNameGetter func() (string, bool)

type slurmValidator func(*SiteProfile) error

type Detector struct {
	checkTool        toolChecker
	getClusterName   clusterNameGetter
	validateProfile  slurmValidator
}

type DetectorOption func(*Detector)

func WithToolChecker(checker toolChecker) DetectorOption {
	return func(detector *Detector) { detector.checkTool = checker }
}

func WithClusterNameGetter(getter clusterNameGetter) DetectorOption {
	return func(detector *Detector) { detector.getClusterName = getter }
}

func WithSlurmValidator(validator slurmValidator) DetectorOption {
	return func(detector *Detector) { detector.validateProfile = validator }
}

func NewDetector(options ...DetectorOption) *Detector {
	detector := &Detector{
		checkTool:        defaultCheckTool,
		getClusterName:   defaultGetClusterName,
		validateProfile:  validateSlurmProfile,
	}
	for _, option := range options {
		option(detector)
	}
	return detector
}

func defaultCheckTool(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func defaultGetClusterName() (string, bool) {
	output, err := exec.Command("scontrol", "show", "config").Output()
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(output), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ClusterName") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				cleaned := strings.TrimSpace(parts[1])
				if cleaned != "" {
					return cleaned, true
				}
			}
		}
	}
	return "", false
}

const slurmToolchainToolCount = 4

var slurmToolchain = []string{"sbatch", "squeue", "sacct", "scancel"}

func (detector *Detector) Detect(locator Locator, siteOverride string) (DetectionResult, error) {
	available := detector.availableSlurmTools()
	commandNames := slurmToolchain

	if len(available) == slurmToolchainToolCount {
		clusterName, hasCluster := detector.getClusterName()
		evidence := configv1.BackendEvidence{Commands: commandNames}
		if hasCluster {
			evidence.Cluster = clusterName
		}

		profile, err := detector.resolveSlurmProfile(locator, siteOverride)
		if err != nil {
			return DetectionResult{}, fmt.Errorf("SLURM detected (cluster: %q) but no matching site profile: %w", clusterName, err)
		}
		if err := detector.validateProfile(profile); err != nil {
			return DetectionResult{}, fmt.Errorf("site profile %q validation failed on cluster %q: %w", profile.Site.ID, clusterName, err)
		}
		evidence.Reason = fmt.Sprintf("full SLURM toolchain present on cluster %q; site profile %q", clusterName, profile.Site.ID)
		siteResources := siteResourcesFromProfile(profile)
		return DetectionResult{
			Backend:       configv1.BackendSlurm,
			SiteID:        profile.Site.ID,
			Evidence:      evidence,
			Source:        configv1.SourceDetection,
			SitePaths:     profile.Paths,
			SiteResources: siteResources,
		}, nil
	}

	if len(available) > 0 && len(available) < slurmToolchainToolCount {
		missing := missingTools(commandNames, available)
		return DetectionResult{}, fmt.Errorf(
			"partial SLURM toolchain detected (found: %s, missing: %s); repair your SLURM installation or use --backend local",
			strings.Join(available, ", "),
			strings.Join(missing, ", "),
		)
	}

	return DetectionResult{
		Backend: configv1.BackendLocal,
		SiteID:  defaultLocalSiteID,
		Evidence: configv1.BackendEvidence{
			Reason: "no SLURM toolchain detected; using local execution",
		},
		Source: configv1.SourceDetection,
	}, nil
}

func (detector *Detector) Validate(requested configv1.Backend, locator Locator, siteOverride string) (DetectionResult, error) {
	if requested != configv1.BackendLocal && requested != configv1.BackendSlurm {
		return DetectionResult{}, fmt.Errorf("cannot validate unsupported backend %q", requested)
	}
	detected, err := detector.Detect(locator, siteOverride)
	if err != nil {
		return DetectionResult{}, err
	}
	if detected.Backend != requested {
		return DetectionResult{}, fmt.Errorf(
			"requested backend %q but detection found %q (site %q)",
			requested, detected.Backend, detected.SiteID,
		)
	}
	return detected, nil
}

func (detector *Detector) resolveSlurmProfile(locator Locator, siteOverride string) (*SiteProfile, error) {
	if siteOverride != "" && siteOverride != "auto" {
		profile, err := locator.Find(siteOverride)
		if err != nil {
			return nil, fmt.Errorf("explicit site %q not found: %w", siteOverride, err)
		}
		return profile, nil
	}
	profile, err := locator.FindByBackend("slurm")
	if err != nil {
		return nil, fmt.Errorf("auto-detection could not find a slurm site profile: %w", err)
	}
	return profile, nil
}

func (detector *Detector) availableSlurmTools() []string {
	var available []string
	for _, tool := range slurmToolchain {
		if detector.checkTool(tool) {
			available = append(available, tool)
		}
	}
	return available
}

func missingTools(all []string, present []string) []string {
	presentSet := make(map[string]bool, len(present))
	for _, tool := range present {
		presentSet[tool] = true
	}
	var missing []string
	for _, tool := range all {
		if !presentSet[tool] {
			missing = append(missing, tool)
		}
	}
	return missing
}

func siteResourcesFromProfile(profile *SiteProfile) configv1.ProjectResources {
	if profile == nil || profile.Slurm == nil {
		return configv1.ProjectResources{}
	}
	defaults := configv1.ResourceSpec{}
	if profile.Slurm.DefaultTime != "" {
		defaults.Time = profile.Slurm.DefaultTime
	}
	if profile.Slurm.Partition != "" {
		defaults.Partition = profile.Slurm.Partition
	}
	return configv1.ProjectResources{Defaults: defaults}
}

func validateSlurmProfile(profile *SiteProfile) error {
	if profile.Slurm == nil {
		return fmt.Errorf("slurm backend requires slurm configuration")
	}
	if strings.TrimSpace(profile.Slurm.Partition) == "" {
		return fmt.Errorf("slurm partition is required")
	}
	if strings.TrimSpace(profile.Slurm.Account) == "" {
		return fmt.Errorf("slurm account is required")
	}
	if err := ValidateSlurmPartition(profile.Slurm.Partition); err != nil {
		return fmt.Errorf("partition %q: %w", profile.Slurm.Partition, err)
	}
	if err := ValidateSlurmAccount(profile.Slurm.Account); err != nil {
		return fmt.Errorf("account %q: %w", profile.Slurm.Account, err)
	}
	if err := ValidateSlurmAccountPartition(profile.Slurm.Account, profile.Slurm.Partition); err != nil {
		return fmt.Errorf("account-partition: %w", err)
	}
	if profile.Slurm.QOS != "" {
		if err := ValidateSlurmQOS(profile.Slurm.QOS); err != nil {
			return fmt.Errorf("qos %q: %w", profile.Slurm.QOS, err)
		}
	}
	if err := ValidateSlurmTimeFormat(profile.Slurm.DefaultTime); err != nil {
		return fmt.Errorf("default_time %q: %w", profile.Slurm.DefaultTime, err)
	}
	if err := ValidateSlurmMaxJobs(profile.Slurm.MaxJobs); err != nil {
		return err
	}
	if err := CheckComputeNodePath(profile.Paths.ReferenceRoot); err != nil {
		return fmt.Errorf("reference_root: %w", err)
	}
	if profile.Paths.ScratchRoot != "" {
		if err := CheckComputeNodePath(profile.Paths.ScratchRoot); err != nil {
			return fmt.Errorf("scratch_root: %w", err)
		}
	}
	return nil
}
