// Copyright (c) 2021, The Grid Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grid

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/oswin"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// Sprites are the type of sprite
type Sprites int

const (
	// SpUnk is an unknown sprite type
	SpUnk Sprites = iota

	// SpReshapeBBox is a reshape bbox -- the overall active selection BBox
	// for active manipulation
	SpReshapeBBox

	// SpSelBBox is a selection bounding box -- display only
	SpSelBBox

	// SpNodePoint is a main coordinate point for path node
	SpNodePoint

	// SpNodeCtrl is a control coordinate point for path node
	SpNodeCtrl

	// SpRubberBand is the draggable sel box
	// subtyp = UpC, LfM, RtM, DnC for sides
	SpRubberBand

	// SpAlignMatch is an alignment match (n of these),
	// subtyp is actually BBoxPoints so we just hack cast that
	SpAlignMatch

	// below are subtypes:

	// Sprite bounding boxes are set as a "bbox" property on sprites
	SpBBoxUpL
	SpBBoxUpC
	SpBBoxUpR
	SpBBoxDnL
	SpBBoxDnC
	SpBBoxDnR
	SpBBoxLfM
	SpBBoxRtM

	// todo: add nodectrl subtypes

	SpritesN
)

//go:generate stringer -type=Sprites

var KiT_Sprites = kit.Enums.AddEnum(SpritesN, kit.NotBitFlag, nil)

func (ev Sprites) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Sprites) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// SpriteNames are name strings to use for naming sprites
var SpriteNames = map[Sprites]string{
	SpBBoxUpL: "up-l",
	SpBBoxUpC: "up-c",
	SpBBoxUpR: "up-r",
	SpBBoxDnL: "dn-l",
	SpBBoxDnC: "dn-c",
	SpBBoxDnR: "dn-r",
	SpBBoxLfM: "lf-m",
	SpBBoxRtM: "rt-m",

	SpReshapeBBox: "reshape-bbox",

	SpSelBBox: "sel-bbox",

	SpNodePoint: "node-point",
	SpNodeCtrl:  "node-ctrl",

	SpRubberBand: "rubber-band",

	SpAlignMatch: "align-match",
}

// SpriteName returns the unique name of the sprite based
// on main type, subtype (e.g., bbox) if relevant, and index if relevant
func SpriteName(typ, subtyp Sprites, idx int) string {
	nm := SpriteNames[typ]
	switch typ {
	case SpReshapeBBox:
		nm += "-" + SpriteNames[subtyp]
	case SpSelBBox:
		nm += fmt.Sprintf("-%d-%s", idx, SpriteNames[subtyp])
	case SpNodePoint:
		nm += fmt.Sprintf("-%d", idx)
	case SpNodeCtrl: // todo: subtype
		nm += fmt.Sprintf("-%d", idx)
	case SpRubberBand:
		nm += "-" + SpriteNames[subtyp]
	case SpAlignMatch:
		nm += fmt.Sprintf("-%d", idx)
	}
	return nm
}

// SetSpriteProps sets sprite properties
func SetSpriteProps(sp *gi.Sprite, typ, subtyp Sprites, idx int) {
	sp.Name = SpriteName(typ, subtyp, idx)
	sp.Props.Set("grid-type", typ)
	sp.Props.Set("grid-sub", subtyp)
	sp.Props.Set("grid-idx", idx)
}

// SpriteProps reads the sprite properties -- returns SpUnk if
// not one of our sprites.
func SpriteProps(sp *gi.Sprite) (typ, subtyp Sprites, idx int) {
	typi, has := sp.Props["grid-type"]
	if !has {
		typ = SpUnk
		return
	}
	typ = typi.(Sprites)
	subtyp = sp.Props["grid-sub"].(Sprites)
	idx = sp.Props["grid-idx"].(int)
	return
}

