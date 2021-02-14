package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	zemmExePath    string
	zemmPluginDirs string
	zemmPWD        string
)

func createDynamicCommand(executeable string) *cobra.Command {
	e := path.Base(executeable)
	cmdName := e[5:]
	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s", cmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdRun := exec.Command(executeable, args...)
			cmdRun.Env = append(os.Environ(), fmt.Sprintf("ZEMM=%s", zemmExePath))
			cmdRun.Env = append(cmdRun.Env, fmt.Sprintf("ZEMM_PLUGINDIRS=%s", zemmPluginDirs))
			cmdRun.Env = append(cmdRun.Env, fmt.Sprintf("ZEMM_PWD=%s", zemmPWD))
			cmdRun.Stdout = os.Stdout
			cmdRun.Stderr = os.Stderr
			cmdRun.Stdin = os.Stdin

			return cmdRun.Run()
		},
	}

	// Disable Flag parsing
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func main() {
	// Get exe path and PWD
	var err error
	zemmExePath, err = os.Executable()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	zemmPWD, err = os.Getwd()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use: path.Base(zemmExePath),

		// Stolen from wash: https://github.com/puppetlabs/wash/pull/60
		// Need to set these so that Cobra will not output the usage +
		// error object when Execute() returns an error.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Use predefined plugindirs by them from env
	zemmPluginDirs = "/usr/lib/zemm:/usr/local/lib/zemm"
	if dirs, ok := os.LookupEnv("ZEMM_PLUGINDIRS"); ok && dirs != "" {
		zemmPluginDirs = dirs
	}

	// Parse plugindirs and find plugins
	dirs := strings.Split(zemmPluginDirs, ":")
	pluginPaths := []string{}
	for _, d := range dirs {
		matches, err := filepath.Glob(fmt.Sprintf("%s/zemm-*", d))
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		pluginPaths = append(pluginPaths, matches...)
	}

	if len(pluginPaths) == 0 {
		fmt.Printf("ERROR: No plugins found in %s\n", zemmPluginDirs)
		os.Exit(1)
	}

	// Add commands
	for _, m := range pluginPaths {
		e := path.Base(m)
		if strings.Contains(e[5:], "-") {
			// Skip commands with a dash, those are subcommandss
			continue
		}

		rootCmd.AddCommand(createDynamicCommand(m))
	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}

		os.Exit(144)
	}
}
