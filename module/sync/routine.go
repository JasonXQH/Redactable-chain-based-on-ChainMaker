/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"fmt"
	"sync/atomic"

	commonErrors "chainmaker.org/chainmaker/common/v2/errors"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/Workiva/go-datastructures/queue"
)

type (
	getServiceState func() string
	handleFunc      func(event queue.Item) (queue.Item, error)
)

//const bufferSize = 1024
const bufferSize = 128

// Routine Provide hosting of the service in goroutine
type Routine struct {
	// The name of the hosted service
	name string
	// Processing of the hosted service
	handle handleFunc
	// get state in the service
	queryState getServiceState
	log        protocol.Logger
	// The flag which detects whether the service is started
	start int32
	// A queue to store tasks
	queue *queue.PriorityQueue
	// Outputs the execution results of the task
	out chan queue.Item
	// Notify whether the service has stopped
	stop chan struct{}
}

//NewRoutine create a Routine instance
func NewRoutine(name string, handle handleFunc, queryState getServiceState, log protocol.Logger) *Routine {
	return &Routine{
		name:       name,
		handle:     handle,
		queryState: queryState,
		log:        log,

		queue: queue.NewPriorityQueue(bufferSize, true),
		out:   make(chan queue.Item),
		stop:  make(chan struct{}),
	}
}

//begin start a background coroutine to loop to get event
func (r *Routine) begin() error {
	if !atomic.CompareAndSwapInt32(&r.start, 0, 1) {
		return commonErrors.ErrSyncServiceHasStarted
	}
	go r.loop()
	return nil
}

//loop to get those items that are placed in the queue sorted by priority
//and pass item to the registered hander for processing
//if the handler returns a non-null event item, put it to the out channel
//because there may be other handlers that need to get this event item to handle
func (r *Routine) loop() {
	for {
		items, err := r.queue.Get(1)
		if err != nil && err != queue.ErrDisposed {
			panic(fmt.Sprintf("retrieves item from queue failed, reason: %s", err.Error()))
		}
		var ret queue.Item
		if len(items) > 0 {
			if ret, err = r.handle(items[0]); err != nil {
				r.log.Errorf("process msg failed, reason: %s", err.Error())
			}
		}
		select {
		case <-r.stop:
			return
		default:
			if ret != nil {
				r.out <- ret
			}
		}
	}
}

func (r *Routine) addTask(event queue.Item) error {
	if atomic.LoadInt32(&r.start) != 1 {
		r.log.Warn("add task to Routine failed, the sync service has been stopped")
		return nil
	}
	if err := r.queue.Put(event); err != nil {
		return fmt.Errorf("add task to the queue failed, reason: %s", err)
	}
	return nil
}

func (r *Routine) getServiceState() string {
	return r.queryState()
}

func (r *Routine) end() {
	if !atomic.CompareAndSwapInt32(&r.start, 1, 0) {
		return
	}
	r.queue.Dispose()
	close(r.stop)
	close(r.out)
}
