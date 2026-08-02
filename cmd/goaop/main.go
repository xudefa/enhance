package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/xudefa/enhance/aop/generator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "gen", "generate":
		runGenerate()
	case "clean":
		runClean()
	case "validate":
		runValidate()
	case "version":
		fmt.Println("enhance aop generator v0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("enhance-aop - AOP code generator for Go (idiomatic Go style)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  enhance-aop <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  gen, generate    Generate AOP proxy code from //go:generate directives")
	fmt.Println("  clean            Clean generated AOP proxy code")
	fmt.Println("  validate         Validate AOP annotations and directives")
	fmt.Println("  version          Show version")
	fmt.Println()
	fmt.Println("Generation Modes:")
	fmt.Println("  simple   Simple delegation proxy (no AOP)")
	fmt.Println("  aop      Runtime AOP with reflection")
	fmt.Println("  static   Static AOP proxy (zero reflection, recommended)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Scan current directory for //go:generate directives")
	fmt.Println("  enhance-aop gen")
	fmt.Println()
	fmt.Println("  # Scan specific directory with static mode")
	fmt.Println("  enhance-aop gen -dir ./service -mode static")
	fmt.Println()
	fmt.Println("  # Generate for specific types")
	fmt.Println("  enhance-aop gen -type UserService,OrderService")
	fmt.Println()
	fmt.Println("  # Generate for specific interfaces")
	fmt.Println("  enhance-aop gen -interface ServiceInterface")
	fmt.Println()
	fmt.Println("Usage in code (//go:generate directives):")
	fmt.Println("  //go:generate enhance aop gen -type=UserService")
	fmt.Println("  //go:generate enhance aop gen -type=UserService,OrderService -mode=static")
	fmt.Println("  //go:generate enhance aop gen -interface=ServiceInterface")
	fmt.Println("  //go:generate enhance aop gen -all")
	fmt.Println()
	fmt.Println("Then run: go generate ./...")
}

func runGenerate() {
	dir := "."
	mode := "static"
	typeFlag := ""
	interfaceFlag := ""
	output := ""

	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.StringVar(&dir, "dir", ".", "Directory to scan")
	fs.StringVar(&mode, "mode", "static", "Generation mode: simple, aop, static")
	fs.StringVar(&typeFlag, "type", "", "Comma-separated type names to generate proxies for")
	fs.StringVar(&interfaceFlag, "interface", "", "Comma-separated interface names to generate proxies for")
	fs.StringVar(&output, "output", "", "Output file path (default: *_aop.go)")
	_ = fs.Parse(os.Args[2:])

	slog.Info("enhance-aop: generating AOP proxies",
		"dir", dir,
		"mode", mode,
	)

	// 检查是否有 //go:generate 指令
	directives, err := scanGoGenerateDirectives(dir)
	if err != nil {
		slog.Warn("enhance-aop: failed to scan directives, falling back to annotation mode", "error", err)
	}

	if len(directives) > 0 {
		// 使用 //go:generate 指令模式
		if err := generateFromDirectives(dir, directives, mode, output); err != nil {
			slog.Error("enhance-aop: failed to generate from directives", "error", err)
			os.Exit(1)
		}
		fmt.Println("AOP proxies generated from //go:generate directives!")
	} else {
		// 回退到注解模式
		if typeFlag != "" || interfaceFlag != "" {
			// 命令行指定类型/接口
			gen, err := generator.NewGenerator()
			if err != nil {
				slog.Error("enhance-aop: failed to create generator", "error", err)
				os.Exit(1)
			}

			opts := generator.GenerateOptions{Output: output}
			if typeFlag != "" {
				for _, t := range strings.Split(typeFlag, ",") {
					t = strings.TrimSpace(t)
					if t == "" {
						continue
					}
					opts.Types = append(opts.Types, t)
					slog.Info("enhance-aop: generating proxy for type", "type", t)
				}
			}
			if interfaceFlag != "" {
				for _, iface := range strings.Split(interfaceFlag, ",") {
					iface = strings.TrimSpace(iface)
					if iface == "" {
						continue
					}
					opts.Interfaces = append(opts.Interfaces, iface)
					slog.Info("enhance-aop: generating proxy for interface", "interface", iface)
				}
			}

			if err := gen.Generate(dir, mode, opts); err != nil {
				slog.Error("enhance-aop: failed to generate proxies", "error", err)
				os.Exit(1)
			}

			registry := gen.GetRegistry()
			list := registry.List()
			slog.Info("enhance-aop: generated proxies", "count", len(list))
		} else {
			// 扫描注解模式
			gen, err := generator.NewGenerator()
			if err != nil {
				slog.Error("enhance-aop: failed to create generator", "error", err)
				os.Exit(1)
			}

			if err := gen.Generate(dir, mode); err != nil {
				slog.Error("enhance-aop: failed to generate proxies", "error", err)
				os.Exit(1)
			}

			registry := gen.GetRegistry()
			list := registry.List()
			slog.Info("enhance-aop: generated proxies", "count", len(list))
		}
		fmt.Println("AOP proxies generated successfully!")
	}

	fmt.Println("Build with: go build")
}

