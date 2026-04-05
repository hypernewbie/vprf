package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runRecord(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputPath := fs.String("output", "profile.json.gz", "Output profile path")
	fs.StringVar(outputPath, "o", "profile.json.gz", "Output profile path")
	duration := fs.Float64("duration", 0, "Optional duration in seconds")
	rate := fs.Float64("rate", 1000, "Sampling rate in Hz")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cmdArgs := fs.Args()
	if len(cmdArgs) == 0 {
		return fmt.Errorf("record requires a command after flags")
	}
	samplyPath, err := findSamply()
	if err != nil {
		return err
	}

	argv := []string{"record", "--save-only", "--unstable-presymbolicate", "-o", *outputPath, "--rate", fmt.Sprintf("%.0f", *rate)}
	if *duration > 0 {
		argv = append(argv, "--duration", fmt.Sprintf("%g", *duration))
	}
	argv = append(argv, "--")
	argv = append(argv, cmdArgs...)

	cmd := exec.Command(samplyPath, argv...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("samply record failed: %w", err)
	}
	fmt.Fprintf(stdout, "profile saved to %s\n", *outputPath)
	fmt.Fprintf(stdout, "command: samply %s\n", strings.Join(argv, " "))
	return nil
}

func findSamply() (string, error) {
	if path, err := exec.LookPath("samply"); err == nil {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("samply not found on PATH")
	}
	fallback := filepath.Join(home, ".cargo", "bin", "samply")
	if _, err := os.Stat(fallback); err == nil {
		return fallback, nil
	}
	return "", fmt.Errorf("samply not found on PATH or at %s", fallback)
}
