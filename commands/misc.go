package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/yz3358/clash-ctl/common"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/text"
)

func HandleMiscCommand(args []string) {
	if len(args) == 0 {
		return
	}

	cfg, err := common.ReadCfg()
	if err != nil {
		return
	}

	switch args[0] {
	case "now":
		current, server, err := common.GetCurrentServer(cfg)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		serverURL := server.URL()
		fmt.Printf("now selected %s - %s\n", current, serverURL.String())
	case "use":
		if len(args) < 2 {
			fmt.Println("should input server name")
			return
		}

		name := args[1]
		if _, ok := cfg.Servers[name]; !ok {
			fmt.Printf("serber %s not found\n", name)
			return
		}

		cfg.Selected = name
		if err := common.SaveCfg(cfg); err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("now use %s\n", text.FgGreen.Sprint(name))
	case "ping":
		servers := cfg.Servers

		pw := progress.NewWriter()
		pw.SetAutoStop(true)
		pw.SetMessageWidth(20)
		pw.SetSortBy(progress.SortByNone)
		pw.SetNumTrackersExpected(len(servers))
		style := progress.StyleVisibilityDefault
		style.Time = false
		style.Tracker = false
		style.Percentage = false
		pw.SetStyle(progress.Style{
			Name:       "none",
			Chars:      progress.StyleChars{},
			Colors:     progress.StyleColorsExample,
			Visibility: style,
			Options:    progress.StyleOptions{},
		})
		pw.SetTrackerPosition(progress.PositionRight)
		pw.SetUpdateFrequency(time.Millisecond * 10)

		wg := sync.WaitGroup{}
		for name, server := range servers {
			wg.Add(1)
			go trackPing(&wg, pw, name, server)
		}

		pw.Render()
		wg.Wait()
	}
}

func trackPing(wg *sync.WaitGroup, pw progress.Writer, name string, server common.Server) {
	defer wg.Done()

	tracker := progress.Tracker{
		Message: name, Total: 1,
		Units: progress.Units{
			Notation: "",
			Formatter: func(value int64) string {
				// 0 loading
				// 1 success
				// 2 error
				switch value {
				case 0:
					return "loading"
				case 1:
					return text.FgGreen.Sprint("success")
				case 2:
					return text.FgRed.Sprint("error")
				default:
					return text.FgRed.Sprint("unknown")
				}
			},
		},
	}
	defer tracker.MarkAsDone()

	pw.AppendTracker(&tracker)
	req := common.MakeRequest(server).SetTimeout(3 * time.Second)

	resp, err := req.R().Get("/version")
	if err != nil {
		tracker.SetValue(2)
		return
	}

	if resp.StatusCode() != 200 {
		tracker.SetValue(2)
		return
	}

	time.Sleep(time.Millisecond * 100)
	tracker.SetValue(1)
}
