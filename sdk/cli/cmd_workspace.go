package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "List, create and switch workspaces",
	}
	cmd.AddCommand(
		newWorkspaceListCmd(),
		newWorkspaceCreateCmd(),
		&cobra.Command{
			Use:   "switch <id>",
			Short: "Remember the current workspace in the config",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid workspace id: %s", args[0])
				}
				path, err := configPath(cmd)
				if err != nil {
					return err
				}
				cfg, err := LoadConfig(path)
				if err != nil {
					return err
				}
				apiURL := apiURLFor(cmd, cfg)
				name := ""
				if patFor(cfg) != "" {
					cli := clientFor(cfg, apiURL)
					ws, err := cli.ListWorkspaces(context.Background())
					if err == nil {
						for _, w := range ws {
							if w.ID == id {
								name = w.Name
								break
							}
						}
					}
				}
				cfg.WorkspaceID = id
				if err := SaveConfig(path, cfg); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Switched to workspace %d %s\n", id, name)
				return nil
			},
		},
	)
	return cmd
}

func newWorkspaceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accessible workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			ws, err := cli.ListWorkspaces(context.Background())
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(ws))
			for _, w := range ws {
				rows = append(rows, []string{strconv.FormatUint(w.ID, 10), w.Name, w.Slug})
			}
			return printResult(cmd, []string{"ID", "Name", "Slug"}, rows, ws)
		},
	}
}

func newWorkspaceCreateCmd() *cobra.Command {
	var name, slug string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			ws, err := cli.CreateWorkspace(context.Background(), client.WorkspaceCreateRequest{
				Name: name,
				Slug: slug,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created workspace %s (id %d)\n", ws.Name, ws.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "workspace name (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&slug, "slug", "", "URL-friendly slug, max 50 chars (required)")
	cmd.MarkFlagRequired("slug")
	return cmd
}

// clientFor builds a client from config (helper for commands that only
// need config access). The PAT falls back to REQMANGO_PAT.
func clientFor(cfg *Config, apiURL string) *client.Client {
	if apiURL == "" {
		apiURL = client.DefaultBaseURL
	}
	return client.New(apiURL, patFor(cfg))
}
