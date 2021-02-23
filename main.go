package main

import (
	"context"
	"fmt"
	"github.com/leaanthony/mewn"
	"github.com/wailsapp/wails"
	"log"
	"spyglass-2/engine"
	//"spyglass-2/feeds"
	"spyglass-2/maps"
	"time"
)

func basic() string {
	return "Hello World!"
}

func main() {

	_ = context.Background()

	// Will need to move this to backend eventually

	//lw := feeds.LogFeed{}
	//dir, err := lw.LogDirHint()
	//if err != nil {
	//	panic(err)
	//}
	//err = lw.SetLogDir(dir)
	//if err != nil {
	//	panic(err)
	//}
	//reports := make(chan feeds.Report, 32)
	//locations := make(chan feeds.Locstat, 32)
	//errs := make(chan error, 32)
	//lw.SetChatRooms([]string{"asdfghjkl"})
	//go func() {
	//	for {
	//		select {
	//		case r := <-reports:
	//			log.Printf("Got Intel Report: %#v", r)
	//		case err := <-errs:
	//			log.Printf("Got Watcher Error: %#v", err)
	//		case l := <-locations:
	//			log.Printf("Got Location Report: %#v", l)
	//
	//		case <-ctx.Done():
	//			return
	//		}
	//	}
	//}()
	//
	//go func() {
	//	err = lw.Feed(ctx, reports, locations, errs)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
	//
	//log.Println("Starting intel engine")

	ie, err := engine.NewIntelEngine()
	if err != nil{
		log.Fatalln(fmt.Errorf("failed to init intel engine: %w", err))
	}
	_ = ie.CurrentMap

	em, err := maps.NewEveMapper()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to create mapper: %w", err))
	}

	em.SetMap("Catch")

	em.SetIntelResource(ie)

	fmt.Println(em.GetCurrentMapSVG())

	time.Sleep(10 * time.Minute)
	//START FRONTEND

	js := mewn.String("./frontend/dist/app.js")
	css := mewn.String("./frontend/dist/app.css")

	app := wails.CreateApp(&wails.AppConfig{
		Width:     1024,
		Height:    768,
		Resizable: true,
		Title:     "Spyglass 2",
		JS:        js,
		CSS:       css,
		Colour:    "#ff6666",
	})
	app.Bind(basic)
	err = app.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
