/*
   All types and functions related to the mousebind package.
   They live here because they rely on state in XUtil.
*/
package xgbutil

import "github.com/BurntSushi/xgb/xproto"

// MouseBindCallback operates in the spirit of Callback, except that it works
// specifically on mouse bindings.
type MouseBindCallback interface {
	Connect(xu *XUtil, win xproto.Window, buttonStr string,
		propagate bool, grab bool) error
	Run(xu *XUtil, ev interface{})
}

// MouseBindKey is the type of the key in the map of mouse bindings.
// It essentially represents the tuple
// (event type, window id, modifier, button).
type MouseBindKey struct {
	Evtype int
	Win    xproto.Window
	Mod    uint16
	Button xproto.Button
}

// AttackMouseBindCallback associates an (event, window, mods, button)
// with a callback.
func (xu *XUtil) AttachMouseBindCallback(evtype int, win xproto.Window,
	mods uint16, button xproto.Button, fun MouseBindCallback) {

	xu.mousebindsLck.Lock()
	defer xu.mousebindsLck.Unlock()

	// Create key
	key := MouseBindKey{evtype, win, mods, button}

	// Do we need to allocate?
	if _, ok := xu.mousebinds[key]; !ok {
		xu.mousebinds[key] = make([]MouseBindCallback, 0)
	}

	xu.mousebinds[key] = append(xu.mousebinds[key], fun)
	xu.mousegrabs[key] += 1
}

// MouseBindKeys returns a copy of all the keys in the 'mousebinds' map.
func (xu *XUtil) MouseBindKeys() []MouseBindKey {
	xu.mousebindsLck.RLock()
	defer xu.mousebindsLck.RUnlock()

	keys := make([]MouseBindKey, len(xu.mousebinds))
	i := 0
	for key, _ := range xu.mousebinds {
		keys[i] = key
		i++
	}
	return keys
}

// MouseBindCallbacks returns a slice of callbacks for a particular key.
func (xu *XUtil) MouseBindCallbacks(key MouseBindKey) []MouseBindCallback {
	xu.mousebindsLck.RLock()
	defer xu.mousebindsLck.RUnlock()

	cbs := make([]MouseBindCallback, len(xu.mousebinds[key]))
	for i, cb := range xu.mousebinds[key] {
		cbs[i] = cb
	}
	return cbs
}

// RunMouseBindCallbacks executes every callback corresponding to a
// particular event/window/mod/button tuple.
func (xu *XUtil) RunMouseBindCallbacks(event interface{}, evtype int,
	win xproto.Window, mods uint16, button xproto.Button) {

	key := MouseBindKey{evtype, win, mods, button}
	for _, cb := range xu.MouseBindCallbacks(key) {
		cb.Run(xu, event)
	}
}

// ConnectedMouseBind checks to see if there are any key binds for a particular
// event type already in play. This is to work around comparing function
// pointers (not allowed in Go), which would be used in 'Connected'.
func (xu *XUtil) ConnectedMouseBind(evtype int, win xproto.Window) bool {
	xu.mousebindsLck.RLock()
	defer xu.mousebindsLck.RUnlock()

	// Since we can't create a full key, loop through all mouse binds
	// and check if evtype and window match.
	for key, _ := range xu.mousebinds {
		if key.Evtype == evtype && key.Win == win {
			return true
		}
	}

	return false
}

// DetachMouseBindWindow removes all callbacks associated with a particular
// window and event type (either ButtonPress or ButtonRelease)
// Also decrements the counter in the corresponding 'mousegrabs' map
// appropriately.
func (xu *XUtil) DetachMouseBindWindow(evtype int, win xproto.Window) {
	xu.mousebindsLck.Lock()
	defer xu.mousebindsLck.Unlock()

	// Since we can't create a full key, loop through all mouse binds
	// and check if evtype and window match.
	for key, _ := range xu.mousebinds {
		if key.Evtype == evtype && key.Win == win {
			xu.mousegrabs[key] -= len(xu.mousebinds[key])
			delete(xu.mousebinds, key)
		}
	}
}

// MouseBindGrabs returns the number of grabs on a particular
// event/window/mods/button combination. Namely, this combination
// uniquely identifies a grab. If it's repeated, we get BadAccess.
func (xu *XUtil) MouseBindGrabs(evtype int, win xproto.Window, mods uint16,
	button xproto.Button) int {

	xu.mousebindsLck.RLock()
	defer xu.mousebindsLck.RUnlock()

	key := MouseBindKey{evtype, win, mods, button}
	return xu.mousegrabs[key] // returns 0 if key does not exist
}

// MouseDragFun is the kind of function used on each dragging step
// and at the end of a drag.
type MouseDragFun func(xu *XUtil, rootX, rootY, eventX, eventY int)

// MouseDragBeginFun is the kind of function used to initialize a drag.
// The difference between this and MouseDragFun is that the begin function
// returns a bool (of whether or not to cancel the drag) and an X resource
// identifier corresponding to a cursor.
type MouseDragBeginFun func(xu *XUtil, rootX, rootY,
	eventX, eventY int) (bool, xproto.Cursor)

// MouseDrag true when a mouse drag is in progress.
func (xu *XUtil) MouseDrag() bool {
	return xu.mouseDrag
}

// MouseDragSet sets whether a mouse drag is in progress.
func (xu *XUtil) MouseDragSet(dragging bool) {
	xu.mouseDrag = dragging
}

// MouseDragStep returns the function currently associated with each
// step of a mouse drag.
func (xu *XUtil) MouseDragStep() MouseDragFun {
	return xu.mouseDragStep
}

// MouseDragStepSet sets the function associated with the step of a drag.
func (xu *XUtil) MouseDragStepSet(f MouseDragFun) {
	xu.mouseDragStep = f
}

// MouseDragEnd returns the function currently associated with the
// end of a mouse drag.
func (xu *XUtil) MouseDragEnd() MouseDragFun {
	return xu.mouseDragEnd
}

// MouseDragEndSet sets the function associated with the end of a drag.
func (xu *XUtil) MouseDragEndSet(f MouseDragFun) {
	xu.mouseDragEnd = f
}
