package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Aliases: []string{"agents"},
		Short:   "List and trigger AI agents, inspect agent tasks",
	}
	cmd.AddCommand(newAgentListCmd(), newAgentDispatchCmd(), newAgentTaskCmd())
	return cmd
}

func newAgentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List agents in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			agents, err := cli.ListAgents(context.Background(), wsID)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(agents))
			for _, a := range agents {
				rows = append(rows, []string{strconv.FormatUint(a.ID, 10), a.Name, a.AgentType, a.Status})
			}
			return printResult(cmd, []string{"ID", "Name", "Type", "Status"}, rows, agents)
		},
	}
}

func newAgentDispatchCmd() *cobra.Command {
	var issueArg string
	cmd := &cobra.Command{
		Use:   "dispatch <agent-id> <task...>",
		Short: "Trigger an agent to run a task (long-running)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid agent id: %s", args[0])
			}
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			task := ""
			for i, part := range args[1:] {
				if i > 0 {
					task += " "
				}
				task += part
			}
			var issueID *uint64
			if issueArg != "" {
				id, err := resolveIssueID(context.Background(), cli, wsID, issueArg)
				if err != nil {
					return err
				}
				issueID = &id
			}
			act, err := cli.DispatchAgent(context.Background(), wsID, agentID, task, issueID, nil)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), act)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Dispatched to %s (activity id %d)\n", act.AgentName, act.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&issueArg, "issue", "", "related issue id or code")
	return cmd
}

func newAgentTaskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task <task-id>",
		Short: "Show an agent task's status, progress and recent logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %s", args[0])
			}
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			task, err := cli.GetAgentTask(ctx, wsID, taskID)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), task)
			}
			rows := [][]string{
				{"Title", task.Title},
				{"Status", task.Status},
				{"Progress", fmt.Sprintf("%d%%", task.Progress)},
			}
			if task.ErrorInfo != "" {
				rows = append(rows, []string{"Error", task.ErrorInfo})
			}
			PrintTable(cmd.OutOrStdout(), []string{"Field", "Value"}, rows)
			logs, err := cli.GetAgentTaskLogs(ctx, wsID, taskID)
			if err == nil && len(logs) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nRecent logs:\n")
				start := 0
				if len(logs) > 10 {
					start = len(logs) - 10
				}
				for _, l := range logs[start:] {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s %s\n", l.CreatedAt.Format("15:04:05"), l.Level, l.Message)
				}
			}
			return nil
		},
	}
}
