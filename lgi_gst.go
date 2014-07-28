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
			can_focus = true,
		},
		Gtk.Label{
			id = 'uri',
		},
		Gtk.Entry{
			id = 'input',
			visible = false,
		},
	},
}
function win.on_destroy() Exit() end
function win.on_realize()
	win.child.input:hide()
end

function win.child.output:on_key_press_event(event)
	Key(event.keyval)
	return true
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

function seek(position, duration, flag)
	if position > duration then
		return
	end
	if position < 0 then
		position = 0
	end
	pipeline:seek_simple(Gst.Format.TIME, {'FLUSH', 'KEY_UNIT', flag}, position)
end
function seek_abs(position)
	pipeline:seek_simple(Gst.Format.TIME, {'FLUSH', 'ACCURATE'}, position)
end
function seek_time(n)
	position = pipeline:query_position('TIME')
	if position == nil then return end
	position = position + n
	duration = pipeline:query_duration('TIME')
	flag = 'SNAP_AFTER'
	if n < 0 then flag = 'SNAP_BEFORE' end
	seek(position, duration, flag)
end
function seek_percent(n)
	position = pipeline:query_position('TIME')
	if position == nil then return end
	duration = pipeline:query_duration('TIME')
	position = position + duration / 100 * n
	flag = 'SNAP_AFTER'
	if n < 0 then flag = 'SNAP_BEFORE' end
	seek(position, duration, flag)
end

paused = false
function toggle_pause()
	if paused then
		pipeline.state = 'PLAYING'
		paused = false
	else
		pipeline.state = 'PAUSED'
		paused = true
	end
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
	if message.type.ERROR then
		print('Error:', message:parse_error().message)
	end
	if message.type.EOS then
		print('end of stream')
		seek_abs(0)
		pipeline.state = 'PLAYING'
	end
	return true
end)

input = win.child.input
input.on_activate:connect(function()
	input:hide()
	win.child.output:grab_focus()
	Return(input:get_text())
	pipeline.state = 'PLAYING'
	return true
end)

win:show_all()
Gtk.main()

pipeline.state = 'NULL'
	`)

	// wait lua
	<-connReady
	p("connected.\n")

	// helper functions
	getPos := func() int64 {
		run(`
			local pos = pipeline:query_position('TIME')
			Return(pos)
			`)
		return int64((<-values).(float64))
	}
	getInput := func() string {
		run(`
		pipeline.state = 'PAUSED'
		input:show()
		input:grab_focus()
		`)
		return (<-values).(string)
	}

	for {
		key := <-keys
		switch key {
		case 'q':
			os.Exit(0)

		case ' ':
			// toggle pause
			run("toggle_pause()")

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

		case 'd':
			// seek forward
			run("seek_time(3000000000)")
		case 'a':
			// seek backward
			run("seek_time(-3000000000)")
		case 's':
			// seek forward long
			run("seek_time(10000000000)")
		case 'w':
			// seek backward long
			run("seek_time(-10000000000)")
		case 'D':
			// seek percent forward
			run("seek_percent(3)")
		case 'A':
			// seek percent backward
			run("seek_percent(-3)")
		case 'S':
			// seek percent forward long
			run("seek_percent(10)")
		case 'W':
			// seek percent backward long
			run("seek_percent(-10)")

		case 'e':
			// tag
			pos := getPos()
			desc := getInput()
			if desc != "" {
				p("add tag %d %s\n", pos, desc)
				infos[index].file.AddTag(pos, desc)
				db.Save()
			}
		case 'f':
			// next tag
			pos := getPos()
			var next int64
			for _, tag := range infos[index].file.Tags {
				if tag.Position > pos {
					if tag.Position < next || next == 0 {
						next = tag.Position
					}
				}
			}
			if next > 0 {
				p("jump to tag %d\n", next)
				run(s("seek_abs(%d)", next))
			}
		case 'c':
			// prev tag
			pos := getPos()
			var prev int64
			for _, tag := range infos[index].file.Tags {
				if tag.Position < pos {
					if tag.Position > prev || prev == 0 {
						prev = tag.Position
					}
				}
			}
			if prev > 0 {
				p("jump to tag %d\n", prev)
				run(s("seek_abs(%d)", prev))
			}
		case 'x':
			// to first tag
			if len(infos[index].file.Tags) > 0 {
				run(s("seek_abs(%d)", infos[index].file.Tags[0].Position))
			}

		}
	}

}
