package main

import (
	"os"

	"github.com/reusee/lgtk"
)

func (db *Db) lgi_gst(infos []*PathInfo) {
	index := 0
	keys := make(chan rune)

	g, err := lgtk.New(`
Gst = lgi.require('Gst', '1.0')
GstVideo = lgi.require('GstVideo', '1.0')
GdkX11 = lgi.GdkX11

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
	`,
		"GetPath", func() string {
			return infos[index].path
		},
		"Key", func(val rune) {
			select {
			case keys <- val:
			default:
			}
		},
	)
	if err != nil {
		panic(err)
	}

	// helper functions
	getPos := func() int64 {
		g.Exec(`
			local pos = pipeline:query_position('TIME')
			Return(pos)
			`)
		return int64((<-g.Return).(float64))
	}
	getInput := func() string {
		g.Exec(`
		pipeline.state = 'PAUSED'
		input:show()
		input:grab_focus()
		`)
		return (<-g.Return).(string)
	}

	for {
		key := <-keys
		switch key {
		case 'q':
			os.Exit(0)

		case ' ':
			// toggle pause
			g.Exec("toggle_pause()")

		case 'j', 'r':
			// next video
			index += 1
			if index >= len(infos) {
				index = 0
			}
			g.Exec("load_video()")
		case 'k', 'z':
			// prev video
			index -= 1
			if index < 0 {
				index = len(infos) - 1
			}
			g.Exec("load_video()")

		case 'd':
			// seek forward
			g.Exec("seek_time(3000000000)")
		case 'a':
			// seek backward
			g.Exec("seek_time(-3000000000)")
		case 's':
			// seek forward long
			g.Exec("seek_time(10000000000)")
		case 'w':
			// seek backward long
			g.Exec("seek_time(-10000000000)")
		case 'D':
			// seek percent forward
			g.Exec("seek_percent(3)")
		case 'A':
			// seek percent backward
			g.Exec("seek_percent(-3)")
		case 'S':
			// seek percent forward long
			g.Exec("seek_percent(10)")
		case 'W':
			// seek percent backward long
			g.Exec("seek_percent(-10)")

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
				g.Exec(s("seek_abs(%d)", next))
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
				g.Exec(s("seek_abs(%d)", prev))
			}
		case 'x':
			// to first tag
			if len(infos[index].file.Tags) > 0 {
				g.Exec(s("seek_abs(%d)", infos[index].file.Tags[0].Position))
			}

		}
	}

}
