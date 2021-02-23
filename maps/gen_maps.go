// +build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/anaskhan96/soup"
	"log"
	"os"
	"strconv"
)

var (
	dotlanMaps = []string{
		"Aridia",
		"Black_Rise",
		"The_Bleak_Lands",
		"Branch",
		"Cache",
		"Catch",
		"The_Citadel",
		"Cloud_Ring",
		"Cobalt_Edge",
		"Curse",
		"Deklein",
		"Delve",
		"Derelik",
		"Detorid",
		"Devoid",
		"Domain",
		"Esoteria",
		"Essence",
		"Etherium_Reach",
		"Everyshore",
		"Fade",
		"Feythabolis",
		"The_Forge",
		"Fountain",
		"Geminate",
		"Genesis",
		"Great_Wildlands",
		"Heimatar",
		"Immensea",
		"Impass",
		"Insmother",
		"Kador",
		"The_Kalevala_Expanse",
		"Khanid",
		"Kor-Azor",
		"Lonetrek",
		"Malpais",
		"Metropolis",
		"Molden_Heath",
		"Oasa",
		"Omist",
		"Outer_Passage",
		"Outer_Ring",
		"Paragon_Soul",
		"Period_Basis",
		"Perrigen_Falls",
		"Placid",
		"Pochven",
		"Providence",
		"Pure_Blind",
		"Querious",
		"Scalding_Pass",
		"Sinq_Laison",
		"Solitude",
		"The_Spire",
		"Stain",
		"Syndicate",
		"Tash-Murkon",
		"Tenal",
		"Tenerifis",
		"Tribute",
		"Vale_of_the_Silent",
		"Venal",
		"Verge_Vendor",
		"Wicked_Creek",
	}
)

const (
	urlDotlanMap = "https://evemaps.dotlan.net/svg/%s.svg"
)

type (
	MapCollection map[string]EveMap
	EveMap        struct {
		Systems map[int32]EveSystem `json:"systems"`
		Width   int32               `json:"width"`
		Height  int32               `json:"height"`
	}
	EveSystem struct {
		ID int32 `json:"id"`
		X  int   `json:"x"`
		Y  int   `json:"y"`
	}
)

func main() {

	theLot := make(MapCollection, len(dotlanMaps))

	for _, dotlanMap := range dotlanMaps {

		log.Printf("Downloda map: %s", dotlanMap)

		thisMap := EveMap{
			Systems: make(map[int32]EveSystem),
			Width:   1024, //TODO Actually detect the svg width
			Height:  768,  // TODO Actually detect the svg height
		}

		url := fmt.Sprintf(urlDotlanMap, dotlanMap)
		resp, err := soup.Get(url)
		if err != nil {
			log.Printf("WARN: Dotlan Map Download Failed: %s", err.Error())
			continue
		}
		doc := soup.HTMLParse(resp)
		systems := doc.Find("g", "id", "sysuse").FindAll("use")
		for _, sys := range systems {
			//log.Printf("\t%s, \t%s\t(%s,%s)\n", dotlanMap, sys.Attrs()["id"], sys.Attrs()["x"], sys.Attrs()["y"])
			id := sys.Attrs()["id"]
			i, err := strconv.Atoi(id[3:])
			if err != nil {
				log.Fatal(fmt.Errorf("failed to get system id. map (%s), ident (%s). err: %w", dotlanMap, id, err))
			}
			x, err := strconv.Atoi(sys.Attrs()["x"])
			if err != nil {
				log.Fatal(fmt.Errorf("failed to get system x. map (%s), ident (%s). err: %w", dotlanMap, sys.Attrs()["x"], err))
			}
			y, err := strconv.Atoi(sys.Attrs()["y"])
			if err != nil {
				log.Fatal(fmt.Errorf("failed to get system y. map (%s), ident (%s). err: %w", dotlanMap, sys.Attrs()["y"], err))
			}
			sys := EveSystem{
				ID: int32(i),
				X:  x,
				Y:  y,
			}
			thisMap.Systems[int32(i)] = sys
		}

		theLot[dotlanMap] = thisMap
	}

	f, err := os.Create("maplayout.json")
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to create map layout file: %w", err))
	}

	defer f.Close()

	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	err = enc.Encode(theLot)

	bw.Flush()
	f.Sync()
}
