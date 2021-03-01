package engine

import (
	"errors"
	"fmt"
	"github.com/eve-spyglass/spyglass2/feeds"
	"gonum.org/v1/gonum/graph/simple"
	"strconv"
	"time"
)

type (
	IntelEngine struct {
		Galaxy           NewEden
		CurrentMap       string
		monitoredSystems []int32

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

func NewIntelEngine() (*IntelEngine, error) {

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

	return ie, nil
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
	//	TODO implement this
	return make(map[int32]bool)
}

// LastUpdated returns the time since any information was received about a system
func (ie *IntelEngine) LastUpdated() map[int32]time.Time {
	return make(map[int32]time.Time)
}

//	SetSystems will notify the IntelResource which systems to monitor for intel
func (ie *IntelEngine) SetMonitoredSystems(systems []int32) error {
	ie.monitoredSystems = make([]int32, len(systems))
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

