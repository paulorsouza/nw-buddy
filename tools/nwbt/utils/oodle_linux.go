package utils

import (
	"errors"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Linux: carrega libooz.so (reimplementação open source do Kraken) com a
// mesma assinatura do OodleLZ_Decompress. Caminho em NWBT_OODLE ou ao lado
// do executável.

type Oodle interface {
	Decompress(input []byte, inSize int, output []byte, outSize int) (int, error)
}

type oodle struct {
	decompress func(unsafe.Pointer, int, unsafe.Pointer, int, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr
}

var OodleLib = oodle{}

func (it oodle) Check() (string, bool) {
	candidates := []string{os.Getenv("NWBT_OODLE")}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "libooz.so"))
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func (it oodle) Info() string {
	return "libooz.so (Kraken decompressor), required for unpacking game archives\n" +
		"set NWBT_OODLE or place libooz.so next to the nwbt binary"
}

func (it *oodle) Load() error {
	if it.decompress != nil {
		return nil
	}
	p, ok := it.Check()
	if !ok {
		return errors.New("libooz.so not found")
	}
	handle, err := purego.Dlopen(p, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	purego.RegisterLibFunc(&it.decompress, handle, "OodleLZ_Decompress")
	return nil
}

func (it *oodle) Decompress(input []byte, inSize int, output []byte, outSize int) (int, error) {
	if err := it.Load(); err != nil {
		return 0, err
	}
	r1 := it.decompress(unsafe.Pointer(&input[0]), inSize, unsafe.Pointer(&output[0]), outSize, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3)
	return int(r1), nil
}

func OodleInstance() Oodle {
	return &OodleLib
}
