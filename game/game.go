package game

import (
	"encoding/json"
	"log"
	"math"

	uuid "github.com/satori/go.uuid"
)

//Unit >>> Player
type Unit struct {
	ID        string  `json:"id"`
	PosX      float64 `json:"posx"`
	PosY      float64 `json:"posy"`
	Direction int     `json:"dir"`
	State     string  `json:"state"`
	HorizDir  int     `json:"horiz_dir"`
	Frame     float64 `json:"frame"`
}

//Units >>> Players
type Units map[string]*Unit

//World >>> Mainroom
type World struct {
	Units Units `json:"units"`
}

//Event >>> State
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

//EventIdle >>> Idle
type EventIdle struct {
	UnitID string `json:"unit_id"`
}

//EventMove >>> Move
type EventMove struct {
	UnitID      string `json:"unit_id"`
	Direction   int    `json:"unit_dir"`
	FrameUpdate bool   `json:"frame_upd"`
}

//EventInit >>> Init
type EventInit struct {
	ID   string `json:"player_id"`
	Unit Unit   `json:"units"`
}

const (
	ScreenWidth  int = 640
	ScreenHeight int = 480

	EventTypeInit   string = "init"
	EventTypeMove   string = "move"
	EventTypeIdle   string = "idle"
	EventTypeUpdate string = "update"

	PlayerSpeed    float64 = 2
	DirectionNone  int     = 0
	DirectionUp    int     = 1
	DirectionDown  int     = 2
	DirectionLeft  int     = 3
	DirectionRight int     = 4

	AnimationSpeed float64 = 0.2
	FrameWidth     int     = 21
	FrameHeight    int     = 33
	FrameNum       int     = 8
	FrameOX        int     = 0
	FrameOY        int     = 0
)

//PlayerInit function to create player unit
func PlayerInit() (Event, string) {
	id := uuid.Must(uuid.NewV4())
	log.Printf("UUIDv4: %s\n", id)

	return Event{
		Type: EventTypeInit,
		Data: EventInit{
			ID: id.String(),
			Unit: Unit{
				ID:        id.String(),
				PosX:      float64(ScreenWidth / 2),
				PosY:      float64(ScreenHeight / 2),
				Direction: DirectionNone,
				State:     EventTypeIdle,
				HorizDir:  DirectionRight,
				Frame:     0,
			},
		},
	}, id.String()
}

//EventHandler function to create world
func (world *World) EventHandler(event Event) {
	bytedata, _ := json.Marshal(event.Data)

	updateFrame := func(id string) {
		if math.Floor(world.Units[id].Frame*AnimationSpeed) >= 7 {
			world.Units[id].Frame = 0
		} else {
			world.Units[id].Frame++
		}
	}

	switch event.Type {

	case EventTypeInit:
		var data EventInit
		json.Unmarshal(bytedata, &data)

		world.Units[data.ID] = &data.Unit

	case EventTypeUpdate:
		var data Units
		json.Unmarshal(bytedata, &data)
		world.Units = data

	case EventTypeMove:
		var data EventMove
		json.Unmarshal(bytedata, &data)
		world.Units[data.UnitID].Direction = data.Direction

		switch world.Units[data.UnitID].Direction {
		case DirectionUp:
			world.Units[data.UnitID].PosY -= PlayerSpeed
		case DirectionDown:
			world.Units[data.UnitID].PosY += PlayerSpeed
		case DirectionLeft:
			world.Units[data.UnitID].PosX -= PlayerSpeed
			world.Units[data.UnitID].HorizDir = DirectionLeft
		case DirectionRight:
			world.Units[data.UnitID].PosX += PlayerSpeed
			world.Units[data.UnitID].HorizDir = DirectionRight
		}

		world.Units[data.UnitID].State = EventTypeMove
		if data.FrameUpdate {
			updateFrame(data.UnitID)
		}

	case EventTypeIdle:
		var data EventMove
		json.Unmarshal(bytedata, &data)
		world.Units[data.UnitID].State = EventTypeIdle
		updateFrame(data.UnitID)
	}
}
