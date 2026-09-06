package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newCycleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cycle",
		Aliases: []string{"cycles"},
		Short:   "List cycles and view progress / burndown",
	}
	cmd.AddCommand(newCycleListCmd(), newCycleProgressCmd(), newCycleBurndownCmd())
	return cmd
}

func newCycleListCmd() *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cycles of the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			projectID, err := resolveProject(cmd, cfg)
			if err != nil {
				return err
			}
			res, err := cli.ListCycles(context.Background(), projectID, status, 50, 0)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(res.Items))
			for _, c := range res.Items {
				rows = append(rows, []string{
					strconv.FormatUint(c.ID, 10), c.Name, c.Status,
					fmt.Sprintf("%.0f%%", c.Progress), c.StartDate,
				})
			}
			return printResult(cmd, []string{"ID", "Name", "Status", "Progress", "Start"}, rows, res)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter: upcoming|active|completed|cancelled")
	return cmd
}

func newCycleProgressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "progress <cycle-id>",
		Short: "Show a cycle's progress breakdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid cycle id: %s", args[0])
			}
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			p, err := cli.GetCycleProgress(context.Background(), id)
			if err != nil {
				return err
			}
			rows := [][]string{
				{"Cycle", fmt.Sprintf("%s (%d)", p.CycleName, p.CycleID)},
				{"Progress", fmt.Sprintf("%.0f%%", p.Progress)},
				{"Issues", fmt.Sprintf("%d/%d completed", p.CompletedIssues, p.TotalIssues)},
			}
			for _, sb := range p.StateBreakdown {
				rows = append(rows, []string{"State: " + sb.State, strconv.FormatInt(sb.Count, 10)})
			}
			PrintTable(cmd.OutOrStdout(), []string{"Metric", "Value"}, rows)
			return nil
		},
	}
}

func newCycleBurndownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "burndown <cycle-id>",
		Short: "Show a cycle's burndown data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid cycle id: %s", args[0])
			}
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			b, err := cli.GetCycleBurndown(context.Background(), id)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), b)
			}
			onTrack := "yes"
			if !b.IsOnTrack {
				onTrack = "no"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Cycle: %s\nIssues: %d total, %d completed\nRemaining: %.0f (ideal %.0f)\nOn track: %s\n",
				b.CycleName, b.TotalIssues, b.ActualCompleted, b.ActualRemaining, b.IdealRemaining, onTrack)
			return nil
		},
	}
}
