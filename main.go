package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"ndlocgo/internal/localization"
)

const appName = "NDLocGo"

func usage() {
	fmt.Printf(`%s — Naughty Dog Localization Tool (Go port)

Drag and drop:
  • Drop a localization file (.subtitles, .common, .subtitles-systemic, .web)
    onto %s.exe to export it.
  • Drop an exported .txt file onto %s.exe to import it.

Command line:
  %s <file.subtitles|file.common|file.subtitles-systemic|file.web>
  %s <exported-file.txt>

Export creates <file>.txt and <file>.ids.
Import creates <base-name>.<extension>.new.

`, appName, appName, appName, appName, appName)
}

func run(input string) error {
	ext := strings.ToLower(filepath.Ext(input))
	switch ext {
	case ".subtitles", ".common", ".subtitles-systemic", ".web":
		result, err := localization.Export(input)
		if err != nil {
			return err
		}
		fmt.Printf("Export completed.\nFormat version: %d\nString count  : %d\nTXT           : %s\nIDS           : %s\n",
			result.Version, result.Count, result.TextPath, result.IDsPath)
		return nil
	case ".txt":
		result, err := localization.Import(input)
		if err != nil {
			return err
		}
		fmt.Printf("Import completed.\nFormat version: %d\nString count  : %d\nOutput        : %s\n",
			result.Version, result.Count, result.OutputPath)
		fmt.Println("Note: the .new output does not automatically replace the original game file.")
		return nil
	default:
		return fmt.Errorf("unsupported extension %q", ext)
	}
}

func main() {
	if len(os.Args) != 2 {
		usage()
		if len(os.Args) > 2 {
			os.Exit(2)
		}
		return
	}

	input := strings.TrimSpace(os.Args[1])
	if input == "" {
		usage()
		os.Exit(2)
	}
	if _, err := os.Stat(input); err != nil {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", input)
		pauseOnWindows()
		os.Exit(1)
	}
	if err := run(input); err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR: %v\n", err)
		fmt.Println("File asli tidak diubah.")
		pauseOnWindows()
		os.Exit(1)
	}
	pauseOnWindows()
}

func pauseOnWindows() {
	if runtime.GOOS != "windows" {
		return
	}
	// Drag-and-drop from Explorer launches a console process that may close
	// immediately; wait so the user can read the result.
	fmt.Print("\nPress Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
