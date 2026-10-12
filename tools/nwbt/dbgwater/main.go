// dbgwater converte region.waterqt (quadtree de água do New World) em altura da água por metro: float32 LE,
// origem no canto superior esquerdo (mesma orientação do dbgheight). Sem água = valor negativo.
//
//	go run ./dbgwater <region.waterqt> <saida.f32>
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"nw-buddy/tools/formats/waterqt"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	doc, err := waterqt.Parse(data)
	if err != nil {
		panic(err)
	}
	hm, err := doc.ToHeightmap()
	if err != nil {
		panic(err)
	}
	lo, hi, wet := math.MaxFloat64, -math.MaxFloat64, 0
	out := make([]byte, len(hm.Data)*4)
	for i, v := range hm.Data {
		if v > 0 {
			lo, hi, wet = math.Min(lo, v), math.Max(hi, v), wet+1
		}
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(float32(v)))
	}
	if err := os.WriteFile(os.Args[2], out, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("lado=%d regiao=%d nos=%d com_agua=%d min=%.2f max=%.2f\n", hm.Size, doc.RegionSize, len(doc.QuadtreeNodes.Element), wet, lo, hi)
}
