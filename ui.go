package main

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/jakecoffman/cp"
)

type (
	UI struct {
		game *Game

		layer *ebiten.Image

		info InfoDrawer
		code CodeDrawer
	}

	InfoDrawer interface {
		DrawInfo(*UI, *ebiten.Image)
	}

	CodeDrawer interface {
		DrawCode(*UI, *ebiten.Image)
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
		op.GeoM.Reset()
		op.GeoM.Translate(float64(u.game.w)/3*2, 0)
		info := u.InfoScreen()
		u.info.DrawInfo(u, info)
		u.layer.DrawImage(info, op)
	}

	if u.code != nil {
		op.GeoM.Reset()
		code := u.CodeScreen()
		u.code.DrawCode(u, code)
		u.layer.DrawImage(code, op)
	}

	op.GeoM.Reset()
	screen.DrawImage(u.layer, op)
}

func (u *UI) Update() {
	if !inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		return
	}
	x, y := u.game.camera.ScreenToWorld(ebiten.CursorPosition())
	log.Printf("click %.2f,%.2f", x, y)
	info := u.game.space.PointQueryNearest(cp.Vector{X: x, Y: y}, 5,
		cp.ShapeFilter{},
	)
	log.Printf("query: %+v", info)
	if info == nil || info.Shape == nil {
		return
	}
	bot, ok := info.Shape.UserData.(*Bot)
	if !ok {
		return
	}
	u.info = bot
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

	case "yellowScreenPanelTopLeft":
		return u.game.assets.ui.SubImage(image.Rect(774, 186, 774+72, 186+72)).(*ebiten.Image)

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

func (u *UI) CodeScreen() *ebiten.Image {
	codeScreen := ebiten.NewImage(u.game.w/3, u.game.h)
	// w, h := codeScreen.Size()
	op := &ebiten.DrawImageOptions{}
	codeScreen.DrawImage(
		u.uiImg("yellowScreenPanelTopLeft"),
		op,
	)

	text.Draw(codeScreen,
		"Gene",
		u.game.assets.font,
		24, 24,
		color.White,
	)
	return codeScreen
}
