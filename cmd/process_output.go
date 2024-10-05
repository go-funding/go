package main

import (
	"context"
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/ctx"
	"fuk-funding/go/database"
	"fuk-funding/go/database/dbtypes"
	"fuk-funding/go/fp"
	"fuk-funding/go/services"
	"fuk-funding/go/utils"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

// Simple checker
var _ = CommandRunnable(ProcessOutput{})

type ProcessOutput struct {
	runner *ProcessOutputRunner
}

func (acc ProcessOutput) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "process_output",
		Usage: "Process output from a file",
		Flags: []cli.Flag{
			flag.OutputDir.Flag,
		},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (acc ProcessOutput) Run(appCtx *ctx.Context, cliCtx *cli.Context) error {
	outputDir := flag.OutputDir.CliParse(cliCtx)

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(cliCtx))
	if err != nil {
		return err
	}

	if err = db.Connect(); err != nil {
		return err
	}

	defer db.Close()

	acc.runner = &ProcessOutputRunner{
		Log:            appCtx.Logger.Named("process_output"),
		Sql:            db,
		FlagsService:   services.NewFlagsService(appCtx.Logger, db),
		DomainsService: services.NewDomainsService(appCtx.Logger, db),
	}

	dirPaths := fp.Map(entries, func(entry os.DirEntry, _ int) string {
		return filepath.Join(outputDir, entry.Name())
	})

	return fp.ForEachErr1(dirPaths, cliCtx.Context, acc.runner.Run)
}

type ProcessOutputRunner struct {
	Log *zap.SugaredLogger
	Sql dbtypes.Sql

	FlagsService   *services.Flags
	DomainsService *services.Domains
}

func (pr ProcessOutputRunner) Run(ctx context.Context, dirPath string, _ int) error {
	dirName := fp.SliceAt(strings.Split(dirPath, "/"), -1)
	websiteUrl := utils.DirnameHost(dirName)
	pr.Log.Infof("Processing %s as URL: %s", color.GreenString(dirPath), color.YellowString(websiteUrl))

	err := pr.DomainsService.UpsertNewHost(ctx, websiteUrl)
	if err != nil {
		return err
	}

	domainID, err := pr.DomainsService.GetDomainID(ctx, websiteUrl)
	if err != nil {
		return err
	}

	err = pr.processFlags(ctx, domainID, dirPath, []func(context.Context, int, string) error{
		pr.processFlagNextJsManifest,
		pr.processSourceMapsExposure,
		pr.processWPDirectories,
	})
	if err != nil {
		return err
	}

	return nil
}

func (pr ProcessOutputRunner) processFlags(ctx context.Context, domainId int, dirname string, flags []func(context.Context, int, string) error) error {
	for _, flagProcessor := range flags {
		err := flagProcessor(ctx, domainId, dirname)
		if err != nil {
			return err
		}
	}
	return nil
}

const sourceMapsExposure = "source-maps-exposure"

func (pr ProcessOutputRunner) processSourceMapsExposure(ctx context.Context, domainId int, dirname string) error {
	log := pr.Log.Named(color.BlueString(sourceMapsExposure))

	has, err := pr.FlagsService.HasFlag(ctx, domainId, sourceMapsExposure)
	if err != nil {
		log.Error("error", err)
		return err
	}

	if has {
		return nil
	}

	var found bool
	err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".map") {
			log.Infof("found %s", color.GreenString(path))
			found = true
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		log.Error("error while walking", err)
		return err
	}

	if found {
		err = pr.FlagsService.UpsertFlag(ctx, domainId, sourceMapsExposure, "")
	}

	return nil
}

const flagCodeNextJsManifest = "nextjs-build-manifest"

func (pr ProcessOutputRunner) processFlagNextJsManifest(ctx context.Context, domainId int, dirname string) error {
	log := pr.Log.Named(color.BlueString(flagCodeNextJsManifest))

	has, err := pr.FlagsService.HasFlag(ctx, domainId, flagCodeNextJsManifest)
	if err != nil {
		log.Error("error", err)
		return err
	}

	if has {
		return nil
	}

	path, err := findBuildManifestFile(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("not found")
			return nil
		}

		log.Error("error while walking", err)
		return err
	}

	log.Infof("found %s", color.GreenString(path))

	err = pr.FlagsService.UpsertFlag(ctx, domainId, flagCodeNextJsManifest, path)
	if err != nil {
		log.Error("error", err)
		return err
	}

	return nil
}

func findBuildManifestFile(root string) (string, error) {
	var manifestPath string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "_buildManifest.js" {
			manifestPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if manifestPath == "" {
		return "", os.ErrNotExist
	}
	return manifestPath, err
}

const flagWPDirectories = "wordpress-directories"

func (pr ProcessOutputRunner) processWPDirectories(ctx context.Context, domainId int, dirname string) error {
	log := pr.Log.Named(color.BlueString(flagWPDirectories))

	has, err := pr.FlagsService.HasFlag(ctx, domainId, flagWPDirectories)
	if err != nil {
		log.Error("error", err)
		return err
	}

	if has {
		return nil
	}

	wpDirs := []string{"wp-admin", "wp-includes", "wp-content"}
	foundDirs := make([]string, 0)

	err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, wpDir := range wpDirs {
				if info.Name() == wpDir {
					log.Infof("found %s", color.GreenString(path))
					foundDirs = append(foundDirs, path)
					return filepath.SkipDir
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Error("error while walking", err)
		return err
	}

	if len(foundDirs) > 0 {
		dirList := strings.Join(foundDirs, ", ")
		err = pr.FlagsService.UpsertFlag(ctx, domainId, flagWPDirectories, dirList)
		if err != nil {
			log.Error("error", err)
			return err
		}
	}

	return nil
}
