package models

import (
	"log/slog"
	"nw-buddy/tools/formats/gltf/importer"
	"nw-buddy/tools/utils"
	"nw-buddy/tools/utils/logging"
	"nw-buddy/tools/utils/math/mat4"
	"strings"

	"github.com/spf13/cobra"
)

// models slices: converte dynamicslices inteiras (ex.: uma cidade "complete") na origem, um glb por slice.
var cmdCollectSlices = &cobra.Command{
	Use:   "slices",
	Short: "converts whole .dynamicslice files (e.g. a complete settlement) to gltf at the origin",
	Run:   runCollectSlices,
}

func init() {
	cmdCollectSlices.Flags().AddFlagSet(Cmd.Flags())
}

func runCollectSlices(ccmd *cobra.Command, args []string) {
	slog.SetDefault(logging.DefaultFileHandler())
	c := utils.Must(initCollector())
	for i := range args {
		args[i] = strings.ToLower(args[i])
	}
	files, err := c.Archive.Glob(args...)
	if err != nil || len(files) == 0 {
		slog.Error("no slices found", "err", err)
		return
	}
	for _, file := range files {
		group := importer.AssetGroup{}
		c.collectSliceInto(&group, file, mat4.Identity())
		if len(group.Geometries) == 0 {
			slog.Warn("slice without geometry", "file", file.Path())
			continue
		}
		group.TargetFile = c.outputPath(utils.ReplaceExt(file.Path(), ""))
		c.models.Store(group.TargetFile, group)
	}
	c.Process()
	slog.SetDefault(logging.DefaultTerminalHandler())
}
