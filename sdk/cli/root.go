package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

// NewRootCommand builds the reqmango command tree.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "reqmango",
		Short:         "reqmango CLI - manage issues, cycles and agents from the terminal",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().String("config", "", "config file (default ~/.reqmango/config.json)")
	root.PersistentFlags().String("api-url", "", "override API base URL")
	root.PersistentFlags().Uint64("workspace", 0, "override workspace ID")
	root.PersistentFlags().Uint64("project", 0, "override project ID")
	root.PersistentFlags().String("output", "table", "output format: table|json")

	root.AddCommand(newAuthCmd())
	root.AddCommand(newWorkspaceCmd())
	root.AddCommand(newProjectCmd())
	root.AddCommand(newIssueCmd())
	root.AddCommand(newCycleCmd())
	root.AddCommand(newMetaCmd())
	root.AddCommand(newAgentCmd())
	root.AddCommand(newAskCmd())
	return root
}

// configPath resolves the --config flag to a path.
func configPath(cmd *cobra.Command) (string, error) {
	p := cmd.Flag("config").Value.String()
	if p != "" {
		return p, nil
	}
	return DefaultConfigPath()
}

// apiURLFor resolves the API base URL:
// --api-url flag > REQMANGO_API_URL env > config file > client.DefaultBaseURL.
func apiURLFor(cmd *cobra.Command, cfg *Config) string {
	if v := cmd.Flag("api-url").Value.String(); v != "" {
		return v
	}
	if v := os.Getenv("REQMANGO_API_URL"); v != "" {
		return v
	}
	if cfg.APIURL != "" {
		return cfg.APIURL
	}
	return client.DefaultBaseURL
}

// patFor resolves the PAT: the config file wins; REQMANGO_PAT is only a
// fallback when the config has no PAT (auth login stays authoritative).
func patFor(cfg *Config) string {
	if cfg.PAT != "" {
		return cfg.PAT
	}
	return os.Getenv("REQMANGO_PAT")
}

// setup resolves the config file and builds the API client.
func setup(cmd *cobra.Command) (*client.Client, *Config, error) {
	path, err := configPath(cmd)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, nil, err
	}
	pat := patFor(cfg)
	if pat == "" {
		return nil, nil, fmt.Errorf("not logged in: run `reqmango auth login` first")
	}
	return client.New(apiURLFor(cmd, cfg), pat), cfg, nil
}

// resolveWorkspace returns --workspace flag > config > error.
func resolveWorkspace(cmd *cobra.Command, cfg *Config) (uint64, error) {
	if v, _ := cmd.Flags().GetUint64("workspace"); v != 0 {
		return v, nil
	}
	if cfg.WorkspaceID != 0 {
		return cfg.WorkspaceID, nil
	}
	return 0, fmt.Errorf("no workspace selected: pass --workspace or run `reqmango workspace switch <id>`")
}

// resolveProject returns --project flag > config > error.
func resolveProject(cmd *cobra.Command, cfg *Config) (uint64, error) {
	if v, _ := cmd.Flags().GetUint64("project"); v != 0 {
		return v, nil
	}
	if cfg.ProjectID != 0 {
		return cfg.ProjectID, nil
	}
	return 0, fmt.Errorf("no project selected: pass --project <id>")
}
