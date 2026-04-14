package dom

import (
	"github.com/furkandgn/goglweb/internal/parser/html"
	"reflect"
)

// EventType represents an event type
type EventType string

const (
	EventClick   EventType = "click"
	EventHover   EventType = "hover"
	EventKeyDown EventType = "keydown"
	EventKeyUp   EventType = "keyup"
)

// Event represents a DOM event
type Event struct {
	Type           EventType
	Target         *html.Node
	X, Y           float64 // Mouse position (for click, hover)
	Key            string  // For keyboard event
	PreventDefault func()  // To prevent the event's default behavior
}

// EventListener represents an event listener function
type EventListener func(Event)

// EventManager manages events
type EventManager struct {
	listeners map[*html.Node]map[EventType][]EventListener
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[*html.Node]map[EventType][]EventListener),
	}
}

// AddEventListener adds an event listener to a node
func (em *EventManager) AddEventListener(node *html.Node, eventType EventType, listener EventListener) {
	if em.listeners[node] == nil {
		em.listeners[node] = make(map[EventType][]EventListener)
	}
	em.listeners[node][eventType] = append(em.listeners[node][eventType], listener)
}

// RemoveEventListener removes an event listener from a node
func (em *EventManager) RemoveEventListener(node *html.Node, eventType EventType, listener EventListener) {
	if em.listeners[node] == nil {
		return
	}
	listeners := em.listeners[node][eventType]
	listenerPtr := reflect.ValueOf(listener).Pointer()
	for i, l := range listeners {
		if reflect.ValueOf(l).Pointer() == listenerPtr {
			em.listeners[node][eventType] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// findParentChain finds the parent chain from target to root
func findParentChain(target, root *html.Node) []*html.Node {
	if target == nil || root == nil {
		return nil
	}

	var path []*html.Node
	var found bool

	var findPath func(node *html.Node, target *html.Node) bool
	findPath = func(node *html.Node, target *html.Node) bool {
		if node == target {
			path = append(path, node)
			return true
		}
		for _, child := range node.Children {
			if findPath(child, target) {
				path = append([]*html.Node{node}, path...)
				return true
			}
		}
		return false
	}

	if findPath(root, target) {
		found = true
	}

	if !found {
		return []*html.Node{target}
	}
	return path
}

// DispatchEvent dispatches an event
// Event bubbling: calls listeners in all parent chain from target to root
func (em *EventManager) DispatchEvent(event Event, root *html.Node) {
	if event.Target == nil {
		return
	}

	// Find parent chain (from target to root)
	chain := findParentChain(event.Target, root)

	// Bubbling phase: call listeners on all nodes from target to root
	for _, node := range chain {
		if nodeListeners, ok := em.listeners[node]; ok {
			if listeners, ok := nodeListeners[event.Type]; ok {
				for _, listener := range listeners {
					listener(event)
				}
			}
		}
	}
}

// ClearListeners clears listeners of a node and all its subtree
func (em *EventManager) ClearListeners(node *html.Node) {
	if node == nil {
		return
	}
	delete(em.listeners, node)
	for _, child := range node.Children {
		em.ClearListeners(child)
	}
}
