// dbgheight converte region.heightmap (TIFF do New World) em .r16 cru (uint16 LE, origem no canto superior esquerdo).
//
//	go run ./dbgheight <region.heightmap> <saida.r16>
package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"nw-buddy/tools/formats/heightmap"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	hm, err := heightmap.ParseTIFF(data)
	if err != nil {
		panic(err)
	}
	lo, hi := uint16(0xffff), uint16(0)
	out := make([]byte, len(hm.Samples)*2)
	for i, v := range hm.Samples {
		lo, hi = min(lo, v), max(hi, v)
		binary.LittleEndian.PutUint16(out[i*2:], v)
	}
	if err := os.WriteFile(os.Args[2], out, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("%d %d %d %d\n", hm.Width, hm.Height, lo, hi)
}
