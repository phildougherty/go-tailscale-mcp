package tailscale

import (
	"encoding/json"
	"time"
)

// Status represents the Tailscale status
type Status struct {
	BackendState  string              `json:"BackendState"`
	AuthURL       string              `json:"AuthURL,omitempty"`
	Self          *PeerStatus         `json:"Self"`
	Health        []string            `json:"Health"`
	CurrentTailnet *TailnetStatus     `json:"CurrentTailnet,omitempty"`
	Peer          map[string]*PeerStatus `json:"Peer"`
	User          map[string]*User    `json:"User,omitempty"`
}

// PeerStatus represents the status of a peer device
type PeerStatus struct {
	ID               string          `json:"ID"`
	PublicKey        string          `json:"PublicKey"`
	HostName         string          `json:"HostName"`
	DNSName          string          `json:"DNSName"`
	OS               string          `json:"OS"`
	UserID           json.RawMessage `json:"UserID"`
	TailscaleIPs     []string  `json:"TailscaleIPs"`
	AllowedIPs       []string  `json:"AllowedIPs"`
	Addrs            []string  `json:"Addrs"`
	CurAddr          string    `json:"CurAddr"`
	RxBytes          int64     `json:"RxBytes"`
	TxBytes          int64     `json:"TxBytes"`
	Created          time.Time `json:"Created"`
	LastWrite        time.Time `json:"LastWrite"`
	LastSeen         time.Time `json:"LastSeen"`
	LastHandshake    time.Time `json:"LastHandshake"`
	Online           bool      `json:"Online"`
	ExitNode         bool      `json:"ExitNode"`
	ExitNodeOption   bool      `json:"ExitNodeOption"`
	Active           bool      `json:"Active"`
	PeerAPIURL       []string  `json:"PeerAPIURL"`
	Capabilities     []string  `json:"Capabilities"`
	Tags             []string  `json:"Tags"`
	PrimaryRoutes    []string  `json:"PrimaryRoutes,omitempty"`
	Expired          bool      `json:"Expired"`
	KeyExpiry        time.Time `json:"KeyExpiry"`
}

// TailnetStatus represents the current tailnet status
type TailnetStatus struct {
	Name            string `json:"Name"`
	MagicDNSSuffix  string `json:"MagicDNSSuffix"`
	MagicDNSEnabled bool   `json:"MagicDNSEnabled"`
}

// User represents a Tailscale user
type User struct {
	ID          json.RawMessage `json:"ID"`
	LoginName   string          `json:"LoginName"`
	DisplayName string          `json:"DisplayName"`
	ProfilePicURL string        `json:"ProfilePicURL"`
}

// Profile represents a Tailscale profile
type Profile struct {
	ID       string `json:"id"`       // Profile ID (e.g., "826b")
	Tailnet  string `json:"tailnet"`  // Tailnet name/email
	Account  string `json:"account"`  // Account email
	Active   bool   `json:"active"`   // Whether this profile is currently active
}

// Device represents a device in the network
type Device struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Hostname      string    `json:"hostname"`
	OS            string    `json:"os"`
	Addresses     []string  `json:"addresses"`
	User          string    `json:"user"`
	Tags          []string  `json:"tags"`
	Authorized    bool      `json:"authorized"`
	KeyExpiry     time.Time `json:"keyExpiry"`
	LastSeen      time.Time `json:"lastSeen"`
	Online        bool      `json:"online"`
	ExitNode      bool      `json:"exitNode"`
	PrimaryRoutes []string  `json:"primaryRoutes,omitempty"`
}

// ACL represents Access Control List configuration
type ACL struct {
	Groups     map[string][]string `json:"groups"`
	Hosts      map[string]string   `json:"hosts"`
	TagOwners  map[string][]string `json:"tagOwners"`
	ACLs       []ACLRule           `json:"acls"`
	Tests      []ACLTest           `json:"tests,omitempty"`
	AutoApprovers map[string][]string `json:"autoApprovers,omitempty"`
	RawPolicy  string              `json:"-"` // Raw HuJSON policy from API
}

// ACLRule represents a single ACL rule
type ACLRule struct {
	Action string   `json:"action"`
	Users  []string `json:"users"`
	Ports  []string `json:"ports"`
}

// ACLTest represents an ACL test case
type ACLTest struct {
	User  string   `json:"user"`
	Allow []string `json:"allow"`
	Deny  []string `json:"deny,omitempty"`
}

// AuthKey represents an authentication key
type AuthKey struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Created     time.Time `json:"created"`
	Expires     time.Time `json:"expires"`
	Reusable    bool      `json:"reusable"`
	Ephemeral   bool      `json:"ephemeral"`
	Preauthorized bool    `json:"preauthorized"`
	Tags        []string  `json:"tags,omitempty"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	MagicDNS    bool     `json:"magicDNS"`
	Nameservers []string `json:"nameservers"`
	Domains     []string `json:"domains"`
	Routes      map[string][]string `json:"routes,omitempty"`
}