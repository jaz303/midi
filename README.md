# midi

Low-level MIDI library for Go. Very much alpha for now.

Supported platforms: Mac OS (Core MIDI)

## Features

  - Simple API
  - Rock-solid timing; uses underyling driver's best available scheduling method - no `time.Sleep()` here
  - Device enumeration
  - Schedule outgoing events

## Basic Usage

```go
func main() {
    driver := &darwin.Driver

    driver.Init(&DriverConfig{
        ReceiveHandler: func(time time.Time, entity Entity, data []Word) {
            // Handle events received on open input ports here
            // Note 1: this function is probably called by the OS in a high-priority thread
            // Note 2: data must not be altered/retained - copy if needed outside this function
        }
    })

    // Print all ports
    ports, _ := driver.Enumerate()
    ports.Print()

    // Open all inputs
    for _, p := range ports.Inputs() {
        driver.OpenInput(p)
    }

    // Open the first output
    output := ports.Outputs[0]
    driver.OpenOutput(p)
    
    // Send note-on/note-off
    driver.Send(time.Now(), output, []Word{
        midi1.NoteOn(1, 64, 100),
    })
    
    driver.Send(time.Now().Add(500 * time.Millisecond), output, []Word{
        midi1.NoteOff(1, 64, 0)
    })
}
```
