package spaces

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/porte/directory"
)

// fakeRegistre answers the directory API the way an undeployed then deployed Registre
// would: the first import is a 404, every later call succeeds. It records the calls in
// order, with the body of the import that succeeded.
type fakeRegistre struct {
	mu       sync.Mutex
	calls    []string
	imported []directory.ImportSpace
}

func (f *fakeRegistre) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id, secret, ok := r.BasicAuth(); !ok || id != "sablier" || secret != "s3cret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	f.calls = append(f.calls, r.Method+" "+r.URL.Path)
	if r.URL.Path == "/api/directory/import" {
		if len(f.calls) == 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body struct {
			Spaces []directory.ImportSpace `json:"spaces"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.imported = body.Spaces
	}
	_ = json.NewEncoder(w).Encode(teamSnapshot())
}

func TestRunImportsBeforeTheFirstSync(t *testing.T) {
	orm := withIdentities(t, newTestORM(t, "registre-run"))
	seedTeam(t, orm)
	fake := &fakeRegistre{}
	server := httptest.NewServer(fake)
	defer server.Close()

	registre := newRegistre(orm)
	registre.retry = retryPolicy{first: time.Millisecond, max: time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	syncer := &directory.Syncer{
		Client: directory.Client{Issuer: server.URL, ClientID: "sablier", ClientSecret: "s3cret"},
		Apply: func(ctx context.Context, snapshot directory.Snapshot) error {
			defer cancel()
			return registre.Apply(ctx, snapshot)
		},
	}
	registre.Run(ctx, syncer)

	want := []string{"POST /api/directory/import", "POST /api/directory/import", "GET /api/directory/snapshot"}
	if !slices.Equal(fake.calls, want) {
		t.Fatalf("calls = %v, want %v: retry the import, then sync", fake.calls, want)
	}
	if len(fake.imported) != 1 || fake.imported[0].Name != "team" {
		t.Fatalf("imported = %+v, want the one local space", fake.imported)
	}
	members := fake.imported[0].Members
	wantMembers := []directory.ImportMember{
		{Subject: "sub-alice", Role: directory.RoleOwner},
		{Email: "carol@x.test", Role: directory.RoleMember},
		{Subject: "sub-dave", Role: directory.RoleAdmin},
	}
	if !slices.Equal(members, wantMembers) {
		t.Fatalf("imported members = %+v, want %+v: subject when known, else email", members, wantMembers)
	}
	if roleOf(membershipState(t, orm), "r-team:Team/bob@x.test") != directory.RoleAdmin {
		t.Fatal("the sync after the import did not apply the snapshot")
	}
}
