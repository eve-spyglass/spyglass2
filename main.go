package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eve-spyglass/spyglass2/config"
	"github.com/eve-spyglass/spyglass2/engine"
	"github.com/eve-spyglass/spyglass2/feeds"
	"github.com/eve-spyglass/spyglass2/maps"
	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails"

	wailsLogger "github.com/wailsapp/wails/lib/logger"
)

var (
	guaranteedError = errors.New("guaranteed error, PLEASE IGNORE ")

	//go:embed frontend/dist/app.js
	js string

	//go:embed frontend/dist/app.css
	css string
)

func basic() string {
	return "Hello World!"
}

func main() {

	ctx := context.Background()

	cfg := config.NewConfig()
	cfg.LoadConfig()

	// Setup a ticker to send update events to the UI.
	ui := &UserInterface{
		errors: make([]string, 0),
	}

	// For testing purposes
	go func() {
		time.Sleep(40 * time.Second)
		ui.errors = append(ui.errors, guaranteedError.Error())
	}()

	// Set the logger to write out to the correct directory
	outlogWails := filepath.Join(cfg.GetConfigDirectory(), "spyglass_log_wails.log")

	fi, err := os.OpenFile(outlogWails, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ui.errors = append(ui.errors, fmt.Sprintf("failed to create wails logfile: %s", err))
		fi.Close()
	} else {
		fileLogger := logrus.New()
		fileLogger.SetOutput(fi)
		fileLogger.SetLevel(logrus.ErrorLevel)
		fileLogger.Formatter = &logrus.TextFormatter{}
		wailsLogger.GlobalLogger = fileLogger
	}

	// Set the logger to write out to the correct directory
	outlogDefault := filepath.Join(cfg.GetConfigDirectory(), "spyglass_log_default.log")

	fi2, err := os.OpenFile(outlogDefault, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ui.errors = append(ui.errors, fmt.Sprintf("failed to create spyglass logfile: %s", err))
		fi.Close()
	} else {
		fileLogger := logrus.New()
		fileLogger.SetOutput(fi2)
		fileLogger.SetLevel(logrus.DebugLevel)
		fileLogger.Formatter = &logrus.TextFormatter{}
		log.SetOutput(fileLogger.Writer())
	}

	log.Println("Starting Spyglass2")

	// Will need to move this to backend eventually

	lw := feeds.LogFeed{}
	err = lw.SetLogDir(cfg.Data.ChatLogDirectory)
	if err != nil {
		ui.errors = append(ui.errors, fmt.Sprintf("failed to set log directory: %s", err))
		log.Printf("FATAL: SET YOUR LOG DIRECTORY: %s", err)
	}
	errs := make(chan error, 32)
	lw.SetChatRooms(cfg.Data.Channels)
	go func() {
		for {
			select {
			case err := <-errs:
				// TODO UNSAFE
				ui.errors = append(ui.errors, fmt.Sprintf("watcher error: %s", err))
				log.Printf("Got Watcher Error: %#v", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println("Starting intel engine")

	ie, err := engine.NewIntelEngine(ctx)
	if err != nil {
		ui.errors = append(ui.errors, fmt.Sprintf("failed to init intel engine: %s", err))
		log.Fatalln(fmt.Errorf("failed to init intel engine: %w", err))
	}

	em, err := maps.NewEveMapper()
	if err != nil {
		ui.errors = append(ui.errors, fmt.Sprintf("failed to create mapper: %s", err))
		log.Fatalln(fmt.Errorf("failed to create mapper: %s", err))
	}

	em.SetIntelResource(ie)

	ie.SetCurrentMap(cfg.Data.SelectedMap)
	em.SetMap(cfg.Data.SelectedMap)

	ie.SetClearWords(cfg.Data.ClearWords)

	// Set the log Watcher Feeders

	reports, locations, _ := ie.GetFeeders()
	go func() {
		err = lw.Feed(ctx, reports, locations, errs)
		if err != nil {
			// TODO UNSAFE APPEND HERE
			ui.errors = append(ui.errors, fmt.Sprintf("failed to start log feed: %s", err))
		}
	}()

	//START FRONTEND

	ui.intelEngine = ie

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
	//app.Bind(ie)

	app.Bind(ui)

	log.Println("APP RUN")

	err = app.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

// TODO Have a better place for all below this line

type (
	UserInterface struct {
		runtime *wails.Runtime
		errors  []string

		intelEngine *engine.IntelEngine
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

func (ui *UserInterface) ReadErrorList() []string {
	return ui.errors
}

func (ui *UserInterface) GetIntelMessages() []string {
	return ui.intelEngine.GetIntelMessages()
}
