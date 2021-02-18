package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"io/ioutil"
)

type (
	NewEden map[int32]Region

	Region struct {
		Constellations map[int32]Constellation `json:"constellations,omitempty"`
		Description    string                  `json:"description,omitempty"`
		Name           string                  `json:"name"`
		RegionID       int32                   `json:"region_id"`
	}

	Constellation struct {
		ConstellationID int32            `json:"constellation_id"`
		Name            string           `json:"name"`
		Position        Position         `json:"position"`
		Systems         map[int32]System `json:"systems"`
	}

	System struct {
		Name           string             `json:"name,omitempty"`
		Planets        []SystemPlanet     `json:"planets"`
		Position       Position           `json:"position"`
		SecurityClass  string             `json:"security_class,omitempty"`
		SecurityStatus float64            `json:"security_status"`
		StarID         int32              `json:"star_id,omitempty"`
		Stargates      map[int32]Stargate `json:"stargates,omitempty"`
		Stations       []int32            `json:"stations,omitempty"`
		SystemID       int32              `json:"system_id"`
	}

	Stargate struct {
		Destination StargateDestination `json:"destination"`
		Name        string              `json:"name"`
		Position    Position            `json:"-"`
		StargateID  int32               `json:"stargate_id"`
		TypeID      int32               `json:"type_id"`
	}

	StargateDestination struct {
		StargateID int32 `json:"stargate_id"`
		SystemID   int32 `json:"system_id"`
	}

	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	}

	SystemPlanet struct {
		AsteroidBelts []int32 `json:"asteroid_belts,omitempty"`
		Moons         []int32 `json:"moons,omitempty"`
		PlanetID      int32   `json:"planet_id"`
	}
)

func (n NewEden) LoadData() (err error) {
	raw := bytes.NewReader(mapdata)
	decompressor := brotli.NewReader(raw)
	//jdata := json.NewDecoder(decompressor)

	bss, err := ioutil.ReadAll(decompressor)
	fmt.Println(string(bss))
	return json.Unmarshal(bss, &n)
	//err = jdata.Decode(&n)
	//return err
}
