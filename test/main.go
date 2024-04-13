package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jaz303/midi"
	_ "github.com/jaz303/midi/alsa"
	"github.com/jaz303/midi/ump"
)

type event struct {
	Time   time.Time
	Entity midi.Entity
	Words  []ump.Word
}

var events = make(chan event)

func onEvent(time time.Time, entity midi.Entity, words []ump.Word) {
	out := make([]ump.Word, len(words))
	copy(out, words)

	events <- event{
		Time:   time,
		Entity: entity,
		Words:  out,
	}
}

func main() {
	d, err := midi.NewDriverByName("alsa", "devtest")
	if err != nil {
		panic(fmt.Errorf("create driver failed: %s", err))
	}

	d.SetReceiveHandler(onEvent)

	ports, err := d.Enumerate()
	if err != nil {
		log.Fatal(err)
	}

	inputs := ports.Collect(func(n *midi.Node) bool {
		return n.Type == midi.Input && strings.Contains(n.Name, "X2mini")
	})

	if len(inputs) == 0 {
		log.Fatalf("no input found")
	}

	if err := d.OpenInput(inputs[0].Entity); err != nil {
		log.Fatalf("failed to open input: %s", err)
	}

	//time.Sleep(5 * time.Second)
	//d.Close()

	for {
		select {
		case evt := <-events:
			fmt.Printf("%v %04X %08X\n", evt.Time, evt.Entity, evt.Words[0])
		}
	}

	// for _, i := range ports.Inputs() {
	// 	if i.Name == "SH-4d" {
	// 		log.Printf("OPEN: %d %s", i.Entity, i)
	// 		d.OpenInput(i.Entity)
	// 	}
	// }

	// var op midi.Entity
	// for _, i := range ports.Outputs() {
	// 	if i.Name == "SH-4d" {
	// 		log.Printf("OPEN: %d %s", i.Entity, i)
	// 		d.OpenOutput(i.Entity)
	// 		op = i.Entity
	// 	}
	// }

	// ticker1 := time.NewTicker(500 * time.Millisecond)
	// for {
	// 	select {
	// 	case <-ticker1.C:
	// 		now := time.Now()
	// 		d.Send(now, op, []midi.Word{
	// 			midi1.NoteOn(1, 64, 100),
	// 			midi1.NoteOn(1, 67, 100),
	// 			midi1.NoteOn(1, 69, 100),
	// 		})
	// 		d.Send(now.Add(100*time.Millisecond), op, []midi.Word{
	// 			midi1.NoteOff(1, 64, 100),
	// 			midi1.NoteOff(1, 67, 100),
	// 			midi1.NoteOff(1, 69, 100),
	// 		})
	// 	case evt := <-events:
	// 		log.Printf("%v %d %d", evt.Time, evt.Entity, evt.Words[0])
	// 	}
	// }
}
