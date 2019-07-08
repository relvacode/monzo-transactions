package monzo

import (
	"bytes"
	"errors"
	"github.com/buger/jsonparser"
	"github.com/mailru/easyjson"
	"sort"
)

var (
	ErrNoSuchEventType = errors.New("webhook: no registered type for that event")
)

// An Event is a web-hook event delivered from Monzo
type Event interface {
	easyjson.Unmarshaler
	easyjson.Marshaler
}

type binarySearchNode struct {
	Key     []byte
	Factory func() Event
}

type binarySearch []binarySearchNode

func (t binarySearch) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t binarySearch) Len() int {
	return len(t)
}

func (t binarySearch) Less(i, j int) bool {
	return bytes.Compare(t[i].Key, t[j].Key) < 0
}

func (t binarySearch) Get(k []byte) (binarySearchNode, bool) {
	if t == nil || len(t) == 0 {
		return binarySearchNode{}, false
	}
	if len(t) == 1 && bytes.Equal(t[0].Key, k) {
		return t[0], true
	}

	i := sort.Search(len(t), func(i int) bool {
		return bytes.Compare(t[i].Key, k) >= 0
	})
	if i < len(t) && bytes.Equal(t[i].Key, k) {
		return t[i], true
	}
	return binarySearchNode{}, false
}

var search binarySearch

func register(t string, f func() Event) {
	search = append(search, binarySearchNode{
		Key:     []byte(t),
		Factory: f,
	})
	sort.Sort(search)
}

// GetEvent gets event data from the given raw JSON encoded bytes.
// It first uses jsonparser to get the event type, and using the registry
func GetEvent(b []byte) (Event, error) {
	t, _, _, err := jsonparser.Get(b, "type")
	if err != nil {
		return nil, err
	}

	fn, ok := search.Get(t)
	if !ok {
		return nil, ErrNoSuchEventType
	}

	raw, _, _, err := jsonparser.Get(b, "data")
	dt := fn.Factory()

	err = easyjson.Unmarshal(raw, dt)
	if err != nil {
		return nil, err
	}
	return dt, nil
}