// Sprite returns given sprite -- renders to window if not yet made.
// trgsz is the target size (e.g., for rubber band boxes)
func Sprite(win *gi.Window, typ, subtyp Sprites, idx int, trgsz image.Point) *gi.Sprite {
	spnm := SpriteName(typ, subtyp, idx)
	sp, ok := win.SpriteByName(spnm)
	if !ok {
		sp = gi.NewSprite(spnm, image.ZP, image.ZP)
		SetSpriteProps(sp, typ, subtyp, idx)
		win.AddSprite(sp)
	}
	switch typ {
	case SpReshapeBBox:
		DrawSpriteReshape(sp, subtyp)
	case SpSelBBox:
		DrawSpriteSel(sp, subtyp)
	case SpNodePoint:
		DrawSpriteNodePoint(sp, subtyp)
	case SpNodeCtrl:
		DrawSpriteNodeCtrl(sp, subtyp)
	case SpRubberBand:
		switch subtyp {
		case SpBBoxUpC, SpBBoxDnC:
			DrawRubberBandHoriz(sp, trgsz)
		case SpBBoxLfM, SpBBoxRtM:
			DrawRubberBandVert(sp, trgsz)
		}
	case SpAlignMatch:
		switch {
		case trgsz.X > trgsz.Y:
			DrawAlignMatchHoriz(sp, trgsz)
		default:
			DrawAlignMatchVert(sp, trgsz)
		}
	}
	win.ActivateSprite(sp.Name)
	return sp
}

// SpriteConnectEvent activates and sets mouse event functions to given function
func SpriteConnectEvent(win *gi.Window, typ, subtyp Sprites, idx int, trgsz image.Point, recv ki.Ki, fun ki.RecvFunc) *gi.Sprite {
	sp := Sprite(win, typ, subtyp, idx, trgsz)
	if recv != nil {
		sp.ConnectEvent(recv, oswin.MouseEvent, fun)
		sp.ConnectEvent(recv, oswin.MouseDragEvent, fun)
	}
	return sp
}

// SetSpritePos sets sprite position, taking into account relative offsets
func SetSpritePos(sp *gi.Sprite, pos image.Point) {
	typ, subtyp, _ := SpriteProps(sp)
	switch {
	case typ == SpRubberBand:
		_, sz := LineSpriteSize()
		switch subtyp {
		case SpBBoxUpC:
			pos.Y -= sz
		case SpBBoxLfM:
			pos.X -= sz
		}
	case typ == SpAlignMatch:
		_, sz := LineSpriteSize()
		bbtp := BBoxPoints(subtyp) // just hack it
		switch bbtp {
		case BBLeft:
			pos.X -= sz
		case BBCenter:
			pos.X -= sz / 2
		case BBTop:
			pos.Y -= sz
		case BBMiddle:
			pos.Y -= sz / 2
		}
	case typ == SpNodePoint || typ == SpNodeCtrl:
		_, sz := HandleSpriteSize(1)
		pos.X -= sz.X / 2
		pos.Y -= sz.Y / 2
	case subtyp >= SpBBoxUpL && subtyp <= SpBBoxRtM: // Reshape, Sel BBox
		sc := float32(1)
		if typ == SpSelBBox {
			sc = .8
		}
		_, sz := HandleSpriteSize(sc)
		if subtyp == SpBBoxDnL || subtyp == SpBBoxUpL || subtyp == SpBBoxLfM {
			pos.X -= sz.X
		}
		if subtyp == SpBBoxUpL || subtyp == SpBBoxUpC || subtyp == SpBBoxUpR {
			pos.Y -= sz.Y
		}
	}
	sp.Geom.Pos = pos
}

// InactivateSprites inactivates sprites of given type
func InactivateSprites(win *gi.Window, typ Sprites) {
	for _, spkv := range win.Sprites.Names.Order {
		sp := spkv.Val
		st, _, _ := SpriteProps(sp)
		if st == typ {
			win.InactivateSprite(sp.Name)
		}
	}
}

///////////////////////////////////////////////////////////////////
//  Sprite rendering

var (
	HandleSpriteScale = float32(18)
	HandleSizeMin     = 4
	HandleBorderMin   = 2
)

// HandleSpriteSize returns the border size and overall size
// of handle-type sprites, with given scaling factor
func HandleSpriteSize(scale float32) (int, image.Point) {
	sz := int(mat32.Ceil(scale * gi.Prefs.LogicalDPIScale * HandleSpriteScale))
	sz = ints.MaxInt(sz, HandleSizeMin)
	bsz := ints.MaxInt(sz/6, HandleBorderMin)
	bbsz := image.Point{sz, sz}
	return bsz, bbsz
}

