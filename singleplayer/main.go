package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"os"
	"sort"

	event "../game"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

var (
	playerSprite, background *ebiten.Image
	world                    event.World
)

func spriteInit(spriteDir string) *ebiten.Image {
	file, err := os.Open(spriteDir)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	imgbytes := make([]byte, size)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(imgbytes)

	img, _, err := image.Decode(bytes.NewReader(imgbytes))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

func init() {
	playerSprite = spriteInit("../sprites/player.png")
	background = spriteInit("../sprites/background.png")

	world = event.World{Units: event.Units{}}
}

//Game is main game struct
type Game struct {
	pressed []ebiten.Key
	ID      string
}

func moveUnit(id string, dir int, frameUpd bool) {
	event := event.Event{
		Type: event.EventTypeMove,
		Data: event.EventMove{UnitID: id, Direction: dir, FrameUpdate: frameUpd},
	}
	world.EventHandler(event)

}

//Update is func for update game logic
func (g *Game) Update() error {
	g.pressed = nil
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if ebiten.IsKeyPressed(k) {
			g.pressed = append(g.pressed, k)
		}
	}

	f := func(keys []ebiten.Key, i, b int) []ebiten.Key {
		var newKeys []ebiten.Key
		for index, key := range keys {
			if index != i && index != b {
				newKeys = append(newKeys, key)
			}
		}
		return newKeys
	}

	for index := 0; index < len(g.pressed)-1; index++ {
		if (g.pressed[index] == ebiten.KeyD && g.pressed[index+1] == ebiten.KeyA) ||
			(g.pressed[index+1] == ebiten.KeyD && g.pressed[index] == ebiten.KeyA) {
			g.pressed = f(g.pressed, index, index+1)
		}
	}

	if len(g.pressed) == 0 {
		event := event.Event{
			Type: event.EventTypeIdle,
			Data: event.EventIdle{UnitID: g.ID},
		}
		world.EventHandler(event)
	}

	var frameUpd bool = true
	for slicePos, key := range g.pressed {
		if slicePos > 0 {
			frameUpd = false
		}
		switch key {
		case ebiten.KeyA:
			moveUnit(g.ID, event.DirectionLeft, frameUpd)
		case ebiten.KeyD:
			moveUnit(g.ID, event.DirectionRight, frameUpd)
		case ebiten.KeyW:
			moveUnit(g.ID, event.DirectionUp, frameUpd)
		case ebiten.KeyS:
			moveUnit(g.ID, event.DirectionDown, frameUpd)
		}
	}
	return nil
}

//Draw is func for draw game image
func (g *Game) Draw(screen *ebiten.Image) {
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	ebitenutil.DebugPrint(screen, msg)

	backop := &ebiten.DrawImageOptions{}
	backop.GeoM.Scale(.65, .7)
	backop.GeoM.Translate(0, 0)
	screen.DrawImage(background, backop)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1.5, 1.5)

	unitsList := []*event.Unit{}
	for _, unit := range world.Units {
		unitsList = append(unitsList, unit)
	}
	sort.Slice(unitsList, func(i, j int) bool { return unitsList[i].PosY < unitsList[j].PosY })

	for _, u := range unitsList {

		if u.HorizDir == event.DirectionLeft {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(16, 0)
		}
		op.GeoM.Translate(u.PosX, u.PosY)

		i := math.Floor(u.Frame * event.AnimationSpeed)

		var sx, sy int
		sx = event.FrameOX + int(i)*event.FrameWidth
		if u.State == event.EventTypeMove {
			sy = event.FrameOY + event.FrameHeight
		} else if u.State == event.EventTypeIdle {
			sy = event.FrameOY
		}

		screen.DrawImage(playerSprite.SubImage(image.Rect(sx, sy, sx+event.FrameWidth,
			sy+event.FrameHeight)).(*ebiten.Image), op)
	}
}

//Layout is func for set resolution
func (g Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	playerInit, id := event.PlayerInit()
	world.EventHandler(playerInit)

	ebiten.SetWindowSize(event.ScreenWidth, event.ScreenHeight)
	ebiten.SetWindowTitle("Golang 2DGame")
	if err := ebiten.RunGame(&Game{ID: id}); err != nil {
		log.Fatal(err)
	}
}
