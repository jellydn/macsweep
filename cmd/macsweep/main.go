package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/jellydn/macsweep/internal/platform"
	"github.com/jellydn/macsweep/internal/policy"
	"github.com/jellydn/macsweep/internal/scan"
	"github.com/jellydn/macsweep/internal/store"
	appui "github.com/jellydn/macsweep/internal/ui"
)

func main() {
	version := flag.Bool("version", false, "print the build version without starting the app")
	flag.Parse()
	if *version {
		v := mygo.App.Version()
		if v == "" {
			v = "development"
		}
		fmt.Println("MacSweep", v)
		return
	}
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "MacSweep does not accept file paths or cleanup commands.")
		os.Exit(2)
	}
	if runtime.GOOS != "darwin" {
		fmt.Fprintln(os.Stderr, "MacSweep requires macOS. Headless view tests can run on other systems.")
		os.Exit(1)
	}
	mygo.App.WhenReady(func() {
		model := &appui.Model{}
		home, err := os.UserHomeDir()
		var p *policy.Policy
		if err == nil {
			p, err = policy.Default(home)
		}
		if err != nil {
			model.PolicyError = err.Error()
		}
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "MacSweep",
			Width:     640,
			Height:    480,
			MinWidth:  520,
			MinHeight: 360,
			Page:      mygo.PageOptions{DevTools: mygo.DevToolsDisabled},
			Content:   ui.View(model.View),
		})
		if p == nil {
			return
		}
		basePolicy := p
		model.PreferencesBusy = true
		go func() {
			paths, err := store.LoadExclusions(home)
			var next *policy.Policy
			if err == nil {
				next, err = basePolicy.WithExclusions(paths)
			}
			win.Update(func() {
				if err == nil {
					p = next
				}
				model.ExclusionsLoaded(paths, err)
			})
		}()
		model.OnSaveExclusions = func(paths []string) {
			go func() {
				next, err := basePolicy.WithExclusions(paths)
				if err == nil {
					err = store.SaveExclusions(home, paths)
				}
				win.Update(func() {
					if err == nil {
						p = next
					}
					model.ExclusionsLoaded(paths, err)
				})
			}()
		}
		// Capacity and scanning run off the UI thread. No content is read or moved.
		go func() {
			capacity, err := platform.Volume(basePolicy.Home())
			win.Update(func() {
				model.Capacity = capacity
				if err != nil {
					model.CapacityError = err.Error()
				}
			})
		}()
		var cancel context.CancelFunc
		model.OnScan = func() {
			ctx, stop := context.WithCancel(context.Background())
			cancel = stop
			model.Begin()
			scanPolicy := p
			go func() {
				defer stop()
				snapshot, err := scan.RunWithProgress(ctx, scanPolicy, func(progress scan.Progress) {
					win.Update(func() { model.Progress(progress) })
				})
				win.Update(func() { model.Finish(snapshot, err) })
			}()
		}
		model.OnCancel = func() {
			if cancel != nil {
				cancel()
			}
		}
		win.OnClosed(func() {
			if cancel != nil {
				cancel()
			}
		})
		win.Invalidate()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
