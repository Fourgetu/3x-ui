package service

import (
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

// Editing an old inbound form must preserve both the official renewal state
// and the fork's separately stored per-client route.
func TestSpeedLimitSurvivesStaleInboundFormAndWeeklyProjection(t *testing.T) {
	setupBulkDB(t)
	ib := seedRenewableNeighbour(t, 43201, nil)
	form := *ib
	rec := lookupClientRecord(t, "x@stale")
	limit := model.ClientSpeedLimit{ClientID: rec.Id, InboundID: ib.Id, UploadMbps: 12, DownloadMbps: 34}
	if err := database.GetDB().Create(&limit).Error; err != nil {
		t.Fatal(err)
	}
	updated := *rec.ToClient()
	updated.ResetWeekday = 3
	if _, err := (&ClientService{}).Update(&InboundService{}, rec.Id, updated, 0); err != nil {
		t.Fatal(err)
	}
	form.Remark = "edited after weekly renewal change"
	if _, _, err := (&InboundService{}).UpdateInbound(&form); err != nil {
		t.Fatal(err)
	}
	page, err := (&ClientService{}).ListPaged(&InboundService{}, nil, ClientPageParams{Search: "x@stale"})
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	got := page.Items[0]
	if got.ResetWeekday != 3 || len(got.SpeedLimits) != 1 || got.SpeedLimits[0].DownloadMbps != 34 {
		t.Fatalf("weekly policy or speed limit lost: %+v", got)
	}
}

// The optional orphan cleanup must use the official delete path and leave a
// client still attached to another inbound (and its limiter) alone.
func TestBulkInboundPurgePreservesSharedClientSpeedLimit(t *testing.T) {
	setupBulkDB(t)
	ib := seedRenewableNeighbour(t, 43202, nil)
	rec := lookupClientRecord(t, "x@stale")
	other := mkInbound(t, 43203, model.VLESS, clientsSettings(t, []model.Client{*rec.ToClient()}))
	if err := (&ClientService{}).SyncInbound(nil, other.Id, []model.Client{*rec.ToClient()}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{ib.Id, other.Id} {
		if err := database.GetDB().Create(&model.ClientSpeedLimit{ClientID: rec.Id, InboundID: id, DownloadMbps: 50}).Error; err != nil {
			t.Fatal(err)
		}
	}
	result, _, err := (&InboundService{}).DelInboundsPurgeOrphans([]int{ib.Id}, true)
	if err != nil || result.Deleted != 1 {
		t.Fatalf("delete=%+v err=%v", result, err)
	}
	if !inboundLinksEmail(t, other.Id, rec.Email) {
		t.Fatal("shared client was removed from the surviving inbound")
	}
	var limits []model.ClientSpeedLimit
	if err := database.GetDB().Find(&limits).Error; err != nil {
		t.Fatal(err)
	}
	if len(limits) != 1 || limits[0].InboundID != other.Id {
		t.Fatalf("wrong speed limits after cleanup: %+v", limits)
	}
	var orphans int64
	if err := database.GetDB().Model(&model.ClientRecord{}).Where("email = ?", "y@stale").Count(&orphans).Error; err != nil || orphans != 0 {
		t.Fatalf("orphan remains: count=%d err=%v", orphans, err)
	}
}

// Both official membership update paths must remove only the detached route,
// while settling the shared record's flow from its surviving inbound.
func TestInboundDetachSettlesFlowAndPreservesOtherSpeedLimit(t *testing.T) {
	for _, delta := range []bool{false, true} {
		name := "full sync"
		if delta {
			name = "client delta"
		}
		t.Run(name, func(t *testing.T) {
			setupBulkDB(t)
			svc := &ClientService{}
			client := model.Client{Email: "shared@speed", ID: "ce8d33df-3a64-4f10-8f9b-91c3a8e0c333", Enable: true, Flow: "xtls-rprx-vision"}
			first := mkInbound(t, 43204, model.VLESS, clientsSettings(t, []model.Client{client}))
			if err := svc.SyncInbound(nil, first.Id, []model.Client{client}); err != nil {
				t.Fatal(err)
			}
			client.Flow = ""
			other := mkInbound(t, 43205, model.VLESS, clientsSettings(t, []model.Client{client}))
			if err := svc.SyncInbound(nil, other.Id, []model.Client{client}); err != nil {
				t.Fatal(err)
			}
			rec := lookupClientRecord(t, client.Email)
			if rec.Flow != "xtls-rprx-vision" {
				t.Fatalf("shared Vision flow lost before detach: %q", rec.Flow)
			}
			for _, id := range []int{first.Id, other.Id} {
				if err := database.GetDB().Create(&model.ClientSpeedLimit{ClientID: rec.Id, InboundID: id, DownloadMbps: 50}).Error; err != nil {
					t.Fatal(err)
				}
			}
			var err error
			if delta {
				err = svc.ApplyInboundClientDelta(nil, first.Id, nil, []string{client.Email})
			} else {
				err = svc.SyncInbound(nil, first.Id, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			rec = lookupClientRecord(t, client.Email)
			if rec.Flow != "" || !inboundLinksEmail(t, other.Id, client.Email) || inboundLinksEmail(t, first.Id, client.Email) {
				t.Fatalf("membership or effective flow incorrect after detach: %+v", rec)
			}
			var limits []model.ClientSpeedLimit
			if err := database.GetDB().Find(&limits).Error; err != nil {
				t.Fatal(err)
			}
			if len(limits) != 1 || limits[0].InboundID != other.Id || limits[0].DownloadMbps != 50 {
				t.Fatalf("surviving speed limit changed: %+v", limits)
			}
		})
	}
}
