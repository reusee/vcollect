package main

import (
	"net"
	"os"

	"github.com/reusee/lgo"
)

func (db *Db) lgi_gst(infos []*PathInfo) {
	// lua vm
	lua := lgo.NewLua()

	// exit function
	lua.RegisterFunction("Exit", func() {
		os.Exit(0)
	})

	// retrive chan
	values := make(chan interface{})
	lua.RegisterFunction("Return", func(i interface{}) {
		values <- i
	})

	// get video path
	index := 0
	lua.RegisterFunction("GetPath", func() string {
		return infos[index].path
	})

	// key press
	keys := make(chan rune)
	lua.RegisterFunction("Key", func(val rune) {
		select {
		case keys <- val:
		default:
		}
	})

	// code eval notification
	ln, err := net.Listen("tcp", "127.0.0.1:38912")
	if err != nil {
		panic(err)
	}
	connReady := make(chan bool)
	var conn net.Conn
	go func() {
		conn, err = ln.Accept()
		if err != nil {
			panic(err)
		}
		close(connReady)
	}()
	codeToRun := make(chan string, 4)
	lua.RegisterFunction("Execute", func() {
		lua.RunString(<-codeToRun)
	})
	run := func(code string) {
		codeToRun <- code
		conn.Write([]byte{'g'})
	}

	go lua.RunString(`
lgi = require('lgi')
Gtk = lgi.require('Gtk', '3.0')
Gst = lgi.require('Gst', '1.0')
GstVideo = lgi.require('GstVideo', '1.0')
GdkX11 = lgi.GdkX11
Gio = lgi.Gio
GLib = lgi.GLib

win = Gtk.Window{
	Gtk.Grid{
		expand = true,
		orientation = 'VERTICAL',
		Gtk.DrawingArea{
			id = 'output',
			expand = true,
		},
		Gtk.Label{
			id = 'uri',
		},
	},
}
function win.on_destroy() Exit() end

function win:on_key_press_event(event)
	Key(event.keyval)
end

pipeline = Gst.ElementFactory.make('playbin', 'bin')
sink = Gst.ElementFactory.make('ximagesink', 'sink')
pipeline.video_sink = sink

xid = true
function load_video()
	pipeline.state = 'NULL'
	sink:set_window_handle(xid)
	path = GetPath()
	pipeline.uri = 'file://' .. path
	pipeline.state = 'PLAYING'
	win.child.uri.label = path
end
function win.child.output:on_realize()
	xid = self.window:get_xid()
	load_video()
end

function seek(position, duration)
	if position > duration then
		position = duration
	end
	if position < 0 then
		position = 0
	end
	pipeline:seek_simple(Gst.Format.TIME, {'FLUSH', 'KEY_UNIT', 'SNAP_AFTER'}, position)
end
function seek_time(n)
	position = pipeline:query_position('TIME')
	if position == nil then return end
	position = position + n
	duration = pipeline:query_duration('TIME')
	seek(position, duration)
end
function seek_percent(n)
	position = pipeline:query_position('TIME')
	if position == nil then return end
	duration = pipeline:query_duration('TIME')
	position = position + duration / 100 * n
	seek(position, duration)
end

socket = Gio.Socket.new(Gio.SocketFamily.IPV4, Gio.SocketType.STREAM, Gio.SocketProtocol.TCP)
socket:connect(Gio.InetSocketAddress.new_from_string("127.0.0.1", 38912))
channel = GLib.IOChannel.unix_new(socket.fd)
bytes = require('bytes')
buf = bytes.new(1)
GLib.io_add_watch(channel, GLib.PRIORITY_DEFAULT, GLib.IOCondition.IN, function()
	Execute()
	socket:receive(buf)
	return true
end)

pipeline.bus:add_watch(GLib.PRIORITY_DEFAULT, function(bus, message)
end)

win:show_all()
Gtk.main()

pipeline.state = 'NULL'
	`)

	// wait lua
	<-connReady
	p("connected.\n")

	for {
		key := <-keys
		switch key {
		case 'q':
			os.Exit(0)

		case 'j':
			// next video
			index += 1
			if index >= len(infos) {
				index = 0
			}
			run("load_video()")
		case 'k':
			// prev video
			index -= 1
			if index < 0 {
				index = len(infos) - 1
			}
			run("load_video()")

		case 's':
			// seek forward
			run("seek_time(3000000000)")
		case 'w':
			// seek backward
			run("seek_time(-3000000000)")
		case 'd':
			// seek forward long
			run("seek_time(10000000000)")
		case 'a':
			// seek backward long
			run("seek_time(-10000000000)")
		case 'S':
			// seek percent forward
			run("seek_percent(3)")
		case 'W':
			// seek percent backward
			run("seek_percent(-3)")
		case 'D':
			// seek percent forward long
			run("seek_percent(10)")
		case 'A':
			// seek percent backward long
			run("seek_percent(-10)")

		case 'e':
			// tag
			run(`
			local pos = pipeline:query_position('TIME')
			Return(pos)
			`)
			pos := int64((<-values).(float64))
			p("%d\n", pos)

		default:
			p("%d\n", key)
		}
	}

}
