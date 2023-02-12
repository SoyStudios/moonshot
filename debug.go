package main

import (
	"image"
	"image/draw"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type PhyicsDebug struct {
	enabled bool

	g *Game

	w, h int
	dc   *gg.Context
	img  *ebiten.Image
}

func (d *PhyicsDebug) initialize() {
	if !d.enabled {
		return
	}
	var rebuild bool
	if d.w != int(d.g.camera.ViewPort.X/d.g.camera.zoomFactor()) {
		rebuild = true
		d.w = int(d.g.camera.ViewPort.X / d.g.camera.zoomFactor())
	}
	if d.h != int(d.g.camera.ViewPort.Y/d.g.camera.zoomFactor()) {
		rebuild = true
		d.h = int(d.g.camera.ViewPort.Y / d.g.camera.zoomFactor())
	}
	if rebuild {
		infoLog.Log("msg", "resize",
			"w", d.w,
			"h", d.h,
		)
		d.dc = gg.NewContext(d.w, d.h)
		d.img = ebiten.NewImage(d.w, d.h)
	}
	d.dc.SetRGBA255(0xff, 0x00, 0xff, 0x30)
	d.dc.Clear()
}

func (d *PhyicsDebug) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data interface{}) {
	x, y := pos.X-d.g.camera.Position.X, pos.Y-d.g.camera.Position.Y
	d.dc.DrawCircle(x, y, radius)
	d.dc.SetRGBA255(int(outline.R), int(outline.G), int(outline.B), int(outline.A))
	d.dc.StrokePreserve()
	d.dc.SetRGBA255(int(fill.R), int(fill.G), int(fill.B), int(fill.A))
	d.dc.Fill()
}

func (d *PhyicsDebug) DrawSegment(a, b cp.Vector, fill cp.FColor, data interface{}) {
	aX, aY := a.X-d.g.camera.Position.X, a.Y-d.g.camera.Position.Y
	bX, bY := b.X-d.g.camera.Position.X, b.Y-d.g.camera.Position.Y
	d.dc.DrawLine(aX, aY, bX, bY)
	d.dc.SetLineWidth(2)
	d.dc.SetRGBA255(int(fill.R), int(fill.G), int(fill.B), int(fill.A))
	d.dc.StrokePreserve()
	d.dc.Fill()
}

func (d *PhyicsDebug) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
}

func (d *PhyicsDebug) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	posX, posY := d.g.camera.Position.X, d.g.camera.Position.Y
	for i, v := range verts {
		if i == 0 {
			d.dc.MoveTo(v.X-posX, v.Y-posY)
			continue
		}
		d.dc.LineTo(v.X-posX, v.Y-posY)
	}
	d.dc.SetLineWidth(2)
	d.dc.SetRGBA255(int(outline.R), int(outline.G), int(outline.B), int(outline.A))
	d.dc.StrokePreserve()
	d.dc.SetRGBA255(int(fill.R), int(fill.G), int(fill.B), int(fill.A))
	d.dc.Fill()
}

func (d *PhyicsDebug) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {
	x, y := pos.X-d.g.camera.Position.X, pos.Y-d.g.camera.Position.Y
	d.dc.DrawCircle(x, y, size)
	d.dc.SetRGBA255(int(fill.R), int(fill.G), int(fill.B), int(fill.A))
	d.dc.Fill()
}

func (d *PhyicsDebug) Draw(target *ebiten.Image) {
	img := d.dc.Image()
	size := img.Bounds().Size()
	w, h := size.X, size.Y
	bs := make([]byte, 4*w*h)
	dstImg := &image.RGBA{
		Pix:    bs,
		Stride: 4 * w,
		Rect:   image.Rect(0, 0, w, h),
	}
	draw.Draw(dstImg, image.Rect(0, 0, w, h), img, img.Bounds().Min, draw.Src)
	d.img.Clear()
	d.img.WritePixels(bs)

	g := ebiten.GeoM{}
	g.Scale(d.g.camera.zoomFactor(), d.g.camera.zoomFactor())
	g.Translate(
		-d.g.camera.ViewPort.X*d.g.camera.zoomFactor(),
		-d.g.camera.ViewPort.Y*d.g.camera.zoomFactor())

	target.DrawImage(d.img, &ebiten.DrawImageOptions{
		GeoM: g,
	})
}

func (d *PhyicsDebug) Flags() uint {
	return cp.DRAW_SHAPES | cp.DRAW_COLLISION_POINTS
}

func (d *PhyicsDebug) OutlineColor() cp.FColor {
	return cp.FColor{
		R: 0x00,
		G: 0xA0,
		B: 0xA0,
		A: 0xFF,
	}
}

func (d *PhyicsDebug) ShapeColor(shp *cp.Shape, data interface{}) cp.FColor {
	return cp.FColor{
		R: 0x00,
		G: 0xA0,
		B: 0xA0,
		A: 0xF0,
	}
}

func (d *PhyicsDebug) ConstraintColor() cp.FColor {
	return cp.FColor{
		R: 0xeb,
		G: 0x71,
		B: 0x34,
		A: 0xFF,
	}
}

func (d *PhyicsDebug) CollisionPointColor() cp.FColor {
	return cp.FColor{
		R: 0xff,
		G: 0x00,
		B: 0x00,
		A: 0xFF,
	}
}

func (d *PhyicsDebug) Data() interface{} {
	return nil
}
