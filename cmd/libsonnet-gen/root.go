package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"path/filepath"

	"github.com/thisisibrahimd/libsonnet-gen/pkg/compiler/jsonschemacompiler"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/config"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/format"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/model"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/swagger"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/targetgenerator"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/telemetry"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/util"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/writer"
	"github.com/mdobak/go-xerrors"
	jschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/urfave/cli/v3"
)

func NewRootCommand() *cli.Command {
	cmd := &cli.Command{
		Name:        "libsonnet-gen",
		Usage:       "libsonnet-gen [global options]",
		Description: "libsonnet-gen generates Jsonnet libraries from OpenAPI specs, CRDs, or JSON Schema",
	}

	// flags
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:  "target",
		Usage: "target type: kubernetes or jsonschema",
		Validator: func(s string) error {
			if s != "kubernetes" && s != "jsonschema" {
				return fmt.Errorf("invalid target %q, must be kubernetes or jsonschema", s)
			}
			return nil
		},
		Required: true,
	})

	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:  "config",
		Value: "config.json",
		Usage: "json config file (kubernetes target)",
	})

	cmd.Flags = append(cmd.Flags, &cli.BoolFlag{
		Name:  "debug",
		Value: false,
		Usage: "debug logging",
	})

	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:  "schema",
		Value: "schema.json",
		Usage: "jsonschema file. can be url or filepath (jsonschema target)",
		Validator: func(s string) error {
			if s == "" {
				return nil
			}
			u, urlErr := url.Parse(s)
			var isURL bool
			if urlErr == nil && (u.Scheme == "http" || u.Scheme == "https") {
				isURL = true
			}

			fileInfo, fileErr := os.Stat(s)
			var isFile bool
			if fileErr == nil && !fileInfo.IsDir() {
				isFile = true
			}

			if !isFile && !isURL {
				return fmt.Errorf("schema is neither a url nor a file that exists")
			}
			return nil
		},
	})

	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:  "output",
		Usage: "libsonnet output file (jsonschema target)",
	})

	// before: set logger
	cmd.Before = func(ctx context.Context, c *cli.Command) (context.Context, error) {
		opts := slog.HandlerOptions{
			AddSource: false,
			Level:     telemetry.NewLoggingLevel(c.Bool("debug")),
		}
		l := slog.New(slog.NewTextHandler(os.Stdout, &opts))
		slog.SetDefault(l)
		return ctx, nil
	}

	// action
	cmd.Action = func(ctx context.Context, c *cli.Command) error {
		target := c.String("target")

		switch target {
		case "kubernetes":
			return runKubernetes(c)
		case "jsonschema":
			return runJSONSchema(c)
		default:
			return fmt.Errorf("unknown target %q", target)
		}
	}

	return cmd
}

func runKubernetes(c *cli.Command) error {
	configFile := c.String("config")
	absConfigFile, err := filepath.Abs(configFile)
	if err != nil {
		panic(err)
	}
	configDir := filepath.Dir(absConfigFile)
	if err := os.Chdir(configDir); err != nil {
		panic(err)
	}

	cfg, err := config.Load(absConfigFile)
	if err != nil {
		panic(err)
	}
	err = config.Validate(cfg)
	if err != nil {
		panic(err)
	}
	slog.Debug("loaded config file", slog.String("file", configFile))

	if cfg.SpecGenerator != nil {
		tg, err := targetgenerator.New(*cfg.SpecGenerator)
		if err != nil {
			return xerrors.New("failed to create target generator", err)
		}

		specs, err := tg.GenerateTargets()
		if err != nil {
			return xerrors.New("failed to generate targets", err)
		}
		cfg.Specs = specs
	}

	args := c.Args().Slice()
	if len(args) > 0 {
		slog.Warn("filtering generation to listed versions", slog.Any("versions", args))
	}

	for _, t := range cfg.Specs {
		if len(args) > 0 && !util.HasStr(args, t.Output) {
			slog.Debug("skipping version", slog.String("version", t.Output))
			continue
		}

		prefix := ""
		if cfg.SpecGenerator != nil {
			prefix = cfg.SpecGenerator.Prefix
		}
		if t.Prefix != "" {
			prefix = t.Prefix
		}

		swaggerDefs := make(swagger.Definitions)
		if len(t.Crds) > 0 {
			for _, url := range t.Crds {
				slog.Info(
					"generating spec",
					slog.String("version", t.Output),
					slog.String("spec", url),
					slog.String("prefix", prefix),
				)

				loadedDefs, err := swagger.Load(&swagger.CRDLoader{}, url)
				if err != nil {
					return xerrors.New("unable to load spec", err)
				}
				maps.Copy(swaggerDefs, loadedDefs)
			}
		} else {
			slog.Info(
				"generating spec",
				slog.String("version", t.Output),
				slog.String("spec", t.Openapi),
				slog.String("prefix", prefix),
			)

			loadedDefs, err := swagger.Load(&swagger.SwaggerLoader{}, t.Openapi)
			if err != nil {
				return xerrors.New("unable to load spec", err)
			}
			swaggerDefs = loadedDefs
		}

		groups := model.Load(&swaggerDefs, prefix)
		path := filepath.Join(cfg.OutputDir, t.Output)

		diskWriter := writer.NewDiskWriter()
		if err := diskWriter.Render(path, groups, t, cfg.LibName, cfg.Description); err != nil {
			return xerrors.New("failed to write libsonnet files", err)
		}
	}

	return nil
}

func runJSONSchema(c *cli.Command) error {
	jsonschemaFile := c.String("schema")
	slog.Debug("generating libsonnet library from jsonschema", slog.String("schema", jsonschemaFile))

	comp := jschema.NewCompiler()

	l, _ := jsonschemacompiler.NewLoader(false, "")
	comp.UseLoader(l)
	slog.Debug("configured loader based on schema")

	sch, err := comp.Compile(jsonschemaFile)
	if err != nil {
		return xerrors.New("error compiling schema", err)
	}
	slog.Debug("complied schema", slog.Int("version", sch.DraftVersion))

	libsonnetFile := jsonschemacompiler.CompileLibsonnet(sch, "schema", []string{})
	slog.Debug("generated libsonnet files")

	formattedLibsonnetFile, err := format.Format("", libsonnetFile.String())
	if err != nil {
		return err
	}
	slog.Debug("formatted libsonnet file")

	if c.String("output") == "" {
		fmt.Print(formattedLibsonnetFile)
	} else {
		err = os.WriteFile(c.String("output"), []byte(formattedLibsonnetFile), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}
