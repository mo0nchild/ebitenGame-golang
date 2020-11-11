package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"os"
	"sort"

	event "../../game"
	"github.com/gorilla/websocket"
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
	playerSprite = spriteInit("../../sprites/player.png")
	background = spriteInit("../../sprites/background.png")

	world = event.World{Units: event.Units{}}

}

func getServerData(conn *websocket.Conn) {
	var (
		serverData event.Event
		byteData   []byte
	)
	defer conn.Close()
	for {
		err := conn.ReadJSON(&byteData)
		if err != nil {
			log.Println("read:", err)
			return
		}
		json.Unmarshal(byteData, &serverData)
		world.EventHandler(serverData)
	}
}

//Game is main game struct
type Game struct {
	pressed []ebiten.Key
	Conn    *websocket.Conn
	ID      string
}

func unitMove(id string, dir int, conn *websocket.Conn, frameUpd bool) {
	jsonData, _ := json.Marshal(
		event.Event{
			Type: event.EventTypeMove,
			Data: event.EventMove{UnitID: id, Direction: dir, FrameUpdate: frameUpd},
		})
	if err := conn.WriteJSON(jsonData); err != nil {
		log.Println(err)
	}
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
		jsonData, _ := json.Marshal(
			event.Event{
				Type: event.EventTypeIdle,
				Data: event.EventIdle{UnitID: g.ID},
			})
		if err := g.Conn.WriteJSON(jsonData); err != nil {
			log.Println(err)
		}
	}

	var frameUpd bool = true
	for slicePos, key := range g.pressed {
		if slicePos > 0 {
			frameUpd = false
		}
		switch key {
		case ebiten.KeyA:
			unitMove(g.ID, event.DirectionLeft, g.Conn, frameUpd)
		case ebiten.KeyD:
			unitMove(g.ID, event.DirectionRight, g.Conn, frameUpd)
		case ebiten.KeyW:
			unitMove(g.ID, event.DirectionUp, g.Conn, frameUpd)
		case ebiten.KeyS:
			unitMove(g.ID, event.DirectionDown, g.Conn, frameUpd)
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

	unitsList := []*event.Unit{}
	for _, unit := range world.Units {
		unitsList = append(unitsList, unit)
	}
	sort.Slice(unitsList, func(i, j int) bool { return unitsList[i].PosY < unitsList[j].PosY })

	for _, u := range unitsList {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1.5, 1.5)

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
	jsonData, _ := json.Marshal(playerInit)
	log.Println(playerInit)

	file, _ := json.MarshalIndent(playerInit, "", " ")
	ioutil.WriteFile("data.json", file, 0600)

	URL := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	wsConnection, _, err := websocket.DefaultDialer.Dial(URL.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	if err = wsConnection.WriteJSON(jsonData); err != nil {
		log.Println(err)
		return
	}

	go getServerData(wsConnection)

	ebiten.SetWindowSize(event.ScreenWidth, event.ScreenHeight)
	ebiten.SetWindowTitle("Golang 2DGame")
	if err := ebiten.RunGame(&Game{Conn: wsConnection, ID: id}); err != nil {
		log.Fatal(err)
	}
}