// scanGoGenerateDirectives 扫描目录中的 //go:generate 指令
func scanGoGenerateDirectives(dir string) ([]*generator.GoGenerateDirective, error) {
	var directives []*generator.GoGenerateDirective

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_aop.go") {
			return nil
		}

		// 读取文件查找 //go:generate 指令
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "//go:generate") && strings.Contains(line, "enhance aop gen") {
				directive, err := generator.ParseGoGenerate(line)
				if err != nil {
					slog.Warn("enhance-aop: failed to parse directive", "line", line, "error", err)
					continue
				}
				if directive.HasTargets() {
					directives = append(directives, directive)
				}
			}
		}

		return nil
	})

	return directives, err
}

// generateFromDirectives 从 //go:generate 指令生成代理代码
func generateFromDirectives(dir string, directives []*generator.GoGenerateDirective, defaultMode string, defaultOutput string) error {
	gen, err := generator.NewGenerator()
	if err != nil {
		return fmt.Errorf("create generator: %w", err)
	}

	for _, d := range directives {
		mode := d.Mode
		if mode == "" {
			mode = defaultMode
		}

		output := d.Output
		if output == "" {
			output = defaultOutput
		}

		slog.Info("enhance-aop: processing directive",
			"types", d.Types,
			"interfaces", d.Interfaces,
			"mode", mode,
		)

		if err := gen.Generate(dir, mode, generator.GenerateOptions{
			Types:      d.Types,
			Interfaces: d.Interfaces,
			Output:     output,
		}); err != nil {
			return fmt.Errorf("generate proxy: %w", err)
		}
	}

	return nil
}

func runClean() {
	dir := "."
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	fs.StringVar(&dir, "dir", ".", "Directory to clean")
	_ = fs.Parse(os.Args[2:])

	slog.Info("enhance-aop: cleaning generated proxies", "dir", dir)

	gen, err := generator.NewGenerator()
	if err != nil {
		slog.Error("enhance-aop: failed to create generator", "error", err)
		os.Exit(1)
	}

	if err := gen.Clean(dir); err != nil {
		slog.Error("enhance-aop: failed to clean proxies", "error", err)
		os.Exit(1)
	}

	fmt.Println("Generated proxies cleaned successfully!")
}

func runValidate() {
	dir := "."
	mode := "static"
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.StringVar(&dir, "dir", ".", "Directory to validate")
	fs.StringVar(&mode, "mode", "static", "Generation mode: simple, aop, static")
	_ = fs.Parse(os.Args[2:])

	slog.Info("enhance-aop: validating annotations", "dir", dir, "mode", mode)

	gen, err := generator.NewGenerator()
	if err != nil {
		slog.Error("enhance-aop: failed to create generator", "error", err)
		os.Exit(1)
	}

	if err := gen.Generate(dir, mode, generator.GenerateOptions{DryRun: true}); err != nil {
		slog.Error("enhance-aop: validation failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("Annotations validated successfully!")
}
