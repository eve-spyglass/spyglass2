package engine

//go:generate go run -race gen_mapdata.go

import (
	"errors"
	"fmt"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
	"spyglass-2/feeds"
)

type (
	IntelEngine struct {
		Galaxy     NewEden
		CurrentMap string

		locationInput chan<- feeds.Locstat
		intelInput    chan<- feeds.Report

		//reportLibrary map[string]feeds.Report

		mapGraph *simple.UndirectedGraph
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

			out, _ := dot.Marshal(ie.mapGraph, ie.CurrentMap, "", "\t")
			fmt.Println(string(out))

			return nil
		}
	}
	return errors.New("map not found")
}
