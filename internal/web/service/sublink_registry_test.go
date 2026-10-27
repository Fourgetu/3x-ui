package service

import (
	"sync"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

type fakeSubLinkProvider struct {
	links []string
}

type scopedSubLinkProvider struct {
	fakeSubLinkProvider
	seen []int
}

func (p *scopedSubLinkProvider) LinksForInbounds(host string, inbounds []*model.Inbound) []string {
	for _, inbound := range inbounds {
		p.seen = append(p.seen, inbound.Id)
	}
	return p.links
}

func TestInboundQRLinksScopeAndOwnership(t *testing.T) {
	setupBulkDB(t)
	ib := mkInbound(t, 43209, model.VLESS, clientsSettings(t, nil))
	other := mkInbound(t, 43210, model.VLESS, clientsSettings(t, nil))
	if err := database.GetDB().Model(other).Update("user_id", ib.UserId+1).Error; err != nil {
		t.Fatal(err)
	}
	previous := subLinkProvider()
	t.Cleanup(func() { RegisterSubLinkProvider(previous) })
	provider := &scopedSubLinkProvider{fakeSubLinkProvider: fakeSubLinkProvider{links: []string{"vless://bound:35001"}}}
	RegisterSubLinkProvider(provider)
	svc := &InboundService{}
	links, err := svc.GetInboundLinks("panel.example", ib.UserId, ib.Id)
	if err != nil || len(links) != 1 || links[0] != provider.links[0] || len(provider.seen) != 1 || provider.seen[0] != ib.Id {
		t.Fatalf("scoped links=%v, rendered=%v, err=%v", links, provider.seen, err)
	}
	for _, id := range []int{other.Id, 999999} {
		if _, err := svc.GetInboundLinks("panel.example", ib.UserId, id); err == nil {
			t.Fatalf("unexpected access to inbound %d", id)
		}
	}
	if len(provider.seen) != 1 {
		t.Fatalf("unauthorized inbound reached renderer: %v", provider.seen)
	}
}

func (f fakeSubLinkProvider) SubLinksForSubId(host, subId string) ([]string, error) {
	return f.links, nil
}

func (f fakeSubLinkProvider) LinksForClient(host string, inbound *model.Inbound, email string) []string {
	return f.links
}

func (f fakeSubLinkProvider) LinksForInbounds(host string, inbounds []*model.Inbound) []string {
	return f.links
}

// Startup writes the provider registry while link requests read it; under -race an
// unguarded registry fails here.
func TestSubLinkProviderRegistrationIsRaceFree(t *testing.T) {
	t.Cleanup(func() { RegisterSubLinkProvider(nil) })

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			RegisterSubLinkProvider(fakeSubLinkProvider{links: []string{"vless://example"}})
		}()
		go func() {
			defer wg.Done()
			_, _ = (&InboundService{}).GetSubLinks("panel.example", "subid")
		}()
	}
	wg.Wait()
}