// DrawSpriteReshape renders a Reshape sprite handle
func DrawSpriteReshape(sp *gi.Sprite, bbtyp Sprites) {
	bsz, bbsz := HandleSpriteSize(1)
	if !sp.SetSize(bbsz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.X += bsz
	bbd.Min.Y += bsz
	bbd.Max.X -= bsz
	bbd.Max.Y -= bsz
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(sp.Pixels, bbd, &image.Uniform{color.Black}, image.ZP, draw.Src)
}

// DrawSpriteSel renders a Select sprite handle -- smaller
func DrawSpriteSel(sp *gi.Sprite, bbtyp Sprites) {
	bsz, bbsz := HandleSpriteSize(.8)
	if !sp.SetSize(bbsz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.X += bsz
	bbd.Min.Y += bsz
	bbd.Max.X -= bsz
	bbd.Max.Y -= bsz
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(sp.Pixels, bbd, &image.Uniform{color.Black}, image.ZP, draw.Src)
}

// DrawSpriteNodePoint renders a NodePoint sprite handle
func DrawSpriteNodePoint(sp *gi.Sprite, bbtyp Sprites) {
	bsz, bbsz := HandleSpriteSize(1)
	if !sp.SetSize(bbsz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.X += bsz
	bbd.Min.Y += bsz
	bbd.Max.X -= bsz
	bbd.Max.Y -= bsz
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(sp.Pixels, bbd, &image.Uniform{color.Black}, image.ZP, draw.Src)
}

// DrawSpriteNodeCtrl renders a NodePoint sprite handle
func DrawSpriteNodeCtrl(sp *gi.Sprite, subtyp Sprites) {
	bsz, bbsz := HandleSpriteSize(1)
	if !sp.SetSize(bbsz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.X += bsz
	bbd.Min.Y += bsz
	bbd.Max.X -= bsz
	bbd.Max.Y -= bsz
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(sp.Pixels, bbd, &image.Uniform{color.Black}, image.ZP, draw.Src)
}

var (
	LineSpriteScale = float32(8)
	LineSizeMin     = 3
	LineBorderMin   = 1
)

// LineSpriteSize returns the border size and overall size of line-type sprites
func LineSpriteSize() (int, int) {
	sz := int(mat32.Ceil(gi.Prefs.LogicalDPIScale * LineSpriteScale))
	sz = ints.MaxInt(sz, LineSizeMin)
	bsz := ints.MaxInt(sz/6, LineBorderMin)
	return bsz, sz
}

// DrawRubberBandHoriz renders a horizontal rubber band line
func DrawRubberBandHoriz(sp *gi.Sprite, trgsz image.Point) {
	bsz, sz := LineSpriteSize()
	ssz := image.Point{trgsz.X, sz}
	if !sp.SetSize(ssz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.Y += bsz
	bbd.Max.Y -= bsz
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	for x := 0; x < ssz.X; x += sz * 2 {
		bbd.Min.X = x
		bbd.Max.X = x + sz
		draw.Draw(sp.Pixels, bbd, &image.Uniform{color.Black}, image.ZP, draw.Src)
	}
}

// DrawRubberBandVert renders a vertical rubber band line
func DrawRubberBandVert(sp *gi.Sprite, trgsz image.Point) {
	bsz, sz := LineSpriteSize()
	ssz := image.Point{sz, trgsz.Y}
	if !sp.SetSize(ssz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.X += bsz
	bbd.Max.X -= bsz
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	for y := sz; y < ssz.Y; y += sz * 2 {
		bbd.Min.Y = y
		bbd.Max.Y = y + sz
		draw.Draw(sp.Pixels, bbd, &image.Uniform{color.Black}, image.ZP, draw.Src)
	}
}

// DrawAlignMatchHoriz renders a horizontal alignment line
func DrawAlignMatchHoriz(sp *gi.Sprite, trgsz image.Point) {
	bsz, sz := LineSpriteSize()
	ssz := image.Point{trgsz.X, sz}
	if !sp.SetSize(ssz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.Y += bsz
	bbd.Max.Y -= bsz
	clr := gist.Color{0, 200, 200, 255}
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(sp.Pixels, bbd, &image.Uniform{clr}, image.ZP, draw.Src)
}

// DrawAlignMatchVert renders a vertical alignment line
func DrawAlignMatchVert(sp *gi.Sprite, trgsz image.Point) {
	bsz, sz := LineSpriteSize()
	ssz := image.Point{sz, trgsz.Y}
	if !sp.SetSize(ssz) { // already set
		return
	}
	ibd := sp.Pixels.Bounds()
	bbd := ibd
	bbd.Min.X += bsz
	bbd.Max.X -= bsz
	clr := gist.Color{0, 200, 200, 255}
	draw.Draw(sp.Pixels, ibd, &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(sp.Pixels, bbd, &image.Uniform{clr}, image.ZP, draw.Src)
}
