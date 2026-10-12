// go run ./dbgflags <gameDir> <caf>... — imprime AssetFlags e contagem de controllers.
package main

import (
	"fmt"
	"os"

	"nw-buddy/tools/formats/cgf"
	"nw-buddy/tools/nwfs"
)

func main() {
	archive, err := nwfs.NewPackedArchive(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, p := range os.Args[2:] {
		f, ok := archive.Lookup(p)
		if !ok {
			fmt.Println("não achei", p)
			continue
		}
		doc, err := cgf.Load(f)
		if err != nil {
			fmt.Println(p, err)
			continue
		}
		mp, _ := cgf.SelectChunk[cgf.ChunkMotionParams](doc)
		ctrls := cgf.SelectChunks[cgf.ChunkController](doc)
		fmt.Printf("flags=0x%08x ctrls=%3d speed=%.2f dist=%.2f turn=%.2f end=%v %s\n", mp.AssetFlags, len(ctrls), mp.MoveSpeed, mp.Distance, mp.AssetTurn, mp.EndLocation, p)
	}
}
