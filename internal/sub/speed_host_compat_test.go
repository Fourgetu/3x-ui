package sub

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

// Host authorization must run before speed ports are applied. Cover every
// subscription/share export and both ordering modes, including a second client
// without a limiter to catch mutable inbound/endpoint leakage.
func TestSubSpeedPortHostBindings(t *testing.T) {
	for _, mode := range []string{"inbound-first", "hosts-first"} {
		for _, scenario := range []string{"bound", "multiple", "disabled", "deleted", "excluded", "no-hosts"} {
			t.Run(mode+"/"+scenario, func(t *testing.T) {
				seedSubDB(t)
				setHostSortMode(t, mode)
				ib := seedSubInbound(t, "s1", "limited", 4499, 1, tcpTLSStream)
				seedSubClient(t, ib, "s1", "unlimited@e")
				db := database.GetDB()
				var client model.ClientRecord
				if err := db.Where("email = ?", "limited@e").First(&client).Error; err != nil {
					t.Fatal(err)
				}
				port := 35001
				if err := db.Create(&model.ClientSpeedLimit{ClientID: client.Id, InboundID: ib.Id, Enabled: true, DownloadMbps: 10, ExternalPort: &port}).Error; err != nil {
					t.Fatal(err)
				}
				var limited, unlimited []string
				if scenario == "no-hosts" {
					limited = []string{"203.0.113.5:35001"}
					unlimited = []string{"203.0.113.5:4499"}
				} else {
					a := seedHost(t, &model.Host{InboundId: ib.Id, Address: "a.example.com", Port: 8443, SortOrder: 1})
					b := seedHost(t, &model.Host{InboundId: ib.Id, Address: "b.example.com", Port: 9443, SortOrder: 2})
					ids := []int{b.Id}
					limited = []string{"b.example.com:35001"}
					unlimited = []string{"a.example.com:8443", "b.example.com:9443"}
					switch scenario {
					case "multiple":
						ids = []int{a.Id, b.Id}
						limited = []string{"a.example.com:35001", "b.example.com:35001"}
					case "disabled", "deleted", "excluded":
						limited = nil
						unlimited = []string{"a.example.com:8443"}
						var err error
						switch scenario {
						case "disabled":
							err = db.Model(b).Update("is_disabled", true).Error
						case "deleted":
							err = db.Delete(b).Error
						case "excluded":
							err = db.Model(b).Update("exclude_from_sub_types", `["raw","json","clash"]`).Error
						}
						if err != nil {
							t.Fatal(err)
						}
					}
					if err := db.Model(&client).Update("client_host_rule_ids", model.EncodeClientHostRuleIds(ids)).Error; err != nil {
						t.Fatal(err)
					}
				}
				for _, format := range []string{"raw", "json", "clash", "client-link", "inbound-links"} {
					t.Run(format, func(t *testing.T) {
						svc := NewSubService("")
						t.Cleanup(svc.enableSettingsCache())
						want := slices.Clone(limited)
						if format != "client-link" {
							want = append(want, unlimited...)
						}
						slices.Sort(want)
						// Repeat on the same service to exercise cached settings too.
						for range 2 {
							got := speedHostEndpoints(t, format, svc, ib)
							slices.Sort(got)
							if !slices.Equal(got, want) {
								t.Fatalf("endpoints = %v, want %v", got, want)
							}
						}
						if ib.Port != 4499 || ib.StreamSettings != tcpTLSStream {
							t.Fatalf("export mutated source inbound: %+v", ib)
						}
					})
				}
			})
		}
	}
}

func speedHostEndpoints(t *testing.T, format string, svc *SubService, ib *model.Inbound) []string {
	t.Helper()
	var out, links []string
	switch format {
	case "raw":
		var err error
		links, _, _, _, err = svc.GetSubs("s1", "sub.example.com")
		if err != nil {
			t.Fatal(err)
		}
	case "client-link":
		links = NewLinkProvider().LinksForClient("sub.example.com", ib, "limited@e")
	case "inbound-links":
		links = NewLinkProvider().LinksForInbounds("sub.example.com", []*model.Inbound{ib})
	case "json":
		body, _, err := NewSubJsonService("", "", "", "", svc).GetJson("s1", "sub.example.com", true)
		if err != nil {
			t.Fatal(err)
		}
		var configs []struct {
			Outbounds []struct {
				Tag      string `json:"tag"`
				Settings struct {
					Address string `json:"address"`
					Port    int    `json:"port"`
				} `json:"settings"`
			} `json:"outbounds"`
		}
		if err := json.Unmarshal([]byte(body), &configs); err != nil {
			t.Fatal(err)
		}
		for _, config := range configs {
			for _, ob := range config.Outbounds {
				if ob.Tag == "proxy" {
					out = append(out, fmt.Sprintf("%s:%d", ob.Settings.Address, ob.Settings.Port))
				}
			}
		}
	case "clash":
		body, _, err := NewSubClashService(false, "", false, svc).GetClash("s1", "sub.example.com")
		if err != nil {
			t.Fatal(err)
		}
		var config struct {
			Proxies []struct {
				Server string `yaml:"server"`
				Port   int    `yaml:"port"`
			} `yaml:"proxies"`
		}
		if err := yaml.Unmarshal([]byte(body), &config); err != nil {
			t.Fatal(err)
		}
		for _, proxy := range config.Proxies {
			out = append(out, fmt.Sprintf("%s:%d", proxy.Server, proxy.Port))
		}
	}
	for _, line := range splitLinkLines(strings.Join(links, "\n")) {
		u, err := url.Parse(line)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, u.Host)
	}
	return out
}
