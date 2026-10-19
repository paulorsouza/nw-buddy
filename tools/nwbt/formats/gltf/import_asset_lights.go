package gltf

import (
	"fmt"
	"nw-buddy/tools/formats/gltf/importer"
	"nw-buddy/tools/utils/math/mat4"
)

func (doc *Document) ImportAssetLights(lights []importer.LightAsset, intensityScale float32) {
	if len(lights) == 0 {
		return
	}
	data := make([]map[string]any, len(lights))
	for i, light := range lights {

		data[i] = map[string]any{
			"type":      "point",
			"color":     light.Color,
			"intensity": light.Intensity * intensityScale,
			"range":     light.Range,
		}
		if light.Type == 2 {
			data[i]["type"] = "spot"
			data[i]["spot"] = map[string]any{
				"innerConeAngle": light.InnerConeAngle,
				"outerConeAngle": light.OuterConeAngle,
			}
		}

		node, _ := doc.NewNode()
		node.Name = light.Entity.Name
		if light.Extras != nil {
			// Nome único para o importador achar os parâmetros do jogo (extras) pelo ator.
			node.Name = fmt.Sprintf("nwlight_%d", i)
			node.Extras = light.Extras
		}
		node.Extensions = map[string]any{
			"KHR_lights_punctual": map[string]any{
				"light": i,
			},
		}
		node.Matrix = mat4.ToFloat64(light.Entity.Transform)
		doc.AddToScene(doc.DefaultScene(), node)
	}
	if doc.Extensions == nil {
		doc.Extensions = map[string]any{}
	}
	doc.Extensions["KHR_lights_punctual"] = map[string]any{
		"lights": data,
	}
}
