package models

import (
	"fmt"
	"nw-buddy/tools/game"
	"nw-buddy/tools/rtti/nwt"
	"nw-buddy/tools/utils"
	"nw-buddy/tools/utils/logging"
	"log/slog"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// models slicestats: diagnóstico de uma dynamicslice — tipos de componente (com slices filhos),
// slices visitados e spawners cujo asset não resolve.
var cmdSliceStats = &cobra.Command{
	Use:   "slicestats",
	Short: "prints component statistics of a .dynamicslice tree",
	Run:   runSliceStats,
}

func init() {
	cmdSliceStats.Flags().AddFlagSet(Cmd.Flags())
}

func runSliceStats(ccmd *cobra.Command, args []string) {
	slog.SetDefault(logging.DefaultFileHandler())
	c := utils.Must(initCollector())
	files, err := c.Archive.Glob(strings.ToLower(args[0]))
	if err != nil || len(files) == 0 {
		fmt.Println("slice não encontrada")
		return
	}
	types := map[string]int{}
	slices := map[string]int{}
	unresolved := map[string]int{}
	check := func(node *game.SliceNode, asset nwt.AzAsset, alias nwt.AzAsset) {
		if alias.Guid != "" && alias.Guid != "00000000-0000-0000-0000-000000000000" {
			if f, _ := c.LookupFileByAsset(alias); f != nil {
				slices["alias "+f.Path()]++
				return
			}
		}
		if f, _ := c.LookupFileByAsset(asset); f == nil && asset.Guid != "" {
			unresolved[asset.Hint]++
		}
	}
	c.WalkSlice(files[0], func(node *game.SliceNode) {
		if node.File != nil {
			slices[node.File.Path()]++
		}
		for _, component := range node.Components {
			types[fmt.Sprintf("%T", component)]++
			switch v := component.(type) {
			case nwt.PrefabSpawnerComponent:
				check(node, v.M_sliceAsset, v.M_aliasAsset)
			case nwt.PointSpawnerComponent:
				check(node, v.BaseClass1.M_sliceAsset, v.BaseClass1.M_aliasAsset)
			}
		}
	})
	dump := func(title string, m map[string]int) {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
		fmt.Printf("== %s (%d)\n", title, len(keys))
		for _, k := range keys {
			fmt.Printf("%6d %s\n", m[k], k)
		}
	}
	dump("componentes", types)
	dump("slices", slices)
	dump("spawners sem asset", unresolved)
}
