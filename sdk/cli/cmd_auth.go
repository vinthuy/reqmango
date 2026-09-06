package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/reqmango/tools/client"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with reqmango (login / logout / status / revoke)",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthLogoutCmd(), newAuthStatusCmd(), newAuthRevokeCmd())
	return cmd
}

// promptLine reads one line from stdin after printing prompt to stderr.
func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	var s string
	_, err := fmt.Fscanln(os.Stdin, &s)
	return s, err
}

// readPassword reads a hidden password from the terminal.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func newAuthLoginCmd() *cobra.Command {
	var email, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in with email+password and store a PAT in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}
			apiURL := cmd.Flag("api-url").Value.String()
			if apiURL == "" {
				apiURL = cfg.APIURL
			}
			if apiURL == "" {
				apiURL = client.DefaultBaseURL
			}

			if email == "" {
				email, err = promptLine("Email: ")
				if err != nil {
					return fmt.Errorf("read email: %w", err)
				}
			}
			if password == "" {
				password, err = readPassword("Password: ")
				if err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			ctx := context.Background()
			anon := client.New(apiURL, "")
			tok, err := anon.Login(ctx, email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			// Mint a PAT with the short-lived JWT, then discard the JWT.
			jwtCli := client.New(apiURL, tok.AccessToken)
			host, _ := os.Hostname()
			patResp, err := jwtCli.CreatePAT(ctx, client.CreatePATRequest{Name: "cli-" + host})
			if err != nil {
				return fmt.Errorf("create PAT failed: %w", err)
			}

			cfg.APIURL = apiURL
			cfg.PAT = patResp.Token
			if err := SaveConfig(path, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s. PAT %q saved to %s\n", email, patResp.TokenPrefix, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "email (prompted if omitted)")
	cmd.Flags().StringVar(&password, "password", "", "password (hidden prompt if omitted)")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored PAT from the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}
			cfg.PAT = ""
			if err := SaveConfig(path, cfg); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current config and authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}
			apiURL := cfg.APIURL
			if apiURL == "" {
				apiURL = client.DefaultBaseURL
			}
			masked := "<none>"
			if cfg.PAT != "" {
				masked = cfg.PAT[:min(len(cfg.PAT), len("reqmango_pat_ab3d"))] + "…"
			}
			rows := [][]string{
				{"API URL", apiURL},
				{"PAT", masked},
				{"Workspace ID", strconv.FormatUint(cfg.WorkspaceID, 10)},
				{"Project ID", strconv.FormatUint(cfg.ProjectID, 10)},
				{"Config", path},
			}
			PrintTable(cmd.OutOrStdout(), []string{"Setting", "Value"}, rows)

			if cfg.PAT != "" {
				cli := client.New(apiURL, cfg.PAT)
				me, err := cli.Me(context.Background())
				if err != nil {
					if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 401 {
						fmt.Fprintln(cmd.OutOrStdout(), "\nPAT is invalid or revoked - run `reqmango auth login` to re-authenticate.")
						return nil
					}
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nLogged in as %s (%s)\n", me.DisplayName, me.Email)
			}
			return nil
		},
	}
}

func newAuthRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke <pat-id>",
		Short: "Revoke a PAT by ID (list them with `reqmango auth revoke --list`)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			listOnly, _ := cmd.Flags().GetBool("list")
			if listOnly || len(args) == 0 {
				pats, err := cli.ListPATs(context.Background())
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(pats))
				for _, p := range pats {
					revoked := ""
					if p.RevokedAt != nil {
						revoked = p.RevokedAt.Format("2006-01-02")
					}
					rows = append(rows, []string{strconv.FormatUint(p.ID, 10), p.Name, p.TokenPrefix, revoked})
				}
				PrintTable(cmd.OutOrStdout(), []string{"ID", "Name", "Prefix", "Revoked"}, rows)
				return nil
			}
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid PAT id: %s", args[0])
			}
			if err := cli.RevokePAT(context.Background(), id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Revoked PAT %d.\n", id)
			return nil
		},
	}
	cmd.Flags().Bool("list", false, "list PATs instead of revoking")
	return cmd
}
