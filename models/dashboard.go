package models

import "time"

// DashboardEvent is one row of the recent events feed shown on the
// dashboard: campaign sends, opens, clicks, public list opt-ins and
// melodies site activity. Only the fields relevant to the event type
// are set.
type DashboardEvent struct {
	Type           string    `db:"type" json:"type"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	CampaignID     *int      `db:"campaign_id" json:"campaign_id"`
	CampaignName   *string   `db:"campaign_name" json:"campaign_name"`
	SubscriberID   *int      `db:"subscriber_id" json:"subscriber_id"`
	Email          *string   `db:"email" json:"email"`
	SubscriberName *string   `db:"subscriber_name" json:"subscriber_name"`
	ListName       *string   `db:"list_name" json:"list_name"`
	URL            *string   `db:"url" json:"url"`
	Track          *string   `db:"track" json:"track"`
	Device         *string   `db:"device" json:"device"`
	Browser        *string   `db:"browser" json:"browser"`
}
