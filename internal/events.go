package internal

import (
	"runtime"
	"sync"
	"syscall"

	"golang.org/x/sys/windows"

	"kakaotalkadblock/internal/win/winapi"
)

const (
	wmSyncHooks = winapi.WmApp + 1
	wmStopHooks = winapi.WmApp + 2
)

var hookedEventRanges = [][2]uint32{
	{winapi.EventObjectCreate, winapi.EventObjectHide},
	{winapi.EventObjectLocationchange, winapi.EventObjectNamechange},
}

var winEvents = make(chan struct{}, 1)

var winEventCallback = syscall.NewCallback(func(_, _, hwnd, idObject, _, _, _ uintptr) uintptr {
	if hwnd == 0 || int32(idObject) != winapi.ObjidWindow {
		return 0
	}
	select {
	case winEvents <- struct{}{}:
	default:
	}
	return 0
})

// WinEvent hooks must be serviced and removed on the thread that installed them.
type winEventHooks struct {
	threadId  uint32
	mutex     sync.Mutex
	processes map[uint32]struct{}
	done      chan struct{}
}

func startWinEventHooks() *winEventHooks {
	hooks := &winEventHooks{done: make(chan struct{})}
	ready := make(chan struct{})
	go hooks.run(ready)
	<-ready
	return hooks
}

func (h *winEventHooks) watch(processes map[uint32]struct{}) {
	h.mutex.Lock()
	h.processes = processes
	h.mutex.Unlock()
	winapi.PostThreadMessage(h.threadId, wmSyncHooks, 0, 0)
}

func (h *winEventHooks) stop() {
	if winapi.PostThreadMessage(h.threadId, wmStopHooks, 0, 0) {
		<-h.done
	}
}

func (h *winEventHooks) run(ready chan<- struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(h.done)

	var msg winapi.Msg
	// PostThreadMessage fails until the thread has a message queue.
	winapi.PeekMessage(&msg, 0, winapi.WmUser, winapi.WmUser, winapi.PmNoremove)
	h.threadId = windows.GetCurrentThreadId()
	close(ready)

	installed := make(map[uint32][]uintptr)
	defer func() {
		for _, hooks := range installed {
			unhookAll(hooks)
		}
	}()
	for winapi.GetMessage(&msg, 0, 0, 0) > 0 {
		switch msg.Message {
		case wmSyncHooks:
			h.sync(installed)
		case wmStopHooks:
			return
		}
	}
}

func (h *winEventHooks) sync(installed map[uint32][]uintptr) {
	h.mutex.Lock()
	wanted := h.processes
	h.mutex.Unlock()

	for processId, hooks := range installed {
		if _, ok := wanted[processId]; !ok {
			unhookAll(hooks)
			delete(installed, processId)
		}
	}
	for processId := range wanted {
		if _, ok := installed[processId]; ok {
			continue
		}
		hooks := make([]uintptr, 0, len(hookedEventRanges))
		for _, events := range hookedEventRanges {
			hook := winapi.SetWinEventHook(events[0], events[1], winEventCallback, processId, 0, winapi.WineventOutofcontext)
			if hook != 0 {
				hooks = append(hooks, hook)
			}
		}
		installed[processId] = hooks
	}
}

func unhookAll(hooks []uintptr) {
	for _, hook := range hooks {
		winapi.UnhookWinEvent(hook)
	}
}
