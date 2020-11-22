package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type UI struct {
	game *Game

	layer *ebiten.Image

	bot *Bot
}

func NewUI(g *Game) *UI {
	return &UI{
		game:  g,
		layer: ebiten.NewImage(g.w/3, g.h),
	}
}

func (u *UI) uiImg(name string) *ebiten.Image {
	switch name {
	case "blueScreenPanelTopLeft":
		return u.game.assets.ui.SubImage(image.Rect(439, 183, 439+72, 173+72)).(*ebiten.Image)
	case "blueScreenPanelTop":
		return u.game.assets.ui.SubImage(image.Rect(450, 183, 450+72, 183+72)).(*ebiten.Image)
	case "blueScreenPanelTopRight":
		return u.game.assets.ui.SubImage(image.Rect(654, 183, 654+72, 173+72)).(*ebiten.Image)
	default:
		img := ebiten.NewImage(24, 24)
		img.Fill(color.CMYK{0, 255, 0, 0})
		return img
	}
}

func (u *UI) Draw(screen *ebiten.Image) {
	u.layer.Fill(color.RGBA{255, 255, 255, 0})

	op := &ebiten.DrawImageOptions{}
	u.layer.DrawImage(
		u.uiImg("blueScreenPanelTopLeft"),
		op)
	// total length of window (w / 3) - 2 corners / tile size
	op.GeoM.Scale((float64(u.game.w)/3-144)/72, 1)
	op.GeoM.Translate(72, 0)
	u.layer.DrawImage(
		u.uiImg("blueScreenPanelTop"),
		op,
	)
	op.GeoM.Reset()
	op.GeoM.Translate(float64(u.game.w)/3-72, 0)
	u.layer.DrawImage(
		u.uiImg("blueScreenPanelTopRight"),
		op,
	)
	text.Draw(u.layer,
		"Information",
		u.game.assets.font,
		24, 24,
		color.White,
	)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(u.game.w)/3*2, 0)
	screen.DrawImage(u.layer, op)
}
