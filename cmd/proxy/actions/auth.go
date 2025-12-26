package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gomods/athens/pkg/config"
	"github.com/mitchellh/go-homedir"
)

// initializeAuthFile checks if provided auth file is at a pre-configured path
// and moves to home directory -- note that this will override whatever
// .netrc/.hgrc file you have in your home directory.
func initializeAuthFile(path string) error {
	if path == "" {
		return nil
	}

	fileBts, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	hdir, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}

	fileName := transformAuthFileName(filepath.Base(path))
	rcp := filepath.Join(hdir, fileName)
	if err := os.WriteFile(rcp, fileBts, 0o600); err != nil {
		return fmt.Errorf("writing to auth file: %w", err)
	}

	return nil
}

// netrcFromToken takes a github token and creates a .netrc
// file for you, overriding whatever might be already there.
func netrcFromToken(tok string) error {
	fileContent := fmt.Sprintf("machine github.com login %s\n", tok)
	hdir, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("getting homedir: %w", err)
	}
	rcp := filepath.Join(hdir, getNETRCFilename())
	if err := os.WriteFile(rcp, []byte(fileContent), 0o600); err != nil {
		return fmt.Errorf("writing to netrc file: %w", err)
	}
	return nil
}

func transformAuthFileName(authFileName string) string {
	if root := strings.TrimLeft(authFileName, "._"); root == "netrc" {
		return getNETRCFilename()
	}
	return authFileName
}

func getNETRCFilename() string {
	if runtime.GOOS == "windows" {
		return "_netrc"
	}
	return ".netrc"
}

// setupGitCredentialHelper configures git to use token-embedded URLs
// to support multi-org and multi-provider PAT access to private repos
func setupGitCredentialHelper(conf *config.Config) error {
	// Only setup if we have any git auth configured
	if len(conf.GitAuths) == 0 && conf.GithubToken == "" {
		return nil
	}

	// Get home directory for the athens user
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/var/lib/athens"
	}

	// Create .gitconfig file for Athens
	gitConfigPath := filepath.Join(homeDir, ".gitconfig")

	// Build git config with URL rewrites for each org/provider
	var gitConfig strings.Builder

	// Add URL rewrites for git providers
	for _, auth := range conf.GitAuths {
		username := auth.User
		if username == "" {
			// Default usernames for common providers
			switch auth.Host {
			case "github.com":
				username = "x-access-token"
			case "gitlab.com", "gitlab.example.com":
				username = "oauth2"
			case "bitbucket.org":
				username = "x-token-auth"
			default:
				username = "token"
			}
		}

		// Rewrite https://host/org/ to https://user:token@host/org/
		gitConfig.WriteString(fmt.Sprintf(`[url "https://%s:%s@%s/%s/"]
	insteadOf = https://%s/%s/
`, username, auth.Token, auth.Host, auth.Org, auth.Host, auth.Org))
	}

	// Fallback: if there's a default GitHub token, use it for any github.com URL
	if conf.GithubToken != "" && len(conf.GitAuths) == 0 {
		gitConfig.WriteString(fmt.Sprintf(`[url "https://x-access-token:%s@github.com/"]
	insteadOf = https://github.com/
`, conf.GithubToken))
	}

	// Write the git config file
	if err := os.WriteFile(gitConfigPath, []byte(gitConfig.String()), 0600); err != nil {
		return fmt.Errorf("writing gitconfig: %w", err)
	}

	// Set GIT_CONFIG_GLOBAL so the Go command uses our gitconfig
	os.Setenv("GIT_CONFIG_GLOBAL", gitConfigPath)

	return nil
}
