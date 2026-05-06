package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/casapps/caspbx/src/config"
	"github.com/casapps/caspbx/src/server"
)

var (
	Version      = "dev"
	CommitID     = "unknown"
	BuildDate    = "unknown"
	OfficialSite = ""
)

const projectDescription = "Asterisk communications platform bootstrap"

func main() {
	run(os.Args[1:], filepath.Base(os.Args[0]), os.Stdout, os.Stderr)
}

func run(args []string, binaryName string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(binaryName, flag.ContinueOnError)
	fs.SetOutput(stderr)

	showHelp := fs.Bool("help", false, "Show help")
	showVersion := fs.Bool("version", false, "Show version")
	showStatus := fs.Bool("status", false, "Show server status and health")

	appModeValue := fs.String("mode", "production", "Application mode")
	fs.String("config", "", "Config directory")
	fs.String("data", "", "Data directory")
	fs.String("cache", "", "Cache directory")
	fs.String("log", "", "Log directory")
	fs.String("backup", "", "Backup directory")
	fs.String("pid", "", "PID file path")
	fs.String("address", "0.0.0.0", "Listen address")
	fs.Int("port", 64580, "Listen port")
	fs.String("baseurl", "/", "URL path prefix")
	fs.Bool("daemon", false, "Run as daemon")
	debugValue := fs.Bool("debug", false, "Enable debug mode")
	fs.String("color", "auto", "Color output")
	fs.String("lang", "auto", "Language for output")
	fs.String("service", "", "Service management")
	fs.String("maintenance", "", "Maintenance operations")
	fs.String("update", "", "Check or perform updates")
	fs.String("shell", "", "Shell integration")

	if err := fs.Parse(args); err != nil {
		printHelp(stdout, binaryName)
		return 2
	}

	visitedFlags := map[string]bool{}
	fs.Visit(func(flagValue *flag.Flag) {
		visitedFlags[flagValue.Name] = true
	})

	currentAppMode := config.ResolveAppMode(*appModeValue, visitedFlags["mode"], os.Getenv("MODE"))
	debugEnabled := config.ResolveDebugEnabled(*debugValue, visitedFlags["debug"], os.Getenv("DEBUG"))

	switch {
	case *showHelp:
		printHelp(stdout, binaryName)
		return 0
	case *showVersion:
		fmt.Fprintf(stdout, "%s %s (%s)\n", binaryName, Version, CommitID)
		return 0
	case *showStatus:
		app, _ := server.NewApp(server.DefaultAPIVersion, config.DefaultConfig().Server.AdminPath, binaryName, Version, CommitID, OfficialSite)
		fmt.Fprintln(stdout, "bootstrap status: ready")
		fmt.Fprintln(stdout, app.Summary())
		return 0
	default:
		app, _ := server.NewApp(server.DefaultAPIVersion, config.DefaultConfig().Server.AdminPath, binaryName, Version, CommitID, OfficialSite)
		printStartupMessage(stdout, binaryName, currentAppMode, debugEnabled, app)
		return 0
	}
}

func printHelp(w io.Writer, binaryName string) {
	fmt.Fprintf(w, `%s %s - %s

Usage:
  %s [flags]

Information:
  --help       Show help
  --version    Show version
  --status     Show server status and health

Server Configuration:
  --mode       Application mode
  --config     Config directory
  --data       Data directory
  --cache      Cache directory
  --log        Log directory
  --backup     Backup directory
  --pid        PID file path
  --address    Listen address
  --port       Listen port
  --baseurl    URL path prefix
  --daemon     Run as daemon
  --debug      Enable debug mode
  --color      Color output
  --lang       Language for output

Management:
  --service        Service management
  --maintenance    Maintenance operations
  --update         Check or perform updates
  --shell          Shell integration
`, binaryName, Version, projectDescription, binaryName)
}

func printStartupMessage(w io.Writer, binaryName string, appMode config.AppMode, debugEnabled bool, app server.App) {
	fmt.Fprintf(w, "%s bootstrap scaffold is active.\n", binaryName)
	fmt.Fprintf(w, "Running in mode: %s\n", config.FormatAppModeLabel(appMode, debugEnabled))
	fmt.Fprintln(w, "Repository foundations are in place and the HTTP runtime scaffold is ready.")
	fmt.Fprintln(w, app.Summary())
	if OfficialSite != "" {
		fmt.Fprintf(w, "Official site: %s\n", OfficialSite)
	}
}
