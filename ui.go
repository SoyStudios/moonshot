package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type (
	UI struct {
		game *Game

		layer *ebiten.Image

		info InfoDrawer
	}

	UIUpdater interface {
		Update()
	}

	InfoDrawer interface {
		DrawInfo(*UI, *ebiten.Image)
	}
)

func NewUI(g *Game) *UI {
	return &UI{
		game:  g,
		layer: ebiten.NewImage(g.w, g.h),
	}
}

func (u *UI) Draw(screen *ebiten.Image) {
	u.layer.Fill(color.RGBA{255, 255, 255, 0})

	op := &ebiten.DrawImageOptions{}

	if u.info != nil {
		op.GeoM.Translate(float64(u.game.w)/3*2, 0)
		info := u.InfoScreen()
		u.info.DrawInfo(u, info)
		u.layer.DrawImage(info, op)
	}

	op.GeoM.Reset()
	screen.DrawImage(u.layer, op)
}

func (u *UI) Update() {
}

func (u *UI) uiImg(name string) *ebiten.Image {
	switch name {
	case "blueScreenPanelTopLeft":
		return u.game.assets.ui.SubImage(image.Rect(439, 183, 439+72, 183+72)).(*ebiten.Image)
	case "blueScreenPanelTop":
		return u.game.assets.ui.SubImage(image.Rect(450, 183, 450+72, 183+72)).(*ebiten.Image)
	case "blueScreenPanelTopRight":
		return u.game.assets.ui.SubImage(image.Rect(654, 183, 654+72, 183+72)).(*ebiten.Image)
	case "blueScreenPanelLeft":
		return u.game.assets.ui.SubImage(image.Rect(439, 241, 439+72, 241+72)).(*ebiten.Image)
	case "blueScreenPanel":
		return u.game.assets.ui.SubImage(image.Rect(450, 241, 450+72, 241+72)).(*ebiten.Image)
	case "blueScreenPanelRight":
		return u.game.assets.ui.SubImage(image.Rect(654, 241, 654+72, 241+72)).(*ebiten.Image)
	case "blueScreenPanelBottomLeft":
		return u.game.assets.ui.SubImage(image.Rect(439, 388, 439+24, 388+72)).(*ebiten.Image)
	case "blueScreenPanelBottom":
		return u.game.assets.ui.SubImage(image.Rect(459, 388, 459+24, 388+72)).(*ebiten.Image)
	case "blueScreenPanelBottomRight":
		return u.game.assets.ui.SubImage(image.Rect(702, 388, 702+24, 388+72)).(*ebiten.Image)
	case "blank72":
		img := ebiten.NewImage(72, 72)
		img.Fill(color.CMYK{0, 255, 0, 0})
		return img
	default:
		img := ebiten.NewImage(24, 24)
		img.Fill(color.CMYK{0, 255, 0, 0})
		return img
	}
}

func (u *UI) InfoScreen() *ebiten.Image {
	infoScreen := ebiten.NewImage(u.game.w/3, u.game.h)
	w, h := infoScreen.Size()
	op := &ebiten.DrawImageOptions{}
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelTopLeft"),
		op,
	)

	op.GeoM.Scale((float64(w)-144)/72, 1)
	op.GeoM.Translate(72, 0)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelTop"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(w)-72, 0)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelTopRight"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Scale(1, (float64(h)-144)/72)
	op.GeoM.Translate(0, 72)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelLeft"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Scale((float64(w)-144)/72, (float64(h)-144)/72)
	op.GeoM.Translate(72, 72)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanel"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Scale(1, (float64(h)-144)/72)
	op.GeoM.Translate(float64(w)-72, 72)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelRight"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Translate(0, float64(h)-72)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelBottomLeft"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Scale((float64(w)-24-24)/24, 1)
	op.GeoM.Translate(24, float64(h)-72)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelBottom"),
		op,
	)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(w)-24, float64(h)-72)
	infoScreen.DrawImage(
		u.uiImg("blueScreenPanelBottomRight"),
		op,
	)

	text.Draw(infoScreen,
		"Information",
		u.game.assets.font,
		24, 24,
		color.White,
	)
	return infoScreen
}
