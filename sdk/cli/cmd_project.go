package cli

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "List, show and create projects",
	}
	cmd.AddCommand(newProjectListCmd(), newProjectShowCmd(), newProjectCreateCmd())
	return cmd
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List projects in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			projects, err := cli.ListProjects(context.Background(), wsID)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(projects))
			for _, p := range projects {
				rows = append(rows, []string{strconv.FormatUint(p.ID, 10), p.Identifier, p.Name, strconv.FormatInt(p.TotalIssues, 10)})
			}
			return printResult(cmd, []string{"ID", "Identifier", "Name", "Issues"}, rows, projects)
		},
	}
}

func newProjectShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id|identifier>",
		Short: "Show one project",
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
			var proj *client.Project
			if id, err := strconv.ParseUint(args[0], 10, 64); err == nil {
				proj, err = cli.GetProject(context.Background(), id)
				if err != nil {
					return err
				}
			} else {
				projects, err := cli.ListProjects(context.Background(), wsID)
				if err != nil {
					return err
				}
				for i := range projects {
					if strings.EqualFold(projects[i].Identifier, args[0]) {
						proj = &projects[i]
						break
					}
				}
				if proj == nil {
					return fmt.Errorf("project with identifier %q not found", args[0])
				}
			}
			rows := [][]string{
				{"ID", strconv.FormatUint(proj.ID, 10)},
				{"Identifier", proj.Identifier},
				{"Name", proj.Name},
				{"Issues", strconv.FormatInt(proj.TotalIssues, 10)},
			}
			PrintTable(cmd.OutOrStdout(), []string{"Field", "Value"}, rows)
			return nil
		},
	}
}

func newProjectCreateCmd() *cobra.Command {
	var name, identifier, desc string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			var created client.Project
			body := map[string]any{"name": name, "identifier": identifier}
			if desc != "" {
				body["description"] = desc
			}
			// POST /projects?workspace_id= returns the bare project object.
			q := map[string]string{"workspace_id": strconv.FormatUint(wsID, 10)}
			hdr, err := cli.PostJSON(context.Background(), "/projects", queryFromMap(q), body, &created)
			_ = hdr
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created project %s (id %d)\n", created.Identifier, created.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&identifier, "identifier", "", "short identifier, max 10 chars (required)")
	cmd.MarkFlagRequired("identifier")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	return cmd
}

// queryFromMap converts a map to url.Values (helper for direct endpoints).
func queryFromMap(m map[string]string) url.Values {
	q := url.Values{}
	for k, v := range m {
		q.Set(k, v)
	}
	return q
}
