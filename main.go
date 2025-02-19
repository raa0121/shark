package main

import (
	"bytes"
	"embed"
	"flag"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const SPRITE_X = 100
const SPRITE_Y = 123

//go:embed sprites/idle/*
var IdleSprites embed.FS

//go:embed sprites/right-click/*
var RightClickSprites embed.FS

//go:embed sprites/drag/*
var DragSprites embed.FS

type Anim struct {
	Frames []*ebiten.Image
}

type Position struct{ x, y int }

type Game struct {
	CurrentAnim      *Anim
	CurrentFrame     int
	Ticks            int
	IsDragging       bool
	PreviousMousePos Vector
	WinStartPos      Vector
	MouseStartPos    Vector
}

type Vector struct{ x, y int }

func CreateVector(x, y int) Vector {
	return Vector{x, y}
}

func (this Vector) Add(that Vector) Vector {
	return Vector{this.x + that.x, this.y + that.y}
}
func (this Vector) Subtract(that Vector) Vector {
	return Vector{this.x - that.x, this.y - that.y}
}

func GlobalCursorPosition() Vector {
	cx, cy := ebiten.CursorPosition()
	wx, wy := ebiten.WindowPosition()
	return Vector{cx + wx, cy + wy}
}

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if g.CurrentAnim == Idle {
			g.CurrentAnim = RightClick
			g.Ticks = 0
			g.CurrentFrame = 0
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.IsDragging = true
		g.CurrentAnim = Drag
		g.Ticks = 0
		g.CurrentFrame = 0
		g.PreviousMousePos = GlobalCursorPosition()
		g.WinStartPos = CreateVector(ebiten.WindowPosition())
		g.MouseStartPos = GlobalCursorPosition()
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.IsDragging = false
		g.CurrentAnim = Idle
		g.Ticks = 0
		g.CurrentFrame = 0
	}

	mousePos := GlobalCursorPosition()
	if g.IsDragging && mousePos != g.PreviousMousePos {
		newWinPos := g.WinStartPos.Add(mousePos.Subtract(g.MouseStartPos))
		ebiten.SetWindowPosition(newWinPos.x, newWinPos.y)
	}
	g.PreviousMousePos = mousePos

	g.Ticks++
	if g.Ticks < 10 {
		return nil
	}

	g.Ticks = 0
	g.CurrentFrame++
	if g.CurrentFrame >= len(g.CurrentAnim.Frames) {
		g.CurrentFrame = 0
		if g.CurrentAnim == RightClick {
			g.CurrentAnim = Idle
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.CurrentAnim.Frames[g.CurrentFrame], nil)
	/*
		debugStr := ""
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			debugStr += "Dragging\n"
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			debugStr += "Right click\n"
		}
		ebitenutil.DebugPrint(screen, debugStr)
	*/
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (w, h int) {
	return SPRITE_X, SPRITE_Y
}

func NewAnim(sprites embed.FS, subdir string) *Anim {
	files, err := sprites.ReadDir("sprites/" + subdir)
	PanicIfErr(err)
	var frames []*ebiten.Image
	for _, direntry := range files {
		fname := direntry.Name()
		frame, err := sprites.ReadFile("sprites/" + subdir + "/" + fname)
		PanicIfErr(err)
		img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(frame))
		PanicIfErr(err)
		frames = append(frames, img)
	}
	return &Anim{frames}
}

var Idle, RightClick, Drag *Anim

func init() {
	Idle = NewAnim(IdleSprites, "idle")
	Drag = NewAnim(DragSprites, "drag")
	RightClick = NewAnim(RightClickSprites, "right-click")
}

func main() {
	var sizeFlag int
	flag.IntVar(
		&sizeFlag, "size", 2, "Size multiplier: make Gura as big as you want.",
	)
	flag.Parse()

	var game Game
	game.CurrentAnim = Idle

	ebiten.SetWindowSize(SPRITE_X*sizeFlag, SPRITE_Y*sizeFlag)
	ebiten.SetWindowTitle("Shark!")
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowPosition(9999, 9999)
	ebiten.SetWindowFloating(true)

	err := ebiten.RunGame(&game)
	PanicIfErr(err)
}

func PanicIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
