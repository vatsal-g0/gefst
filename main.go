package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// colors
const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

type Node struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content,omitempty"`
	Children []Node `json:"children,omitempty"`
}

type Config struct {
	InputFile string
	Root      string
	DryRun    bool
	Overwrite bool
	Verbose   bool
	Validate  bool
	Quiet     bool
	NoColor   bool
	Version   bool
}

type Stats struct {
	Files       int
	Folders     int
	Skipped     int
	Overwritten int
}

// ---------- FLAGS ----------

func parseFlags() Config {
	root := flag.String("root", ".", "output directory")
	dryRun := flag.Bool("dry-run", false, "preview without creating files")
	overwrite := flag.Bool("overwrite", false, "overwrite existing files")
	verbose := flag.Bool("verbose", false, "show detailed output")
	validate := flag.Bool("validate", false, "validate config only")
	quiet := flag.Bool("quiet", false, "suppress all output")
	noColor := flag.Bool("no-color", false, "disable colored output")
	version := flag.Bool("version", false, "show version")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("usage: gefst [options] <config.json>")
		fmt.Println("options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	return Config{
		InputFile: args[0],
		Root:      *root,
		DryRun:    *dryRun,
		Overwrite: *overwrite,
		Verbose:   *verbose,
		Validate:  *validate,
		Quiet:     *quiet,
		NoColor:   *noColor,
		Version:   *version,
	}
}

// ---------- UTILS ----------

func fail(cfg Config) {
	if !cfg.Quiet {
		if cfg.NoColor {
			fmt.Println("error occurred. please check your input.")
		} else {
			fmt.Println(colorRed + "error occurred. please check your input." + colorReset)
		}
	}
	os.Exit(1)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ---------- COLOR ----------

func colorize(cfg Config, kind string) string {
	if cfg.NoColor {
		return kind
	}

	switch kind {
	case "file":
		return colorGreen + kind + colorReset
	case "folder":
		return colorBlue + kind + colorReset
	case "skip":
		return colorGray + kind + colorReset
	case "overwrite":
		return colorRed + kind + colorReset
	default:
		return kind
	}
}

// ---------- LOGGING ----------

func log(cfg Config, kind, path string) {
	if cfg.Quiet {
		return
	}

	if !cfg.Verbose && !cfg.DryRun {
		return
	}

	label := colorize(cfg, kind)

	if !cfg.NoColor {
		if kind == "folder" {
			path = colorBlue + path + "/" + colorReset
		} else {
			path = colorCyan + path + colorReset
		}
	}

	padding := max(8-len(kind), 1)
	fmt.Printf("[%s]%*s%s\n", label, padding, "", path)
}

// ---------- CORE ----------

func genFolder(root Node, path string, cfg Config, stats *Stats) {
	fullPath := filepath.Join(path, root.Name)

	switch root.Type {

	case "folder":
		stats.Folders++

		if cfg.DryRun {
			log(cfg, "folder", fullPath)
		} else {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				fail(cfg)
			}
			log(cfg, "folder", fullPath)
		}

		for _, child := range root.Children {
			genFolder(child, fullPath, cfg, stats)
		}

	case "file":
		if cfg.DryRun {
			log(cfg, "file", fullPath)
			stats.Files++
			return
		}

		alreadyExists := exists(fullPath)

		if alreadyExists && !cfg.Overwrite {
			log(cfg, "skip", fullPath)
			stats.Skipped++
			return
		}

		if err := os.WriteFile(fullPath, []byte(root.Content), 0644); err != nil {
			fail(cfg)
		}

		if alreadyExists {
			log(cfg, "overwrite", fullPath)
			stats.Overwritten++
		} else {
			log(cfg, "file", fullPath)
			stats.Files++
		}

	default:
		fail(cfg)
	}
}

// ---------- VALIDATION ----------

func validateNode(root Node, path string) error {
	if root.Type != "folder" && root.Type != "file" {
		return fmt.Errorf("invalid type")
	}

	if root.Name == "" {
		return fmt.Errorf("missing name")
	}

	currentPath := filepath.Join(path, root.Name)

	if root.Type == "file" && len(root.Children) > 0 {
		return fmt.Errorf("invalid structure")
	}

	if root.Type == "folder" {
		for _, child := range root.Children {
			if err := validateNode(child, currentPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// ---------- MAIN ----------

func main() {
	cfg := parseFlags()
	stats := &Stats{}

	data, err := os.ReadFile(cfg.InputFile)
	if err != nil {
		fail(cfg)
	}

	var root Node
	if err := json.Unmarshal(data, &root); err != nil {
		fail(cfg)
	}

	if cfg.Version {
		fmt.Println(colorBold + "gefst v1.0.0" + colorReset)
		return
	}

	if cfg.Validate {
		if validateNode(root, "") != nil {
			fail(cfg)
		}

		if !cfg.Quiet {
			if cfg.NoColor {
				fmt.Println("validation successful")
			} else {
				fmt.Println(colorGreen + "validation successful" + colorReset)
			}
		}
		return
	}

	genFolder(root, cfg.Root, cfg, stats)

	if !cfg.Quiet {
		fmt.Println("\ngenerated:")

		rootName := root.Name
		if !cfg.NoColor {
			rootName = colorBold + rootName + colorReset
		}

		fmt.Printf("root: %s\n", rootName)

		files := fmt.Sprintf("%d", stats.Files)
		folders := fmt.Sprintf("%d", stats.Folders)
		skipped := fmt.Sprintf("%d", stats.Skipped)
		overwritten := fmt.Sprintf("%d", stats.Overwritten)

		if !cfg.NoColor {
			files = colorGreen + files + colorReset
			folders = colorBlue + folders + colorReset
			skipped = colorYellow + skipped + colorReset
			overwritten = colorRed + overwritten + colorReset
		}

		fmt.Printf("%s file(s)\n", files)
		fmt.Printf("%s folder(s)\n", folders)
		fmt.Printf("%s skipped\n", skipped)
		fmt.Printf("%s overwritten\n", overwritten)
	}
}
