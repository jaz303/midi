package main

import (
	"log"
	"time"

	"github.com/jaz303/midi"
	_ "github.com/jaz303/midi/darwin"
	"github.com/jaz303/midi/midi1"
)

type event struct {
	Time   time.Time
	Entity midi.Entity
	Words  []midi.Word
}

var events = make(chan event)

func onEvent(time time.Time, entity midi.Entity, words []midi.Word) {
	out := make([]midi.Word, len(words))
	copy(out, words)

	events <- event{
		Time:   time,
		Entity: entity,
		Words:  out,
	}
}

func main() {
	d := midi.DriverByName("Core MIDI")
	d.Init(&midi.DriverConfig{
		ReceiveHandler: onEvent,
	})

	ports, err := d.Enumerate()
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range ports.Inputs() {
		if i.Name == "SH-4d" {
			log.Printf("OPEN: %d %s", i.Entity, i)
			d.OpenInput(i.Entity)
		}
	}

	var op midi.Entity
	for _, i := range ports.Outputs() {
		if i.Name == "SH-4d" {
			log.Printf("OPEN: %d %s", i.Entity, i)
			d.OpenOutput(i.Entity)
			op = i.Entity
		}
	}

	ticker1 := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker1.C:
			now := time.Now()
			d.Send(now, op, []midi.Word{
				midi1.NoteOn(1, 64, 100),
				midi1.NoteOn(1, 67, 100),
				midi1.NoteOn(1, 69, 100),
			})
			d.Send(now.Add(100*time.Millisecond), op, []midi.Word{
				midi1.NoteOff(1, 64, 100),
				midi1.NoteOff(1, 67, 100),
				midi1.NoteOff(1, 69, 100),
			})
		case evt := <-events:
			log.Printf("%v %d %d", evt.Time, evt.Entity, evt.Words[0])
		}
	}
}
