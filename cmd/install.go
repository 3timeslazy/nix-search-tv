package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"
)

var Install = &cli.Command{
	Name:      "install",
	Usage:     "Install a package",
	UsageText: "nix-search-tv install <channel> <package>",
	Action:    InstallAction,
	Flags:     BaseFlags(),
}

func InstallAction(_ context.Context, cmd *cli.Command) error {
	fullPkgName := strings.Join(cmd.Args().Slice(), " ")
	if fullPkgName == "" {
		return errors.New("package name is required")
	}

	conf, err := GetConfig(cmd)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	if cmd.IsSet(IndexesFlag) {
		conf.Indexes = cmd.StringSlice(IndexesFlag)
	}

	_, err = SetupIndexes(conf)
	if err != nil {
		return err
	}

	var pkgName string

	if len(conf.Indexes) == 1 {
		pkgName = fullPkgName
	} else {
		var ok bool
		_, pkgName, ok = cutIndexPrefix(fullPkgName)
		if !ok {
			return errors.New("multiple indexes requested, but the package has no index prefix")
		}
	}

	fmt.Printf("Installing %s\n", pkgName)

	if err := writePackageToConfig(pkgName); err != nil {
		fmt.Printf("Error writing to config: %v\n", err)
		return nil
	}

	// current process should run as sudo
	cmd2 := exec.Command("nixos-rebuild", "switch")
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		fmt.Printf("Error rebuilding: %v\n", err)
	}

	return nil
}

func getLineToWrite(path string, pkgName string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	startLine := -1
	foundBlock := false

	// 1. Find the start of the block
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "environment.systemPackages =") && strings.HasSuffix(trimmed, "[") {
			startLine = i
			foundBlock = true
			break
		}
	}

	if !foundBlock {
		return 0, fmt.Errorf("could not find environment.systemPackages block")
	}

	// 2. Read next lines one by one
	for i := startLine + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// 3. Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 5. Check if package matching the trimmed line
		if trimmed == pkgName {
			return 0, fmt.Errorf("package '%s' is already installed", pkgName)
		}

		// 4. Check for closing bracket (end of list)
		if trimmed == "];" || trimmed == "]" {
			return i, nil
		}
	}

	return 0, fmt.Errorf("could not find closing bracket for environment.systemPackages")
}

func writePackageToConfig(pkgName string) error {
	path := "/etc/nixos/configuration.nix"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("configuration.nix not found")
	}

	lineNum, err := getLineToWrite(path, pkgName)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	line := lines[lineNum]

	// Determine indentation (heuristic: same as indentation of closing bracket + 2 spaces)
	indent := ""
	for _, r := range line {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}
	newIndent := indent + "  "

	// Construct the new line
	newLine := newIndent + pkgName

	// Insert the new line into the slice
	// We need to insert at index lineNum
	newLines := append(lines[:lineNum], append([]string{newLine}, lines[lineNum:]...)...)

	// Join lines and write back to file
	newContent := strings.Join(newLines, "\n")
	if err = os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return err
	}

	fmt.Println("Successfully added package to configuration.nix")
	return nil
}
