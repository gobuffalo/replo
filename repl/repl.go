package repl

import (
	"fmt"
	"io"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
)

func (s *Session) Start() error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return errors.WithStack(err)
	}

	defer g.Close()
	t := &terminal{
		Gui:     g,
		Session: s,
	}
	return t.start()
}

type terminal struct {
	*gocui.Gui
	Session    *Session
	mainView   *gocui.View
	outputView *gocui.View
	debugView  *gocui.View
}

func (t *terminal) start() error {
	g := t.Gui
	g.Cursor = true
	g.SetManagerFunc(t.layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, t.quit); err != nil {
		return errors.WithStack(err)
	}

	if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, t.saveCode); err != nil {
		return errors.WithStack(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return errors.WithStack(err)
	}
	return nil
}

func (t *terminal) saveCode(g *gocui.Gui, v *gocui.View) error {
	t.outputView.Clear()
	if t.debugView != nil {
		t.debugView.Clear()
	}
	b, err := t.Session.Execute(t.mainView.Buffer(), t.debugView)
	if err != nil {
		t.outputView.Write([]byte(err.Error()))
	}
	t.outputView.Write(b)
	return nil
}

func (t *terminal) quit(g *gocui.Gui, v *gocui.View) error {
	if !t.Session.SkipHistory {
		if f, err := os.Create(t.Session.History); err == nil {
			io.Copy(f, t.mainView)
			defer f.Close()
		}
	}
	return gocui.ErrQuit
}

func (t *terminal) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	var err error
	if t.mainView, err = g.SetView("main", 0, 0, maxX/2-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		t.mainView.Title = "Console"
		t.mainView.Editable = true
		t.mainView.Autoscroll = true
		t.mainView.Wrap = true

		if f, err := os.Open(t.Session.History); err == nil && !t.Session.SkipHistory {
			io.Copy(t.mainView, f)
			f.Close()
		} else {
			for _, i := range t.Session.originalImports {
				fmt.Fprintln(t.mainView, fmt.Sprintf("import \"%s\"", i))
			}
		}

		t.mainView.SetCursor(0, len(t.Session.originalImports))

		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	outX := maxX - 1
	outY := maxY - 1
	if t.Session.Debug {
		outY = maxY / 2

		if t.debugView, err = g.SetView("debug", maxX/2-1, outY+1, outX, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			t.debugView.Autoscroll = true
			t.debugView.Title = "Debug"
			t.debugView.Wrap = true
		}
	}
	if t.outputView, err = g.SetView("output", maxX/2-1, 0, outX, outY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		t.outputView.Autoscroll = true
		t.outputView.Title = "Output"
		t.outputView.Wrap = true
	}
	return nil
}
