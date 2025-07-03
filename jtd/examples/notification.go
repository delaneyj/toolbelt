package types

import (
	"time"
)

// Notification represents different types of notifications
// Notification is a discriminated union
type Notification interface {
	isNotification()
	Type() string
}

// NotificationEmail is the "email" variant of Notification
type NotificationEmail struct {
	Type_       string     `json:"type"`
	Attachments []struct{} `json:"attachments,omitempty"`
	Bcc         []string   `json:"bcc,omitempty"`
	Body        string     `json:"body"`
	Cc          []string   `json:"cc,omitempty"`
	From        string     `json:"from"`
	IsHtml      bool       `json:"isHtml"`
	Subject     string     `json:"subject"`
	To          string     `json:"to"`
}

func (NotificationEmail) isNotification() {}
func (v NotificationEmail) Type() string  { return v.Type_ }

// NotificationSms is the "sms" variant of Notification
type NotificationSms struct {
	Type_    string  `json:"type"`
	Message  string  `json:"message"`
	Provider *string `json:"provider,omitempty"`
	To       string  `json:"to"`
}

func (NotificationSms) isNotification() {}
func (v NotificationSms) Type() string  { return v.Type_ }

// NotificationPush is the "push" variant of Notification
type NotificationPush struct {
	Type_       string            `json:"type"`
	Badge       *int32            `json:"badge,omitempty"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
	DeviceToken string            `json:"deviceToken"`
	Sound       *string           `json:"sound,omitempty"`
	Title       string            `json:"title"`
}

func (NotificationPush) isNotification() {}
func (v NotificationPush) Type() string  { return v.Type_ }

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
)

type NotificationLog struct {
	Attempts     int32              `json:"attempts"`
	Error        *string            `json:"error,omitempty"`
	Id           string             `json:"id"`
	Notification Notification       `json:"notification"`
	Status       NotificationStatus `json:"status"`
	Timestamp    time.Time          `json:"timestamp"`
}
