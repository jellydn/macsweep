package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
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
		mygo.NewWindow(mygo.WindowOptions{
			Title:     "MacSweep",
			Width:     640,
			Height:    400,
			MinWidth:  520,
			MinHeight: 360,
			Page:      mygo.PageOptions{DevTools: mygo.DevToolsDisabled},
			Content:   ui.View(appui.View),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
