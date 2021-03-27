package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eve-spyglass/spyglass2/config"
	"github.com/eve-spyglass/spyglass2/engine"
	"github.com/eve-spyglass/spyglass2/feeds"
	"github.com/eve-spyglass/spyglass2/maps"
	"github.com/leaanthony/mewn"
	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails"

	wailsLogger "github.com/wailsapp/wails/lib/logger"
)

func basic() string {
	return "Hello World!"
}

func main() {

	ctx := context.Background()

	cfg := config.NewConfig()
	cfg.LoadConfig()

	// Set the logger to write out to the correct directory
	outlogWails := filepath.Join(cfg.GetConfigDirectory(), "spyglass_log_wails.log")
	// outlogDefault := filepath.Join(cfg.GetConfigDirectory(), "spyglass_log_default.log")

	fi, err := os.OpenFile(outlogWails, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fi.Close()
	} else {
		fileLogger := logrus.New()
		fileLogger.SetOutput(fi)
		fileLogger.SetLevel(logrus.DebugLevel)
		fileLogger.Formatter = &logrus.TextFormatter{}
		wailsLogger.GlobalLogger = fileLogger
	}

	log.Println("Starting Spyglass2")

	// Will need to move this to backend eventually

	lw := feeds.LogFeed{}
	err = lw.SetLogDir(cfg.Data.ChatLogDirectory)
	if err != nil {
		log.Printf("FATAL: SET YOUR LOG DIRECTORY: %w", err)
	}
	errs := make(chan error, 32)
	lw.SetChatRooms(cfg.Data.Channels)
	go func() {
		for {
			select {
			case err := <-errs:
				log.Printf("Got Watcher Error: %#v", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println("Starting intel engine")

	ie, err := engine.NewIntelEngine(ctx)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to init intel engine: %w", err))
	}

	em, err := maps.NewEveMapper()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to create mapper: %w", err))
	}

	em.SetIntelResource(ie)

	ie.SetCurrentMap(cfg.Data.SelectedMap)
	em.SetMap(cfg.Data.SelectedMap)

	// Set the log Watcher Feeders

	reports, locations := ie.GetFeeders()
	go func() {
		err = lw.Feed(ctx, reports, locations, errs)
		if err != nil {
			panic(err)
		}
	}()

	//_ = em.GetCurrentMapSVG()

	//time.Sleep(10 * time.Minute)
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
	app.Bind(cfg)
	app.Bind(em)

	// Setup a ticker to send update events to the UI.
	ui := &UserInterface{}
	app.Bind(ui)

	err = app.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

// TODO Have a better place for all below this line

type (
	UserInterface struct {
		runtime *wails.Runtime
	}
)

func (ui *UserInterface) WailsInit(runtime *wails.Runtime) error {
	go func() {
		time.Sleep(3 * time.Second)
		t := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-t.C:
				log.Println("Updating UI")
				runtime.Events.Emit("ui_update")
			}
		}
	}()

	return nil
}
