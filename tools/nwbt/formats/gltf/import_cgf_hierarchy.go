package gltf

import (
	"nw-buddy/tools/formats/cgf"
	"nw-buddy/tools/utils/maps"
	"nw-buddy/tools/utils/math/mat4"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/qmuntal/gltf"
)

type HandleObjectFunc func(node *gltf.Node, chunk cgf.Chunker, name string, lod int)

var lodRegex = regexp.MustCompile(`\$lod(\d+)`)

func (doc *Document) ImportCgfHierarchy(cgfile *cgf.File, handleObject HandleObjectFunc) []*gltf.Node {
	base := doc.Options.Base

	rootNodes := make([]*gltf.Node, 0)
	nodeMap := newIdToNodeMap(doc)
	for _, chunk := range cgf.SelectChunks[cgf.ChunkNode](cgfile) {

		isLod := strings.Contains(chunk.Name, "$lod")
		isHelper := strings.Contains(chunk.Name, "$physics") || strings.HasPrefix(chunk.Name, "$")
		lod := 0

		shouldSkip := false
		if isLod && doc.Options.SkipLod {
			shouldSkip = true
		}
		if !isLod && isHelper && doc.Options.SkipHelper {
			shouldSkip = true
		}

		if isLod {
			// alguns nomes trazem "$lod" sem número
			if m := lodRegex.FindStringSubmatch(chunk.Name); len(m) > 1 {
				lod, _ = strconv.Atoi(m[1])
			}
		}

		node, nodeIndex := nodeMap.lookup(chunk.ChunkHeader.Id)
		if !mat4.IsIdentity(chunk.Transform) {
			node.Matrix = mat4.ToFloat64(base.Mat4(mat4.Transpose(chunk.Transform)))
		}

		if chunk.ParentId == -1 {
			rootNodes = append(rootNodes, node)
		} else {
			parent, _ := nodeMap.lookup(chunk.ParentId)
			parent.Children = append(parent.Children, nodeIndex)
		}
		if shouldSkip {
			continue
		}

		if object, ok := cgf.FindChunk[cgf.Chunker](cgfile, chunk.ObjectId); ok {
			handleObject(node, object, chunk.Name, lod)
		}
	}
	return rootNodes
}

type idToNodeMap struct {
	converter *Document
	idToNode  *maps.Map[int32, *gltf.Node]
}

func newIdToNodeMap(converter *Document) *idToNodeMap {
	return &idToNodeMap{
		converter: converter,
		idToNode:  maps.NewMap[int32, *gltf.Node](),
	}
}

func (n *idToNodeMap) lookup(id int32) (*gltf.Node, int) {
	node, loaded := n.idToNode.LoadOrStore(id, &gltf.Node{})
	if !loaded {
		n.converter.AppendNode(node)
	}
	return node, slices.Index(n.converter.Document.Nodes, node)
}
