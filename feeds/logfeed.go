package feeds

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type (
	LogFeed struct {
		watching   bool
		chatlogDir string
		roomnames  []string

		loglines map[string]uint64
	}
)

var (
	LogFilesNotFound       = errors.New("failed to locate eve log directory, please set the correct directory in the settings")
	PlatformNotImplemented = errors.New("platform not yet implemented")
	PlatformNotSupported   = errors.New("platform requires manual log file directory selection")

	AlreadyWatching = errors.New("cannot change log dir once already watching")

	codec = unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM)
)

const (
	logTimeFormat = "2006.01.02 15:04:05"
)

func (f *LogFeed) CheckLogDir(dir string) (valid bool) {
	reqs := []string{"Chatlogs", "Fleetlogs", "Gamelogs", "Marketlogs"}

	d, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("log check ioutil error: %s", err.Error())
		return false
	}

	ls := make([]string, len(d))
	for i, f := range d {
		ls[i] = f.Name()
	}

	log.Printf("log watcher directory check", ls)

	valid = true

	for _, r := range reqs {
		found := false
		for _, f := range d {
			if f.IsDir() && f.Name() == r {
				found = true
				break
			}
		}
		if !found {
			valid = false
			break
		}
	}

	return valid
}

func (f *LogFeed) SetLogDir(dir string) error {
	log.Printf("LW: setting log dir to %v\n", dir)
	if f.watching {
		return AlreadyWatching
	}
	f.chatlogDir = filepath.Join(dir, "Chatlogs")
	return nil
}

func (f *LogFeed) SetChatRooms(rooms []string) {
	f.roomnames = rooms
}

func (f *LogFeed) GetChatRooms() (rooms []string) {
	return f.roomnames
}

func (f *LogFeed) Feed(ctx context.Context, reps chan<- Report, locs chan<- Locstat, errs chan<- error) (err error) {

	valid := f.CheckLogDir(filepath.Dir(f.chatlogDir))
	if !valid {
		return LogFilesNotFound
	}

	log.Printf("LW: Starting the logwatcher! - DIR: %s", f.chatlogDir)

	if f.loglines == nil {
		f.loglines = make(map[string]uint64)
	}

	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Write)
	go func() {
		for {
			select {
			case event := <-w.Event:
				rs, ls := f.checkLogFile(event)
				for _, r := range rs {
					log.Println("rep")
					reps <- r
				}
				for _, l := range ls {
					log.Println("loc")
					locs <- l
				}

			case err := <-w.Error:
				errs <- err
			case <-w.Closed:
				return
			case <-ctx.Done():
				w.Close()
			}
		}
	}()

	if err := w.Add(f.chatlogDir); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(500 * time.Millisecond); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (f *LogFeed) checkLogFile(fileInfo os.FileInfo) (reps []Report, locs []Locstat) {

	log.Println("DEBUG: LW: Event 1")

	lf, err := os.Open(filepath.Join(f.chatlogDir, fileInfo.Name()))
	if err != nil {
		return nil, nil
	}
	defer lf.Close()
	lfe := transform.NewReader(lf, codec.NewDecoder())
	blf := bufio.NewScanner(lfe)

	isLocal := false
	chanName := ""
	listener := ""
	var skip uint64 = 12

	var line uint64 = 0
	for blf.Scan() {
		line++
		text := blf.Text()

		// Now remove the weird BOM like thing (WTF is this CCP?)
		if len(text) > 3 {
			text = text[3:]
		}

		switch {
		case line == 7:
			//	This is the channel id, use it to check local channel regardless of lang
			isLocal = strings.TrimSpace(strings.Split(text, ":")[1]) == "local"
			if isLocal {
				log.Println("DEBUG: LW: Event LOCAL")
			}
		case line == 8:
			//	This is the channel name
			chanName = strings.TrimSpace(strings.Split(text, ":")[1])
			log.Printf("DEBUG: LW: Event - %s", chanName)
			if lim, ok := f.loglines[chanName]; ok {
				skip = lim
			} else {
				if isLocal {
					// if local we do care about the first log line
					skip = 12
				} else {
					// if not local we dont want to capture the MOTD
					skip = 13
				}
			}
			if !isLocal {
				//	Now check if the channel is one we care about, if not then just return early :)
				found := false
				for _, room := range f.roomnames {
					log.Printf("DEBUG LW: Checking %s = %s => %v", chanName, room, chanName == room)
					if chanName == room {
						found = true
						break
					}
				}
				log.Printf("DEBUG: LW: Found - %v", found)
				if !found {
					return nil, nil
				}
			}

		case line == 9:
			//	This is the listener line
			listener = strings.TrimSpace(strings.Split(text, ":")[1])
			log.Printf("DEBUG: LW: Listener - %v", listener)
		case line >= skip:
			//	This is a line we want to report
			if len(text) < 24 {
				// Happens occasionally
				log.Println("Short Line")
				break
			}
			if isLocal {
				loc := f.parseLocalMessage(text)
				if loc.System != "" {
					loc.Character = listener
					locs = append(locs, loc)
				}
			} else {
				//	Dealing with an intel room
				log.Println("DEBUG: LW: Intel message")
				rep := f.parseIntelMessage(text)
				if rep != (Report{}) {
					rep.Listener = listener
					rep.Source = fmt.Sprintf("log: %s", chanName)
					log.Printf("DEBUG: LW: Making Report - %#v", rep)
					reps = append(reps, rep)
				}
			}
		}
	}

	f.loglines[chanName] = line + 1

	log.Printf("DEBUG: LW: Making %d intel reports and %d location reports", len(reps), len(locs))

	return reps, locs
}

func (f *LogFeed) parseLocalMessage(msg string) (loc Locstat) {
	t, sender, msgp, err := f.splitLogMessage(msg)
	if err != nil {
		log.Println(fmt.Errorf("failed to decode timestamp of local log message; '%s': %w", msg, err))
		return Locstat{}
	}

	if sender != "EVE System" {
		//	Not a location update, return an empty message
		return Locstat{}
	}

	// Character will get populated by parent method
	return Locstat{
		System:    strings.TrimSpace(strings.Split(msgp, ":")[1]),
		Time:      t,
		Character: "",
	}
}

func (f *LogFeed) parseIntelMessage(msg string) (rep Report) {
	t, sender, msgp, err := f.splitLogMessage(msg)

	if err != nil {
		log.Println(fmt.Errorf("failed to decode timestamp of intel log message; '%s': %w", msg, err))
		return Report{}
	}

	rep = Report{
		Message:  msgp,
		Reporter: sender,
		Time:     t,
	}

	return rep
}

func (f *LogFeed) splitLogMessage(msg string) (t time.Time, sender string, message string, err error) {
	dts := msg[2:21]
	tme, err := time.Parse(logTimeFormat, dts)
	if err != nil {
		return time.Time{}, "", "", err
	}
	s := strings.Split(msg[24:], ">")
	return tme, strings.TrimSpace(s[0]), strings.TrimSpace(s[1]), nil
}
