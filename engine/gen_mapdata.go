//+build ignore

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

type (
	UniverseRegions []int32

	UniverseRegion struct {
		Constellations []int32 `json:"constellations,omitempty"`
		Description    string  `json:"description,omitempty"`
		Name           string  `json:"name"`
		RegionID       int32   `json:"region_id"`
	}

	UniverseConstellation struct {
		ConstellationID int32    `json:"constellation_id"`
		Name            string   `json:"name"`
		Position        Position `json:"position"`
		RegionID        int32    `json:"region_id"`
		Systems         []int32  `json:"systems"`
	}

	UniverseSystem struct {
		Name           string         `json:"name,omitempty"`
		Planets        []SystemPlanet `json:"-"`
		Position       Position       `json:"-"`
		SecurityClass  string         `json:"security_class,omitempty"`
		SecurityStatus float64        `json:"security_status"`
		StarID         int32          `json:"-"`
		Stargates      []int32        `json:"stargates,omitempty"`
		Stations       []int32        `json:"-"`
		SystemID       int32          `json:"system_id"`
	}

	SystemPlanet struct {
		AsteroidBelts []int32 `json:"asteroid_belts,omitempty"`
		Moons         []int32 `json:"moons,omitempty"`
		PlanetID      int32   `json:"planet_id"`
	}

	UniverseStargate struct {
		Destination StargateDestination `json:"destination"`
		Name        string              `json:"name"`
		Position    Position            `json:"position"`
		StargateID  int32               `json:"stargate_id"`
		SystemID    int32               `json:"system_id"`
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

	//	These are the types I want to write to file!

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
)

const (
	urlUniverseRegions       = "https://esi.evetech.net/v1/universe/regions/"
	urlUniverseRegion        = "https://esi.evetech.net/v1/universe/regions/%d/"
	urlUniverseConstellation = "https://esi.evetech.net/v1/universe/constellations/%d/"
	urlUniverseSystem        = "https://esi.evetech.net/v4/universe/systems/%d/"
	urlUniverseStargate      = "https://esi.evetech.net/v1/universe/stargates/%d/"
)

func main() {

	log.Println("Removing generated files")

	clearGenFiles()

	log.Println("Starting Map Data Download")

	log.Println("Starting Regions Download")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	//	Start by getting the list of all regions
	var regs UniverseRegions
	err := GetJson(urlUniverseRegions, client, &regs)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get region list: %w", err))
	}

	universeRegions := make(map[int32]UniverseRegion, len(regs))

	//	Now pull in the regions
	workers := 16
	regJobs := make(chan int32, 128)
	regResults := make(chan UniverseRegion, 64)

	for w := 1; w <= workers; w++ {
		go regionWorker(regJobs, regResults, client)
	}

	jobcnt := 0

	for _, r := range regs {
		regJobs <- r
		jobcnt++
	}
	close(regJobs)

	for c := 0; c < jobcnt; c++ {
		reg := <-regResults
		universeRegions[reg.RegionID] = reg
	}
	close(regResults)

	log.Printf("Fetched %d Regions\n", jobcnt)

	//	Now pull in the constellations
	log.Println("Starting Constellations Download")
	workers = 32
	conJobs := make(chan int32, 1024)
	conResults := make(chan UniverseConstellation, 128)

	for w := 1; w <= workers; w++ {
		go constellationWorker(conJobs, conResults, client)
	}

	jobcnt = 0

	for _, r := range universeRegions {
		for _, c := range r.Constellations {
			conJobs <- c
			jobcnt++
		}
	}
	close(conJobs)

	universeConstellations := make(map[int32]UniverseConstellation)
	for c := 0; c < jobcnt; c++ {
		con := <-conResults
		universeConstellations[con.ConstellationID] = con
	}
	close(conResults)

	log.Printf("Fetched %d constellations\n", jobcnt)

	//	Now pull in the systems
	log.Println("Starting Systems Download")
	workers = 64
	sysJobs := make(chan int32, 8096)
	sysResults := make(chan UniverseSystem, 512)

	for w := 1; w <= workers; w++ {
		go systemWorker(sysJobs, sysResults, client)
	}

	jobcnt = 0

	for _, c := range universeConstellations {
		for _, s := range c.Systems {
			sysJobs <- s
			jobcnt++
		}
	}
	close(sysJobs)

	universeSystems := make(map[int32]UniverseSystem)
	for c := 0; c < jobcnt; c++ {
		sys := <-sysResults
		universeSystems[sys.SystemID] = sys
	}
	close(sysResults)

	log.Printf("Fetched %d systems\n", jobcnt)

	//	Now pull in the stargates
	log.Println("Starting Stargates Download")
	workers = 128
	sgJobs := make(chan int32, 16192)
	sgResults := make(chan UniverseStargate, 4096)

	for w := 1; w <= workers; w++ {
		go stargateWorker(sgJobs, sgResults, client)
	}

	jobcnt = 0

	for _, c := range universeSystems {
		for _, s := range c.Stargates {
			sgJobs <- s
			jobcnt++
		}
	}
	close(sgJobs)

	universeStargates := make(map[int32]UniverseStargate)
	for c := 0; c < jobcnt; c++ {
		sg := <-sgResults
		universeStargates[sg.StargateID] = sg
	}
	close(sgResults)

	log.Printf("Fetched %d stargates\n", jobcnt)

	log.Println("Jumping through the EveGate, creating New Eden")

	ne := NewEden{}
	for _, r := range universeRegions {
		region := Region{
			Constellations: make(map[int32]Constellation, len(r.Constellations)),
			Description:    r.Description,
			Name:           r.Name,
			RegionID:       r.RegionID,
		}

		for _, c := range r.Constellations {
			c2 := universeConstellations[c]
			cons := Constellation{
				ConstellationID: c,
				Name:            c2.Name,
				Position:        c2.Position,
				Systems:         make(map[int32]System),
			}

			for _, s := range c2.Systems {
				s2 := universeSystems[s]
				sys := System{
					Name:           s2.Name,
					Planets:        s2.Planets,
					Position:       s2.Position,
					SecurityClass:  s2.SecurityClass,
					SecurityStatus: s2.SecurityStatus,
					StarID:         s2.StarID,
					Stargates:      make(map[int32]Stargate, len(s2.Stargates)),
					Stations:       s2.Stations,
					SystemID:       s,
				}

				for _, sg := range s2.Stargates {
					sg2 := universeStargates[sg]
					stargate := Stargate{
						Destination: sg2.Destination,
						Name:        sg2.Name,
						Position:    sg2.Position,
						StargateID:  sg2.StargateID,
						TypeID:      sg2.TypeID,
					}
					sys.Stargates[sg] = stargate
				}

				cons.Systems[s] = sys
			}

			region.Constellations[c] = cons
		}

		ne[r.RegionID] = region
	}

	log.Println("We have mapped New Eden, jumping in!")

	// Save the raw new eden data to json
	f, err := os.OpenFile("neweden.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	bf := bufio.NewWriter(f)
	//gw := gzip.NewWriter(bf)
	enc := json.NewEncoder(bf)
	enc.SetIndent("", "\t")
	err = enc.Encode(ne)
	if err != nil {
		log.Fatalln(err)
	}
	f.Sync()

	log.Println("DONE!")

}

func clearGenFiles() {
	files := []string{"neweden.json", "mapdata.go"}

	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			log.Printf("WARN: %s", err.Error())
		}
	}
}

