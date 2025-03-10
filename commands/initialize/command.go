package initialize

import (
	"bufio"
	"fmt"
	"ide-config-sync/config"
	"ide-config-sync/fsutil"
	"ide-config-sync/persistance"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func readUseLocalOnly() (bool, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Do you want to use local only mode? [y/N]: ")
	for {
		scanned := scanner.Scan()
		if !scanned {
			return false, fmt.Errorf("failed to read input")
		}

		text := scanner.Text()
		if text == "" {
			return false, nil
		}

		switch strings.ToLower(text) {
		case "y", "yes", "true", "1":
			return true, nil

		case "n", "no", "false", "0":
			return false, nil
		default:
			color.Red("Invalid option '%s'", text)
		}
	}
}

func readDatabasePath() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Where should the configurations be stored? [%s]: ", config.DefaultDatabaseRepoPath)
	scanned := scanner.Scan()
	if !scanned {
		color.Red("failed to read input")
		return ""
	}
	text := scanner.Text()
	if text == "" {
		return config.DefaultDatabaseRepoPath
	}

	return text
}

func readDatabaseRepositoryURL(c *config.Config) string {
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	for {
		if c.DatabaseRepoURL != "" {
			fmt.Printf("Enter the database repository URL [%s]: ", c.DatabaseRepoURL)
		} else {
			fmt.Print("Enter the database repository URL: ")
		}
		scanned := scanner.Scan()
		if !scanned {
			color.Red("failed to read input")
			return ""
		}
		text = scanner.Text()
		_, err := url.Parse(text)
		if err != nil {
			color.Red("invalid URL: %s", err)
			continue
		}
		break
	}
	return text
}

var Command = &cobra.Command{
	Use:   "init",
	Short: "Initialize IDE config sync",
	Long:  "Initialize IDE config sync",
	Run: func(c *cobra.Command, args []string) {
		exists, err := fsutil.Exists(config.StoragePath)
		if err != nil {
			panic(err)
		}

		if !exists {
			err = os.MkdirAll(config.StoragePath, 0755)
			if err != nil {
				color.Red("failed to create storage directory: %s", err)
			}
		}

		cfg, err := config.Load()
		if err != nil {
			color.Red("failed to load config: %s", err)
			return
		}

		cfg.DatabasePath = readDatabasePath()
		localOnly, err := readUseLocalOnly()
		if err != nil {
			color.Red("failed to read input: %s", err)
			return
		}

		cfg.LocalOnly = localOnly
		if !cfg.LocalOnly {
			cfg.DatabaseRepoURL = readDatabaseRepositoryURL(cfg)
		}

		dbPathExists, _ := fsutil.Exists(cfg.DatabasePath)
		if !dbPathExists {
			err = os.MkdirAll(cfg.DatabasePath, 0755)
			if err != nil {
				color.Red("failed to create database path: %s", err)
				return
			}
		} else {
			dbPathEmpty, _ := fsutil.IsDirectoryEmpty(cfg.DatabasePath)
			if !dbPathEmpty {
				color.Red("database path must be an empty directory")
				return
			}
		}

		if cfg.LocalOnly {
			_, err = persistance.InitializeFromPath(cfg.DatabasePath)
		} else {
			_, err = persistance.InitializeFromURL(cfg.DatabaseRepoURL, cfg.DatabasePath)
		}
		if err != nil {
			color.Red("failed to create database repository: %s", err)
			return
		}
		color.Green("Database repository created at %s", cfg.DatabasePath)

		err = config.Save(cfg)
		if err != nil {
			color.Red("failed to save config: %s", err)
		}

	},
}
