package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newAskCmd() *cobra.Command {
	var issueArg string
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask the project AI assistant (long-running, stream aggregated)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			question := ""
			for i, part := range args {
				if i > 0 {
					question += " "
				}
				question += part
			}

			projectID, err := resolveProject(cmd, cfg)
			if err != nil {
				return err
			}
			if issueArg != "" {
				wsID, wsErr := resolveWorkspace(cmd, cfg)
				if wsErr != nil {
					return wsErr
				}
				id, err := resolveIssueID(context.Background(), cli, wsID, issueArg)
				if err != nil {
					return err
				}
				iss, err := cli.GetIssue(context.Background(), id)
				if err != nil {
					return err
				}
				projectID = iss.ProjectID
				question = fmt.Sprintf("(About issue %s) %s", issueArg, question)
			}

			reply, err := cli.AIChat(context.Background(), projectID, question, nil)
			if err != nil {
				if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 401 {
					return fmt.Errorf("PAT is invalid or revoked - run `reqmango auth login`")
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), reply.Text)
			if reply.ThreadID != 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n(thread %d)\n", reply.ThreadID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&issueArg, "issue", "", "ask about a specific issue (id or code)")
	return cmd
}
