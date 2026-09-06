package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issue",
		Aliases: []string{"issues"},
		Short:   "List, show, create, update and search issues",
	}
	cmd.AddCommand(newIssueListCmd(), newIssueShowCmd(), newIssueCreateCmd(), newIssueUpdateCmd(), newIssueSearchCmd())
	return cmd
}

// resolveAssignee maps "" → 0, "me" → current user ID, else parses uint64.
func resolveAssignee(ctx context.Context, cli *client.Client, s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	if s == "me" {
		me, err := cli.Me(ctx)
		if err != nil {
			return 0, err
		}
		return me.ID, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func issueRow(i client.Issue) []string {
	return []string{
		strconv.FormatUint(i.ID, 10),
		strconv.Itoa(i.SequenceID),
		i.Name,
		i.Priority,
		i.StateName,
	}
}

var issueTableHeaders = []string{"ID", "Seq", "Name", "Priority", "State"}

func newIssueListCmd() *cobra.Command {
	var rql, priority, assignee string
	var stateID, cycleID, limit int64
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues (optionally filtered by --rql)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()
			opts := client.IssueListOptions{RQL: rql, Priority: priority, StateID: uint64(stateID), CycleID: uint64(cycleID), Limit: int(limit)}
			if projectID, err := resolveProject(cmd, cfg); err == nil {
				opts.ProjectID = projectID
			}
			if wsID, err := resolveWorkspace(cmd, cfg); err == nil {
				opts.WorkspaceID = wsID
			}
			if assignee != "" {
				id, err := resolveAssignee(ctx, cli, assignee)
				if err != nil {
					return err
				}
				opts.AssigneeID = id
			}
			res, err := cli.ListIssues(ctx, opts)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(res.Items))
			for _, i := range res.Items {
				rows = append(rows, issueRow(i))
			}
			if err := printResult(cmd, issueTableHeaders, rows, res); err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() != "json" {
				fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d\n", res.Total)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&rql, "rql", "", "RQL query, e.g. `priority = \"high\"`")
	cmd.Flags().StringVar(&priority, "priority", "", "filter by priority")
	cmd.Flags().Int64Var(&stateID, "state", 0, "filter by state ID")
	cmd.Flags().StringVar(&assignee, "assignee", "", "filter by assignee user ID or `me`")
	cmd.Flags().Int64Var(&cycleID, "cycle", 0, "filter by cycle ID")
	cmd.Flags().Int64Var(&limit, "limit", 0, "page size")
	return cmd
}

func newIssueShowCmd() *cobra.Command {
	var withComments bool
	cmd := &cobra.Command{
		Use:   "show <id|code>",
		Short: "Show one issue by numeric ID or code like DEMO-42",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()
			wsID, _ := resolveWorkspace(cmd, cfg)
			id, err := cli.ResolveIssueCode(ctx, wsID, args[0])
			if err != nil && !isNumeric(args[0]) {
				return err
			}
			if isNumeric(args[0]) {
				id, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid issue id: %s", args[0])
				}
			}
			iss, err := cli.GetIssue(ctx, id)
			if err != nil {
				return err
			}
			rows := [][]string{
				{"ID", strconv.FormatUint(iss.ID, 10)},
				{"Code", args[0]},
				{"Name", iss.Name},
				{"Priority", iss.Priority},
				{"State", iss.StateName},
			}
			PrintTable(cmd.OutOrStdout(), []string{"Field", "Value"}, rows)
			if withComments {
				comments, _, err := cli.ListComments(ctx, id, 1, 100)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nComments (%d):\n", len(comments))
				for _, c := range comments {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", c.CreatedAt.Format("2006-01-02 15:04"), c.Body)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&withComments, "comments", false, "include comments")
	return cmd
}

func isNumeric(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

func newIssueCreateCmd() *cobra.Command {
	var title, desc, priority, assignee, labels string
	var stateID, typeID, parentID, cycleID int64
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			projectID, err := resolveProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			proj, err := cli.GetProject(ctx, projectID)
			if err != nil {
				return err
			}
			body := &client.CreateIssueRequest{Name: title, DescriptionHTML: desc, Priority: priority}
			if stateID != 0 {
				body.StateID = &[]uint64{uint64(stateID)}[0]
			}
			if typeID != 0 {
				body.TypeID = &[]uint64{uint64(typeID)}[0]
			}
			if parentID != 0 {
				body.ParentID = &[]uint64{uint64(parentID)}[0]
			}
			if cycleID != 0 {
				body.CycleID = &[]uint64{uint64(cycleID)}[0]
			}
			if assignee != "" {
				id, err := resolveAssignee(ctx, cli, assignee)
				if err != nil {
					return err
				}
				body.AssigneeIDs = []uint64{id}
			}
			for _, s := range strings.Split(labels, ",") {
				if n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil {
					body.LabelIDs = append(body.LabelIDs, n)
				}
			}
			iss, err := cli.CreateIssue(ctx, projectID, proj.WorkspaceID, body)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), iss)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created issue %s-%d (id %d)\n", proj.Identifier, iss.SequenceID, iss.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "issue title (required)")
	cmd.MarkFlagRequired("title")
	cmd.Flags().StringVar(&desc, "desc", "", "description (HTML)")
	cmd.Flags().StringVar(&priority, "priority", "", "priority: none|low|medium|high|urgent")
	cmd.Flags().Int64Var(&stateID, "state", 0, "state ID")
	cmd.Flags().Int64Var(&typeID, "type", 0, "issue type ID")
	cmd.Flags().StringVar(&assignee, "assignee", "", "assignee user ID or `me`")
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated label IDs")
	cmd.Flags().Int64Var(&parentID, "parent", 0, "parent issue ID")
	cmd.Flags().Int64Var(&cycleID, "cycle", 0, "cycle ID")
	return cmd
}

func newIssueUpdateCmd() *cobra.Command {
	var title, desc, priority, assignee, labels string
	var stateID, cycleID int64
	cmd := &cobra.Command{
		Use:   "update <id|code>",
		Short: "Update an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()
			wsID, _ := resolveWorkspace(cmd, cfg)
			id, err := resolveIssueID(ctx, cli, wsID, args[0])
			if err != nil {
				return err
			}
			body := &client.UpdateIssueRequest{}
			if title != "" {
				body.Name = &title
			}
			if desc != "" {
				body.DescriptionHTML = &desc
			}
			if priority != "" {
				body.Priority = &priority
			}
			if stateID != 0 {
				body.StateID = &[]uint64{uint64(stateID)}[0]
			}
			if cycleID != 0 {
				body.CycleID = &[]uint64{uint64(cycleID)}[0]
			}
			if assignee != "" {
				aid, err := resolveAssignee(ctx, cli, assignee)
				if err != nil {
					return err
				}
				body.AssigneeIDs = []uint64{aid}
			}
			if labels != "" {
				for _, s := range strings.Split(labels, ",") {
					if n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil {
						body.LabelIDs = append(body.LabelIDs, n)
					}
				}
			}
			iss, err := cli.UpdateIssue(ctx, id, body)
			if err != nil {
				if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 409 {
					fmt.Fprintf(cmd.OutOrStdout(), "Approval required: %v\n", apiErr.Body)
					return nil
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated issue %d\n", iss.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&desc, "desc", "", "new description (HTML)")
	cmd.Flags().StringVar(&priority, "priority", "", "new priority")
	cmd.Flags().Int64Var(&stateID, "state", 0, "new state ID")
	cmd.Flags().StringVar(&assignee, "assignee", "", "new assignee user ID or `me` (replaces)")
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated label IDs (replaces)")
	cmd.Flags().Int64Var(&cycleID, "cycle", 0, "move to cycle ID")
	return cmd
}

// resolveIssueID accepts a numeric ID or a code like DEMO-42.
func resolveIssueID(ctx context.Context, cli *client.Client, wsID uint64, arg string) (uint64, error) {
	if isNumeric(arg) {
		return strconv.ParseUint(arg, 10, 64)
	}
	if wsID == 0 {
		return 0, fmt.Errorf("issue code %q needs --workspace to resolve the project identifier", arg)
	}
	return cli.ResolveIssueCode(ctx, wsID, arg)
}

func newIssueSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search across workspace issues",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			results, err := cli.SearchIssues(context.Background(), wsID, args[0], nil, 20)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(results))
			for _, r := range results {
				rows = append(rows, []string{r.ProjectIdentifier + "-" + strconv.Itoa(r.SequenceID), r.Name, strconv.FormatUint(r.ID, 10)})
			}
			return printResult(cmd, []string{"Code", "Name", "ID"}, rows, results)
		},
	}
}
