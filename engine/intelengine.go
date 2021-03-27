package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/agext/levenshtein"
	"github.com/eve-spyglass/spyglass2/feeds"
	"gonum.org/v1/gonum/graph/simple"
)

type (
	IntelEngine struct {
		Galaxy           NewEden
		CurrentMap       string
		monitoredSystems []int32

		clearWords []string

		reportHistory   []feeds.Report
		locationHistory []feeds.Locstat

		currentStatus map[int32]bool
		lastUpdated   map[int32]time.Time

		locationInput chan feeds.Locstat
		intelInput    chan feeds.Report

		mapGraph *simple.UndirectedGraph
	}

	IntelResource interface {
		// Status returns a map of systems to status, where true is hostile and false is clear
		Status() map[int32]bool
		// LastUpdated returns the time since any information was received about a system
		LastUpdated() map[int32]time.Time
		//	SetSystems will notify the IntelResource which systems to alarm upon
		SetMonitoredSystems(systems []int32) error
		// GetJumps will return the connections between the given systems
		// it returns a string array where each string represents a connection
		// it will be formatted as "1234-5678" and is directional from source to sink
		GetJumps() []string
		// GetFeeders will return the two channels that can e used to feed information into the resource
		GetFeeders() (chan<- feeds.Report, chan<- feeds.Locstat)
	}
)

func NewIntelEngine(ctx context.Context) (*IntelEngine, error) {

	galaxy := make(NewEden)
	err := galaxy.LoadData()
	if err != nil {
		return nil, fmt.Errorf("failed to load galaxy data: %w", err)
	}

	ie := &IntelEngine{
		Galaxy:     galaxy,
		CurrentMap: "Delve",
	}

	err = ie.updateMapGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to update map graph: %w", err)
	}

	reps := make(chan feeds.Report, 64)
	locs := make(chan feeds.Locstat, 64)

	ie.intelInput = reps
	ie.locationInput = locs

	err = ie.startListeners(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start intel listeners")
	}

	return ie, nil
}

func (ie *IntelEngine) SetCurrentMap(m string) error {
	ie.CurrentMap = m
	return ie.updateMapGraph()
}

func (ie *IntelEngine) SetClearWords(words []string) {
	ie.clearWords = words
}

func (ie *IntelEngine) updateMapGraph() error {
	// TODO change this to account for non region mapdefs
	//	Find the correct region based on the current selected map
	for _, r := range ie.Galaxy {
		if r.Name == ie.CurrentMap {
			// This is us!
			ie.mapGraph = simple.NewUndirectedGraph()

			for _, c := range r.Constellations {
				for _, s := range c.Systems {
					for _, g := range s.Stargates {
						// Cant have a system link to itself
						if s.SystemID == g.Destination.SystemID {
							continue
						}

						ie.mapGraph.SetEdge(ie.mapGraph.NewEdge(simple.Node(s.SystemID), simple.Node(g.Destination.SystemID)))
					}
				}
			}

			return nil
		}
	}
	return errors.New("map not found")
}

func (ie *IntelEngine) startListeners(ctx context.Context) error {

	go func() {
		log.Println("DEBUG: IE: Starting to Listen")
		for {
			select {
			case rep := <-ie.intelInput:
				// Received a new intel report
				ie.reportHistory = append(ie.reportHistory, rep)
				ie.checkReport(rep)
				log.Printf("IE: Got Intel - %s", rep.Message)

			case loc := <-ie.locationInput:
				// Received a new location report
				log.Printf("IE - Got locstat - %s", loc.Character)
			case <-ctx.Done():
				panic(1)
				// return
			}
		}
	}()

	return nil
}

func (ie *IntelEngine) checkReport(rep feeds.Report) {
	// Now we need to check each part of the message for potential matches to monitored system names.
	msgParts := strings.Split(rep.Message, " ")

	// TODO: Make these configurable
	const dist = 0.8
	var ignores = []string{"in", "as", "is"}

	var systems []int32
	status := true

	for _, word := range msgParts {
		lowerWord := strings.ToLower(word)
		for _, i := range ignores {
			if lowerWord == strings.ToLower(i) {
				continue
			}
		}
		for _, s := range ie.monitoredSystems {
			system, err := ie.Galaxy.GetSystem(s)
			if err != nil {
				continue
			}

			// D will be in a range of 0 to 1, where 1 is a perfect match
			d := levenshtein.Match(lowerWord, strings.ToLower(system.Name), levenshtein.NewParams().BonusPrefix(3).BonusThreshold(0.3).BonusScale(0.21))
			log.Printf("DEBUG: IE: MATCHER %s to %s with a distance of %.2f", word, system.Name, d)
			if d >= dist {
				// We have a system match here! Yay, intel!
				log.Printf("DEBUG: IE: Matched %s to %s with a distance of %.2f", word, system.Name, d)
				systems = append(systems, system.SystemID)
				break
			}
		}

		for _, cw := range ie.clearWords {
			if lowerWord == strings.ToLower(cw) {
				status = false
			}
		}
	}

	for _, sys := range systems {
		ie.currentStatus[sys] = status
		ie.lastUpdated[sys] = rep.Time
	}

}

func (ie *IntelEngine) IsSystemMonitored(sys int32) bool {
	for _, s := range ie.monitoredSystems {
		if s == sys {
			return true
		}
	}
	return false
}

// The following methods are to satisfy the IntelResource interface

// Status returns a map of systems to status, where true is hostile and false is clear
func (ie *IntelEngine) Status() map[int32]bool {
	// m := make(map[int32]bool, len(ie.monitoredSystems))
	// for _, v := range ie.monitoredSystems {
	// 	m[v] = rand.Float32() > 0.5
	// }

	// return m

	return ie.currentStatus
}

// LastUpdated returns the time since any information was received about a system
func (ie *IntelEngine) LastUpdated() map[int32]time.Time {
	// t := make(map[int32]time.Time, len(ie.monitoredSystems))
	// for _, v := range ie.monitoredSystems {
	// 	t[v] = time.Now().Add(-1 * time.Second * time.Duration(rand.Intn(300)))
	// }
	// return t

	return ie.lastUpdated
}

//	SetSystems will notify the IntelResource which systems to monitor for intel
func (ie *IntelEngine) SetMonitoredSystems(systems []int32) error {
	ie.monitoredSystems = make([]int32, len(systems))
	ie.currentStatus = make(map[int32]bool, len(systems))
	ie.lastUpdated = make(map[int32]time.Time, len(systems))

	for _, system := range systems {
		sys, err := ie.Galaxy.GetSystem(system)
		if err != nil {
			continue
		}
		ie.monitoredSystems = append(ie.monitoredSystems, sys.SystemID)
	}
	return nil
}

// GetJumps will return the connections between the monitored systems
// This list will contain both directions ie 1 -> 2 and 2 -> 1
func (ie *IntelEngine) GetJumps() []string {
	// TODO find a way to preallocate this to some extent
	jumps := make([]string, 0)

	for _, s := range ie.monitoredSystems {
		source, err := ie.Galaxy.GetSystem(s)
		if err != nil {
			// TODO this shouldnt ever happen so I probably shouldnt be silent here but will do for now
			continue
		}

		for _, gate := range source.Stargates {
			if ie.IsSystemMonitored(gate.Destination.SystemID) {
				jumps = append(jumps, strconv.Itoa(int(source.SystemID))+"-"+strconv.Itoa(int(gate.Destination.SystemID)))
			}
		}
	}
	return jumps
}

func (ie *IntelEngine) GetFeeders() (chan<- feeds.Report, chan<- feeds.Locstat) {
	return ie.intelInput, ie.locationInput
}
