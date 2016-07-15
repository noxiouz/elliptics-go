/*
 * 2016+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
 * All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 */

package elliptics

type DChannel struct {
	In		chan interface{}
	Out		chan interface{}
	buffer		[]interface{}
}

func NewDChannel() *DChannel {
	dch := &DChannel {
		In:		make(chan interface{}, defaultVOLUME),
		Out:		make(chan interface{}, defaultVOLUME),
		buffer:		make([]interface{}, 0),
	}

	// schedule reader
	go dch.run()
	return dch
}

func (dch *DChannel) run() {
	// close output channel when exiting from this infinite processing loop
	// exit can only happen if input channel has been closed
	defer close(dch.Out)
recv:
	for {
		if len(dch.buffer) > 0 {
			select {
			case dch.Out <- dch.buffer[0]:
				dch.buffer = dch.buffer[1:]

			case v, ok := <-dch.In:
				if !ok {
					break recv
				}

				dch.buffer = append(dch.buffer, v)
			}
		} else {
			v, ok := <-dch.In
			if !ok {
				break recv
			}

			dch.buffer = append(dch.buffer, v)
		}
	}

	// if there is something we haven't yet pushed to the output channel
	for _, v := range dch.buffer {
		dch.Out <- v
	}
}
