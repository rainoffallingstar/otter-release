package cmd

import (
	"fmt"

	"github.com/rainoffallingstar/otter/internal/site"
	"github.com/spf13/cobra"
)

var siteCmd = &cobra.Command{
	Use:   "site",
	Short: "Manage site profiles for backend auto-detection",
}

var siteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List discovered site profiles",
	RunE: func(command *cobra.Command, args []string) error {
		locator := site.DefaultLocator()
		profiles, err := locator.List()
		if err != nil {
			return err
		}
		for _, id := range profiles {
			profile, err := locator.Find(id)
			if err != nil {
				fmt.Fprintf(command.OutOrStdout(), "%-24s (error: %v)\n", id, err)
				continue
			}
			backend := profile.Site.Backend
			details := ""
			if profile.Slurm != nil {
				details = fmt.Sprintf("partition=%s account=%s", profile.Slurm.Partition, profile.Slurm.Account)
				if profile.Slurm.QOS != "" {
					details += fmt.Sprintf(" qos=%s", profile.Slurm.QOS)
				}
			}
			fmt.Fprintf(command.OutOrStdout(), "%-24s backend=%-8s %s\n", id, backend, details)
		}
		return nil
	},
}

var siteValidateCmd = &cobra.Command{
	Use:   "validate [site-id]",
	Short: "Validate a site profile against the current environment",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		siteID := "auto"
		if len(args) > 0 {
			siteID = args[0]
		}
		locator := site.DefaultLocator()
		detector := site.NewDetector()
		result, err := detector.Detect(locator, siteID)
		if err != nil {
			return err
		}
		fmt.Fprintf(command.OutOrStdout(), "Backend:  %s\n", result.Backend)
		fmt.Fprintf(command.OutOrStdout(), "Site:    %s\n", result.SiteID)
		fmt.Fprintf(command.OutOrStdout(), "Source:  %s\n", result.Source)
		fmt.Fprintf(command.OutOrStdout(), "Reason:  %s\n", result.Evidence.Reason)
		if result.Evidence.Cluster != "" {
			fmt.Fprintf(command.OutOrStdout(), "Cluster: %s\n", result.Evidence.Cluster)
		}
		if len(result.Evidence.Commands) > 0 {
			fmt.Fprintf(command.OutOrStdout(), "Tools:   %v\n", result.Evidence.Commands)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(siteCmd)
	siteCmd.AddCommand(siteListCmd)
	siteCmd.AddCommand(siteValidateCmd)
}
