/*
 * 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
 * 2014+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
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

/*
#include "session.h"
#include <stdio.h>
*/
import "C"

type IteratorResult interface {
	ReplyData() []byte
	Id() uint64
	Error() error
}

type iteratorResult struct {
	replyData []byte
	id        uint64
	err       error
}

func (i *iteratorResult) Id() uint64 {
	return i.id
}

func (i *iteratorResult) ReplyData() []byte {
	return i.replyData
}

func (i *iteratorResult) Error() error {
	return i.err
}

func iteratorHelper(key string, iteratorId uint64) (*Key, uint64, uint64, <-chan IteratorResult) {
	ekey, err := NewKey(key)
	if err != nil {
		panic(err)
	}

	responseCh := make(chan IteratorResult, defaultVOLUME)

	onResultContext := NextContext()
	onFinishContext := NextContext()

	onResult := func(iterres *iteratorResult) {
		responseCh <- iterres
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- &iteratorResult{err: err}
		}
		close(responseCh)

		Pool.Delete(onResultContext)
		Pool.Delete(onFinishContext)
	}

	Pool.Store(onResultContext, onResult)
	Pool.Store(onFinishContext, onFinish)
	return ekey, onResultContext, onFinishContext, responseCh
}

func (s *Session) IteratorPause(key string, iteratorId uint64) <-chan IteratorResult {
	ekey, onResultContext, onFinishContext, responseCh := iteratorHelper(key, iteratorId)
	defer ekey.Free()

	C.session_pause_iterator(s.session, C.context_t(onResultContext), C.context_t(onFinishContext),
		ekey.key,
		C.uint64_t(iteratorId))
	return responseCh
}

func (s *Session) IteratorContinue(key string, iteratorId uint64) <-chan IteratorResult {
	ekey, onResultContext, onFinishContext, responseCh := iteratorHelper(key, iteratorId)
	defer ekey.Free()

	C.session_continue_iterator(s.session, C.context_t(onResultContext), C.context_t(onFinishContext),
		ekey.key,
		C.uint64_t(iteratorId))
	return responseCh
}

func (s *Session) IteratorStop(key string, iteratorId uint64) <-chan IteratorResult {
	ekey, onResultContext, onFinishContext, responseCh := iteratorHelper(key, iteratorId)
	defer ekey.Free()

	C.session_stop_iterator(s.session, C.context_t(onResultContext), C.context_t(onFinishContext),
		ekey.key,
		C.uint64_t(iteratorId))
	return responseCh
}