func regionWorker(jobs <-chan int32, results chan<- UniverseRegion, client http.Client) {
	for id := range jobs {
		var con UniverseRegion
		err := GetJson(fmt.Sprintf(urlUniverseRegion, id), client, &con)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to query region %d: %w", id, err))
		}
		results <- con
	}
}

func constellationWorker(jobs <-chan int32, results chan<- UniverseConstellation, client http.Client) {
	for id := range jobs {
		var con UniverseConstellation
		err := GetJson(fmt.Sprintf(urlUniverseConstellation, id), client, &con)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to query constellation %d: %w", id, err))
		}
		results <- con
	}
}

func systemWorker(jobs <-chan int32, results chan<- UniverseSystem, client http.Client) {
	for id := range jobs {
		var con UniverseSystem
		err := GetJson(fmt.Sprintf(urlUniverseSystem, id), client, &con)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to query system %d: %w", id, err))
		}
		results <- con
	}
}

func stargateWorker(jobs <-chan int32, results chan<- UniverseStargate, client http.Client) {
	for id := range jobs {
		var con UniverseStargate
		err := GetJson(fmt.Sprintf(urlUniverseStargate, id), client, &con)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to query system %d: %w", id, err))
		}
		results <- con
	}
}

func GetJson(url string, client http.Client, dest interface{}) (err error) {
	retries := 8
	for retries > 0 {
		retries--

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", "Crypta Electrica - Spyglass Map Gen")
		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to make request: %w", err)
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		if res.StatusCode != http.StatusOK {
			// If needed, log this
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}
		err = json.Unmarshal(body, dest)
		if err != nil {
			return fmt.Errorf("failed to decode json: body: %s: %w", string(body), err)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("retries exceeded: url %s", url))
}

var packageTemplate = template.Must(template.New("gen").Parse(
	`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}
// using data from CCPs ESI API
package engine

var mapdata = {{ printf "%#v" .Data }}
`))
