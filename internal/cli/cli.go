package cli

import (
	"fmt"
	"os"

	"github.com/bitjungle/gopca/internal/version"
	cli "github.com/urfave/cli/v2"
)

const (
	AppName = "gopca-cli"
)

// NewApp creates and configures the CLI application
func NewApp() *cli.App {
	app := &cli.App{
		Name:    AppName,
		Usage:   "Professional-grade PCA (Principal Component Analysis) toolkit",
		Version: version.Get().Short(),
		Authors: []*cli.Author{
			{
				Name:  "GoPCA Team",
				Email: "support@gopca.example.com",
			},
		},
		Description: `GoPCA is the definitive Principal Component Analysis (PCA) application.
A focused, professional-grade tool that excels at one thing: PCA analysis.

This CLI tool provides fast, scriptable PCA for power users and automation.

QUICK START:
  Analyze a CSV file:     gopca-cli analyze data.csv
  With options:           gopca-cli analyze --scale standard -c 3 data.csv
  Save results:           gopca-cli analyze -f json -o results/ data.csv
  Validate data first:    gopca-cli validate data.csv
  Apply saved model:      gopca-cli transform model.json new_data.csv

For detailed help on any command, use: gopca-cli <command> --help`,
		Commands: []*cli.Command{
			analyzeCommand(),
			validateCommand(),
			transformCommand(),
			versionCommand(),
		},
		Before: func(c *cli.Context) error {
			// If no command is provided, show help
			if c.NArg() == 0 && c.Command.Name == "" {
				_ = cli.ShowAppHelp(c)
				os.Exit(0)
			}
			return nil
		},
		CommandNotFound: func(c *cli.Context, command string) {
			_, _ = fmt.Fprintf(c.App.Writer, "Unknown command '%s'. Try '%s help'\n", command, c.App.Name)
		},
	}

	// Custom help template
	cli.AppHelpTemplate = `NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{.Copyright}}{{end}}
`

	return app
}

// Run executes the CLI application
func Run(args []string) error {
	app := NewApp()
	return app.Run(args)
}

// RunWithOSExit runs the CLI and exits with appropriate code
func RunWithOSExit() {
	if err := Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// versionCommand returns the version command
func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Display version information",
		Action: func(c *cli.Context) error {
			info := version.Get()
			fmt.Println(info.String())
			return nil
		},
	}
}
