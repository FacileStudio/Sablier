package antenne

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"

	antenneclient "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/antenne-client/go/catalog"
)

func TestEveryListenedChannelHasAHandlerAndEveryChannelIsInTheCatalog(t *testing.T) {
	s := NewService(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	var handled []string
	for channel := range s.listeners() {
		handled = append(handled, string(channel))
	}
	slices.Sort(handled)
	listened := ListenEvents()
	slices.Sort(listened)
	if !slices.Equal(handled, listened) {
		t.Fatalf("handlers for %v, registered to listen to %v", handled, listened)
	}
	for _, channel := range append(EmitEvents(), ListenEvents()...) {
		if !slices.Contains(catalog.All(), catalog.Channel(channel)) {
			t.Errorf("%s is not in antenne-client's catalog", channel)
		}
	}
}

func TestABusThatIsDownKeepsTheClientDialling(t *testing.T) {
	s := NewService(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := s.emitQueued("project.created", json.RawMessage(`{}`)); !errors.Is(err, antenneclient.ErrNotConnected) {
		t.Fatalf("emit without a client = %v, want ErrNotConnected", err)
	}
	if err := s.connect("http://127.0.0.1:1", "secret"); err == nil {
		t.Fatal("the first dial's error must reach the settings page")
	}
	defer s.Shutdown()
	if s.client == nil {
		t.Fatal("a failed first dial must leave a client that keeps retrying")
	}
	if err := s.emitQueued("project.created", json.RawMessage(`{}`)); !errors.Is(err, antenneclient.ErrNotConnected) {
		t.Fatalf("emit while dialling = %v, want ErrNotConnected", err)
	}
}
