package image

import (
	"context"
	"fmt"
	"os"
	"strings"

	buildx "github.com/docker/buildx/build"
	"github.com/docker/buildx/driver"
	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/confutil"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"

	_ "github.com/docker/buildx/driver/docker"
)

func runBuildBuildKit(dockerCli command.Cli, options buildOptions) error {
	ctx := appcontext.Context()

	if options.quiet && options.progress != progress.PrinterModeAuto && options.progress != progress.PrinterModeQuiet {
		return errors.Errorf("progress=%s and quiet cannot be used together", options.progress)
	} else if options.quiet {
		options.progress = progress.PrinterModeQuiet
	}

	cacheFrom, err := buildflags.ParseCacheEntry(options.cacheFrom)
	if err != nil {
		return err
	}

	buildOpts := buildx.Options{
		Inputs: buildx.Inputs{
			ContextPath:    options.context,
			DockerfilePath: options.dockerfileName,
			InStream:       os.Stdin,
		},
		BuildArgs:   listToMap(options.buildArgs.GetAllOrEmpty()),
		CacheFrom:   cacheFrom,
		ExtraHosts:  options.extraHosts.GetAllOrEmpty(),
		ImageIDFile: options.imageIDFile,
		Labels:      listToMap(options.labels.GetAllOrEmpty()),
		NetworkMode: options.networkMode,
		NoCache:     options.noCache,
		Pull:        options.pull,
		ShmSize:     options.shmSize,
		Tags:        options.tags.GetAllOrEmpty(),
		Target:      options.target,
		Ulimits:     options.ulimits,
	}

	const defaultTargetName = "default"
	const drivername = "buildx_buildkit_default"

	d, err := driver.GetDriver(ctx, drivername, nil, dockerCli.Client(), dockerCli.ConfigFile(), nil, nil, nil, nil, nil, "")
	if err != nil {
		return err
	}
	driverInfo := []buildx.DriverInfo{
		{
			Name:   drivername,
			Driver: d,
		},
	}

	progressCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := progress.NewPrinter(progressCtx, os.Stdout, "auto")

	resp, err := buildx.Build(ctx, driverInfo, map[string]buildx.Options{defaultTargetName: buildOpts}, nil, confutil.ConfigDir(dockerCli), w)
	errp := w.Wait()
	if err == nil {
		err = errp
	}
	if err != nil {
		return err
	}

	if options.quiet {
		fmt.Fprint(dockerCli.Out(), resp[defaultTargetName].ExporterResponse["containerimage.digest"])
	}

	return nil
}

func listToMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		kv := strings.SplitN(value, "=", 2)
		if len(kv) == 1 {
			result[kv[0]] = ""
		} else {
			result[kv[0]] = kv[1]
		}
	}
	return result
}
