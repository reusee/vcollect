package main

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/reusee/lgo"
)

func (db *Db) lgi_mplayer(infos []*PathInfo) {
	lua := lgo.NewLua()

	lua.RegisterFunction("Exit", func() {
		os.Exit(0)
	})

	getXid := make(chan int)
	lua.RegisterFunction("ProvideXid", func(i int) {
		getXid <- i
	})

	go lua.RunString(`
lgi = require('lgi')
Gtk = lgi.require('Gtk', '3.0')
GdkX11 = lgi.GdkX11
win = Gtk.Window{
	Gtk.Grid{
		Gtk.DrawingArea{
			id = 'output',
			expand = true,
		},
		expand = true,
	},
}
function win.on_destroy() Exit() end
function win.child.output:on_realize()
	ProvideXid(self.window:get_xid())
end

win:show_all()
Gtk.main()
	`)

	xid := <-getXid
	p("%d\n", xid)
	output, err := exec.Command("mplayer", "-wid", strconv.Itoa(xid), infos[0].path).Output()
	p("%s\n", output)
	if err != nil {
		panic(err)
	}

}
