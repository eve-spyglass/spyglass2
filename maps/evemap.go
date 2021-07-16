package maps

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/eve-spyglass/spyglass2/engine"
)

type (
	EveMapper struct {
		currentMap  string
		definitions spyglassMapsCollection
		connections []string

		intelResource engine.IntelResource
	}

	spyglassMapsCollection map[string]spyglassMap

	spyglassMap struct {
		Systems map[int32]spyglassSystem
		Width   int
		Height  int

		Name        string
		Author      string
		Description string
	}

	spyglassSystem struct {
		ID       int32  `json:"id"`
		Name     string `json:"name"`
		Icon     string `json:"icon,omitempty"`
		X        int    `json:"x"`
		Y        int    `json:"y"`
		External bool   `json:"external,omitempty"`
	}
)

var (
	//go:embed mapdefs/*.json
	mapdefs embed.FS

	errMapNotDefined = errors.New("map not defined")
)

func NewEveMapper() (*EveMapper, error) {

	col := make(spyglassMapsCollection)

	fs, err := mapdefs.ReadDir("mapdefs")
	if err != nil {
		return nil, err
	}

	for _, fn := range fs {
		f, err := mapdefs.ReadFile("mapdefs/" + fn.Name())
		if err != nil {
			log.Printf("WARN: fs access: %s", err.Error())
			continue
		}

		var def spyglassMap

		err = json.Unmarshal(f, &def)
		if err != nil {
			log.Printf("WARN: failed to parse map: %s", err.Error())
			continue
		}

		log.Printf("MAPS: loaded map %s with %d systems", def.Name, len(def.Systems))

		col[def.Name] = def
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
	i := 0
	for m := range em.definitions {
		maps[i] = m
		i++
	}
	sort.Strings(maps)
	return maps
}

func (em *EveMapper) SetMap(m string) error {
	var mp spyglassMap
	var ok bool
	if mp, ok = em.definitions[m]; !ok {
		return errMapNotDefined
	}
	em.currentMap = m

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
			return err
		}

		//	Now get the connections list
		em.connections = em.intelResource.GetJumps()

		log.Println(len(em.connections))

	} else {
		log.Println("IR unavailable")
	}

	return nil
}

func (em *EveMapper) GetMap() string {
	return em.currentMap
}

func (em *EveMapper) GetCurrentMapSVG() string {

	start := time.Now()

	const systemWidth = 50
	const systemHeight = 22
	const systemRounded = 10

	var mp spyglassMap

	for s, m := range em.definitions {
		if s == em.currentMap {
			mp = m
			break
		}
	}

	statusi := make(map[int32]uint8)
	timei := make(map[int32]time.Time)

	if em.intelResource != nil {
		statusi = em.intelResource.Status()
		timei = em.intelResource.LastUpdated()
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
		if (sok != nil) || (dok != nil) {
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
		st, ok := statusi[s.ID]
		status := uint8(0)
		if ok {
			status = st
		}
		style := "stroke:rgb(0,0,0);stroke-width:1px"
		switch status {
		// Unknown
		case 0:
			style = style + ";fill:rgb(224,224,224)"
			break
		// Clear
		case 1:
			style = style + ";fill:rgb(128,255,128)"
			break
		// Hostile
		case 2:
			style = style + ";fill:rgb(255,128,128)"
			break

		}
		rnd := systemRounded
		if s.External {
			rnd = 0
		}
		canvas.Roundrect(s.X, s.Y, systemWidth, systemHeight, rnd, rnd, style)

		t, tok := timei[s.ID]

		//	create the system name text
		name := s.Name
		stat := "-"
		if tok {
			stat = time.Since(t).Truncate(time.Second).String()
		}
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
