package maps

//go:generate go run -race gen_maps.go

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	svg "github.com/ajstarks/svgo"
	"golang.org/x/exp/rand"
	"log"
	"spyglass-2/engine"
	"strconv"
	"strings"
	"time"
)

type (
	EveMapper struct {
		currentMap  string
		definitions mapCollection
		connections []string

		intelResource engine.IntelResource
	}

	mapCollection map[string]eveMap
	eveMap        struct {
		Systems map[int32]eveSystem `json:"systems"`
		Width   int32               `json:"width"`
		Height  int32               `json:"height"`
	}
	eveSystem struct {
		ID int32 `json:"id"`
		X  int   `json:"x"`
		Y  int   `json:"y"`
	}

	spyglassMapsCollection map[string]eveMap

	spyglassMap struct {
		Systems map[int32]spyglassSystem
		Width int32
		Height int32


		Name string
		Author string
		Description string
	}

	spyglassSystem struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
		Icon string `json:"icon,omitempty"`
		X    int32  `json:"x"`
		Y    int32  `json:"y"`
		External bool `json:"external,omitempty"`
	}
)

var (
	//go:embed maplayout.json
	mapdefs []byte

	errMapNotDefined = errors.New("map not defined")
)

func NewEveMapper() (*EveMapper, error) {

	col := make(mapCollection)
	err := json.Unmarshal(mapdefs, &col)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mapdefs: %w", err)
	}

	mapper := &EveMapper{
		definitions: col,
		connections: make([]string, 0),
	}

	// TODO load previous map on startup
	err = mapper.SetMap("Delve")

	return mapper, err
}

func (em *EveMapper) SetIntelResource(source engine.IntelResource) {
	em.intelResource = source
}

func (em *EveMapper) GetAvailableMaps() (maps []string) {
	maps = make([]string, len(em.definitions))
	for m := range em.definitions {
		maps = append(maps, m)
	}
	return maps
}

func (em *EveMapper) SetMap(m string) error {
	if _, ok := em.definitions[m]; !ok {
		return errMapNotDefined
	}
	em.currentMap = m
	return nil
}

func (em *EveMapper) GetMap() string {
	return em.currentMap
}

func (em *EveMapper) GetCurrentMapSVG() (string) {

	start := time.Now()

	const systemWidth = 50
	const systemHeight = 22
	const systemRounded = 10

	var mp eveMap

	for s, m := range em.definitions {
		if s == em.currentMap {
			mp = m
			break
		}
	}

	// Load connections from intel resource if it exists
	if em.intelResource != nil {
		log.Println("IR is available")
		//	Make sure we have the correct systems monitored
		systemIDs := make([]int32, len(mp.Systems))
		for id := range mp.Systems {
			systemIDs = append(systemIDs, id)
		}

		err := em.intelResource.SetMonitoredSystems(systemIDs)
		if err != nil {
			return ""
		}

		//	Now get the connections list
		em.connections = em.intelResource.GetJumps()

		log.Println(len(em.connections))

	} else {
		log.Println("IR unavailable")
	}

	var buf bytes.Buffer

	canvas := svg.New(&buf)
	canvas.Start(int(mp.Width), int(mp.Height))

	// First draw all of the connections so that they are beneath all other things. Keep them in their own group

	canvas.Gid("jumps")
	for _, con := range em.connections {
		sp := strings.Split(con, "-")
		if len(sp) != 2 {
			continue
		}

		source, sok := strconv.Atoi(sp[0])
		dest, dok := strconv.Atoi(sp[1])
		if ((sok != nil) || (dok != nil)) {
			log.Println(sp[0], sp[1])
			log.Println("Not ints")
			continue
		}

		src, srok := mp.Systems[int32(source)]
		dst, dtok := mp.Systems[int32(dest)]
		if !(srok || dtok) {
			log.Println(con)
			log.Println("not present")
			continue
		}
		// Get middle point of source system
		startX := src.X + (systemWidth / 2)
		startY := src.Y + (systemHeight / 2)

		//	Get middle point of destination system
		endX := dst.X + (systemWidth / 2)
		endY := dst.Y + (systemHeight / 2)

		// TODO implement line colours
		// TODO investigate use of beziers
		canvas.Line(startX, startY, endX, endY, "stroke:rgb(0,0,0);stroke-width:1px")
	}
	canvas.Gend()

	//	Now add all of the systems to the map
	// Each system is a rounded rect with a height of 30, width of 62, r of 10
	canvas.Gid("systems")
	for _, s := range mp.Systems {
		// Start an individual group for each system
		canvas.Gid(strconv.Itoa(int(s.ID)))
		status := rand.Float32() > 0.5
		style := "fill:rgb(255,255,255);stroke:rgb(0,0,0);stroke-width:1px"
		if status {
			style = "fill:rgb(255,128,128);stroke:rgb(0,0,0);stroke-width:1px"
		}
		canvas.Roundrect(s.X, s.Y, systemWidth, systemHeight, systemRounded, systemRounded, style)

		//	create the system name text
		name := strconv.Itoa(int(s.ID))
		stat := "STATUS!"
		x := s.X + (systemWidth / 2)
		yn := s.Y + (systemHeight / 2)
		ys := s.Y + (systemHeight * 7 / 8)

		canvas.Text(x, yn, name, "text-anchor:middle;font-size:9px")
		canvas.Text(x, ys, stat, "text-anchor:middle;font-size:8px")
		canvas.Gend()
	}

	canvas.Gend()

	canvas.End()

	log.Printf("Generation took %v", time.Since(start))

	return buf.String()
}
