package models

import (
	"log/slog"
	"nw-buddy/tools/formats/gltf/importer"
	"nw-buddy/tools/game"
	"nw-buddy/tools/rtti/nwt"
	"nw-buddy/tools/utils"
	"nw-buddy/tools/utils/logging"
	"nw-buddy/tools/utils/math/mat4"
	"nw-buddy/tools/utils/progress"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// nw-port: os <PLOT> das camadas de capitals têm o conteúdo pré-gerado (penhascos, rochas, detalhes, água) nos
// arquivos .chunks ao lado do .capitals.json. Cada chunk aponta para um slice numa posição do mundo.
var cmdCollectChunks = &cobra.Command{
	Use:   "chunks",
	Short: "Converts the .chunks files of capitals layers (pre-generated plot content) to gltf",
	Run:   runCollectChunks,
	Example: strings.Trim(`
  nwbt models chunks --level newworld_vitaeeterna --region r_+04_+01 --name 06swindsward_capitals -f out.glb`, "\n"),
}

func init() {
	cmdCollectChunks.Flags().AddFlagSet(Cmd.Flags())
	cmdCollectChunks.Flags().StringVar(&flgLevel, "level", "ftue_v2", "The LEVEL_NAME in the coatlicue directory. Can be a glob expression.")
	cmdCollectChunks.Flags().StringVar(&flgRegion, "region", "**", "The REGION_NAME in the coatlicue directory. Can be a glob expression.")
	cmdCollectChunks.Flags().StringVar(&flgName, "name", "*", "The capitals layer to process. Can be a glob expression.")
	cmdCollectChunks.Flags().StringVarP(&flgOutFile, "file", "f", "", "If set, all chunks are merged into a single file.")
	Cmd.AddCommand(cmdCollectChunks)
}

func runCollectChunks(ccmd *cobra.Command, args []string) {
	glob := path.Join("**", "coatlicue", flgLevel, "regions", flgRegion, "capitals", flgName, "*.chunks")
	slog.SetDefault(logging.DefaultFileHandler())
	c := utils.Must(initCollector())
	c.CollectChunks(glob)
	c.Process()
	slog.SetDefault(logging.DefaultTerminalHandler())
}

func (c *Collector) CollectChunks(glob string) {
	merge := flgOutFile != ""
	files, _ := c.Archive.Glob(glob)
	bar := progress.Bar(len(files), "chunks")
	defer func() {
		bar.Detail("")
		bar.Close()
	}()

	group := importer.AssetGroup{}
	for _, file := range files {
		bar.Detail(file.Path())
		bar.Add(1)
		if !merge {
			group = importer.AssetGroup{}
		}
		obj, err := game.LoadObjectStream(file)
		if err != nil {
			slog.Error("chunks not loaded", "file", file.Path(), "error", err)
			continue
		}
		chunks, ok := obj.(nwt.GridGenericAsset_AssetData__)
		if !ok {
			slog.Error("unexpected chunks type", "file", file.Path())
			continue
		}
		for _, chunk := range chunks.Chunks.Element {
			sliceFile, err := c.LookupFileByAssetId(chunk.AssetId)
			if err != nil || sliceFile == nil {
				slog.Warn("chunk slice not found", "file", file.Path(), "asset", chunk.AssetId.Guid, "error", err)
				continue
			}
			root := mat4.FromAzTransformData([]nwt.AzFloat32{
				0, 0, 0, 1,
				1, 1, 1,
				chunk.WorldPosition[0], chunk.WorldPosition[1], chunk.WorldPosition[2],
			})
			c.collectSliceInto(&group, sliceFile, root)
		}
		if !merge && len(group.Geometries) > 0 {
			group.TargetFile = c.outputPath(utils.ReplaceExt(file.Path(), ""))
			c.models.Store(group.TargetFile, group)
		}
	}

	if merge && len(group.Geometries) > 0 {
		group.TargetFile = c.outputPath(utils.ReplaceExt(flgOutFile, ""))
		c.models.Store(group.TargetFile, group)
	}
}
