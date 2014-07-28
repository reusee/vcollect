package main

import (
	"os"

	pyqt "github.com/reusee/go-pyqt5"
)

func (db *Db) qt(infos []*PathInfo) {
	// ui
	qt, err := pyqt.New(`
from PyQt5.QtWidgets import QGraphicsView, QGraphicsScene
class Win(QGraphicsView):
	def __init__(self, **kwds):
		self.scene = QGraphicsScene()
		super().__init__(self.scene, **kwds)
	def keyPressEvent(self, ev):
		Emit("key", ev.key())
win = Win(styleSheet = 'background-color: black;')

from PyQt5.QtMultimedia import QMediaPlayer, QMediaContent
from PyQt5.QtMultimediaWidgets import QGraphicsVideoItem
from PyQt5.QtCore import QUrl
player = QMediaPlayer()
video = QGraphicsVideoItem()
player.setVideoOutput(video)
player.setNotifyInterval(100)
win.scene.addItem(video)
Connect("play", lambda path: [
	player.setMedia(QMediaContent(QUrl.fromLocalFile(path))),
	player.play(),
])

win.showMaximized()
	`)
	if err != nil {
		panic(err)
	}
	qt.OnClose(func() {
		os.Exit(0)
	})

	keys := make(chan rune)
	qt.Connect("key", func(key float64) {
		select {
		case keys <- rune(key):
		default:
		}
	})

	index := 0
	qt.Emit("play", infos[index].path)

	for {
		key := <-keys
		_ = key
	}

}
