package midi

import (
	"errors"
	"time"
)

var ErrNotImplemeneted = errors.New("not implemented")

type NopDriver struct{}

func (d *NopDriver) Close() error                          { return nil }
func (d *NopDriver) SetReceiveHandler(ReceiveEventHandler) {}
func (d *NopDriver) OpenInput(Entity) error                { return ErrNotImplemeneted }
func (d *NopDriver) OpenOutput(Entity) error               { return ErrNotImplemeneted }
func (d *NopDriver) Send(time.Time, Entity, []Word) error  { return ErrNotImplemeneted }
func (d *NopDriver) SendSysEx(Entity, []Word) error        { return ErrNotImplemeneted }
func (d *NopDriver) SendSysExV1(Entity, []byte) error      { return ErrNotImplemeneted }
func (d *NopDriver) Enumerate() (*Node, error)             { return nil, ErrNotImplemeneted }
