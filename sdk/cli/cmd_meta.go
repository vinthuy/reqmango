package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"
)

func newMetaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta",
		Short: "Project metadata: states, labels, issue types",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "states",
			Short: "List workflow states of the current project",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, cfg, err := setup(cmd)
				if err != nil {
					return err
				}
				projectID, err := resolveProject(cmd, cfg)
				if err != nil {
					return err
				}
				states, err := cli.ListStates(context.Background(), projectID)
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(states))
				for _, s := range states {
					rows = append(rows, []string{strconv.FormatUint(s.ID, 10), s.Name, s.Group, s.Color})
				}
				return printResult(cmd, []string{"ID", "Name", "Group", "Color"}, rows, states)
			},
		},
		&cobra.Command{
			Use:   "labels",
			Short: "List labels of the current project",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, cfg, err := setup(cmd)
				if err != nil {
					return err
				}
				projectID, err := resolveProject(cmd, cfg)
				if err != nil {
					return err
				}
				labels, err := cli.ListLabels(context.Background(), projectID)
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(labels))
				for _, l := range labels {
					rows = append(rows, []string{strconv.FormatUint(l.ID, 10), l.Name, l.Color})
				}
				return printResult(cmd, []string{"ID", "Name", "Color"}, rows, labels)
			},
		},
		&cobra.Command{
			Use:   "issue-types",
			Short: "List issue types of the current workspace",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, cfg, err := setup(cmd)
				if err != nil {
					return err
				}
				wsID, err := resolveWorkspace(cmd, cfg)
				if err != nil {
					return err
				}
				var projectID uint64
				if projectID, err = resolveProject(cmd, cfg); err != nil {
					projectID = 0
				}
				types, err := cli.ListIssueTypes(context.Background(), wsID, projectID)
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(types))
				for _, t := range types {
					rows = append(rows, []string{strconv.FormatUint(t.ID, 10), t.Name, t.Level})
				}
				return printResult(cmd, []string{"ID", "Name", "Level"}, rows, types)
			},
		},
	)
	return cmd
}
