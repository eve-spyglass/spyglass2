package engine

//go:generate go run gen_mapdata.go

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
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

var (
	//go:embed neweden.json
	mapdata []byte
)

func (n NewEden) LoadData() (err error) {
	raw := bytes.NewReader(mapdata)
	jdata := json.NewDecoder(raw)
	err = jdata.Decode(&n)
	return err
}


func (ne NewEden) GetSystem(id int32) (System, error) {
	for _, region := range ne {
		for _, constellation := range region.Constellations {
			for _, system := range constellation.Systems{
				if system.SystemID == id{
					return system, nil
				}
			}
		}
	}

	return System{}, errors.New("system not found")
}