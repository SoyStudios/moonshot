package main

import "github.com/hajimehoshi/ebiten/v2"

func defaultInputMap() map[string][]ebiten.Key {
	return map[string][]ebiten.Key{
		"up":    {ebiten.KeyW},
		"down":  {ebiten.KeyS},
		"left":  {ebiten.KeyA},
		"right": {ebiten.KeyD},

		"zoomOut": {ebiten.KeyQ},
		"zoomIn":  {ebiten.KeyE},

		"pause": {ebiten.KeySpace},
		"step":  {ebiten.Key1},

		"speedUp":   {ebiten.KeyEqual},
		"speedDown": {ebiten.KeyMinus},

		"physicsDebug": {ebiten.KeyBackquote},
	}
}
