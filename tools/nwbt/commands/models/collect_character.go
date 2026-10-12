package models

import (
	"log/slog"
	"nw-buddy/tools/formats/adb"
	"nw-buddy/tools/formats/cdf"
	"nw-buddy/tools/formats/gltf/importer"
	"nw-buddy/tools/game"
	"nw-buddy/tools/utils"
	"nw-buddy/tools/utils/logging"

	"github.com/spf13/cobra"
)

var cmdCollectCharacter = &cobra.Command{
	Use:   "character",
	Short: "",
	Long:  "",
	Run:   runCollectCharacter,
}

func init() {
	cmdCollectCharacter.Flags().AddFlagSet(Cmd.Flags())
}

func runCollectCharacter(ccmd *cobra.Command, args []string) {
	slog.SetDefault(logging.DefaultFileHandler())
	c := utils.Must(initCollector())
	c.CollectCharacter()
	c.Process()
	slog.SetDefault(logging.DefaultTerminalHandler())
}

func (c *Collector) CollectCharacter() {
	models := []string{
		game.CDF_MODEL_MALE,
		game.CDF_MODEL_FEMALE,
	}
	for _, model := range models {
		cdf, err := c.LoadCdf(model)
		if err != nil {
			slog.Warn("skip unresolved cdf file", "err", err)
			return
		}
		c.CollectCdf(cdf, "", c.outputPath(model))
	}
}

func (c *Collector) CollectCdf(cdf *cdf.Document, material string, targetFile string) {
	var animations []adb.AnimationFile
	if c.flags.Animations {
		animations, _ = c.CollectAnimations(cdf)
	}

	group := importer.AssetGroup{
		Extra: map[string]any{
			"animationList": animations,
		},
	}

	for _, attachment := range cdf.SkinAndClothAttachments() {
		if model, mtl := c.ResolveCgfAndMtl(attachment.Binding, attachment.Material, material); model != "" {
			group.Geometries = append(group.Geometries, importer.GeometryAsset{
				Entity: importer.Entity{
					Name: attachment.AName,
				},
				GeometryFile: model,
				MaterialFile: mtl,
			})
		}
	}
	group.Animations = append(group.Animations, flgExtraAnims...)
	if len(group.Geometries) > 0 {
		group.TargetFile = targetFile
		c.models.Store(group.TargetFile, group)
	}
}

func (c *Collector) CollectAnimations(cdf *cdf.Document) ([]adb.AnimationFile, error) {
	animations, err := cdf.LoadAnimationFiles(c.Archive)
	if err != nil {
		return nil, err
	}
	result := make([]adb.AnimationFile, 0)
	for _, animation := range animations {
		switch animation.Type {
		case adb.Caf:
			result = append(result, adb.AnimationFile{
				Name: animation.Name,
				File: c.assetPath(animation.File),
				Type: adb.Caf,
			})
			targetFile := c.outputPath(animation.File)
			if !c.shouldProcess(targetFile) {
				continue
			}
			c.models.Store(targetFile, importer.AssetGroup{
				TargetFile: targetFile,
				Animations: []importer.AnimationAsset{
					{
						File: animation.File,
						Name: animation.Name,
					},
				},
			})
		}
	}
	return result, nil
}
