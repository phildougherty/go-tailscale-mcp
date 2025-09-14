package tailscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIClient provides access to the Tailscale API
type APIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	tailnet    string
}

// NewAPIClient creates a new Tailscale API client
func NewAPIClient(apiKey string) (*APIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Extract tailnet from API key (format: tskey-api-xxxxx-xxx)
	// Or use the API to get the tailnet
	client := &APIClient{
		apiKey:  apiKey,
		baseURL: "https://api.tailscale.com/api/v2",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Get tailnet domain
	if err := client.fetchTailnet(); err != nil {
		return nil, fmt.Errorf("failed to fetch tailnet: %w", err)
	}

	return client, nil
}

// NewAPIClientWithTailnet creates a new Tailscale API client with explicit tailnet
func NewAPIClientWithTailnet(apiKey, tailnet string) (*APIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := &APIClient{
		apiKey:  apiKey,
		tailnet: tailnet,
		baseURL: "https://api.tailscale.com/api/v2",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return client, nil
}

// fetchTailnet gets the tailnet domain for the API key
func (c *APIClient) fetchTailnet() error {
	// Try to get devices to determine the tailnet
	// Since whoami endpoint doesn't exist, we'll try a test request
	// For personal accounts, the tailnet is typically the email address

	// First, try with a placeholder - we'll get the real one from the first successful API call
	// For now, we'll set a placeholder and update it when we make our first successful call
	c.tailnet = "-"  // Placeholder, will be determined from API responses

	// Try to list devices to validate the API key and get tailnet info
	testPath := "/tailnet/-/devices"
	resp, err := c.doRequest("GET", testPath, nil)
	if err != nil {
		// If this fails, we might need the user to provide the tailnet
		// For now, we'll continue and let individual API calls handle it
		return nil
	}
	defer resp.Body.Close()

	// Successfully connected, API key is valid
	return nil
}

// doRequest performs an HTTP request to the Tailscale API
func (c *APIClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	// Build full URL
	fullURL := c.baseURL + path
	if !strings.HasPrefix(path, "/") {
		fullURL = c.baseURL + "/" + path
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check for API errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// Device API Methods

// ListDevices lists all devices in the tailnet
func (c *APIClient) ListDevices() ([]Device, error) {
	tailnet := url.QueryEscape(c.tailnet)
	if c.tailnet == "-" || c.tailnet == "" {
		return nil, fmt.Errorf("tailnet not configured - set TAILSCALE_TAILNET environment variable")
	}

	path := fmt.Sprintf("/tailnet/%s/devices", tailnet)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Devices []Device `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Devices, nil
}

// GetDevice gets details for a specific device
func (c *APIClient) GetDevice(deviceID string) (*Device, error) {
	path := fmt.Sprintf("/device/%s", deviceID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var device Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, err
	}

	return &device, nil
}

// AuthorizeDevice authorizes a device
func (c *APIClient) AuthorizeDevice(deviceID string) error {
	path := fmt.Sprintf("/device/%s/authorized", deviceID)
	body := map[string]bool{"authorized": true}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// DeleteDevice removes a device from the tailnet
func (c *APIClient) DeleteDevice(deviceID string) error {
	path := fmt.Sprintf("/device/%s", deviceID)
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// SetDeviceTags sets tags for a device
func (c *APIClient) SetDeviceTags(deviceID string, tags []string) error {
	path := fmt.Sprintf("/device/%s/tags", deviceID)
	body := map[string][]string{"tags": tags}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// ACL/Policy API Methods

// GetACL gets the current ACL policy
func (c *APIClient) GetACL() (*ACL, error) {
	// Use URL encoding for email-based tailnets
	tailnet := url.QueryEscape(c.tailnet)
	if c.tailnet == "-" || c.tailnet == "" {
		// If tailnet is not set, return an error
		return nil, fmt.Errorf("tailnet not configured - set TAILSCALE_TAILNET environment variable")
	}

	path := fmt.Sprintf("/tailnet/%s/acl", tailnet)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The ACL endpoint returns HuJSON (with comments), not pure JSON
	// Read it as raw text for now
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACL response: %w", err)
	}

	// For now, return the raw ACL as a string in a simplified structure
	// A full implementation would parse HuJSON properly
	acl := &ACL{
		RawPolicy: string(bodyBytes),
	}

	return acl, nil
}

// SetACL updates the ACL policy
func (c *APIClient) SetACL(acl *ACL) error {
	tailnet := url.QueryEscape(c.tailnet)
	if c.tailnet == "-" || c.tailnet == "" {
		return fmt.Errorf("tailnet not configured - set TAILSCALE_TAILNET environment variable")
	}

	path := fmt.Sprintf("/tailnet/%s/acl", tailnet)

	// If we have raw policy, send that directly as HuJSON
	var body interface{}
	if acl.RawPolicy != "" {
		// Send raw HuJSON directly
		req, err := http.NewRequest("POST", c.baseURL+path, strings.NewReader(acl.RawPolicy))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/hujson")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
		}
		return nil
	} else {
		// Send structured ACL as JSON
		body = acl
	}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// ValidateACL validates an ACL policy without applying it
func (c *APIClient) ValidateACL(acl *ACL) error {
	tailnet := url.QueryEscape(c.tailnet)
	if c.tailnet == "-" || c.tailnet == "" {
		return fmt.Errorf("tailnet not configured - set TAILSCALE_TAILNET environment variable")
	}

	path := fmt.Sprintf("/tailnet/%s/acl/validate", tailnet)

	// If we have raw policy, validate that directly as HuJSON
	if acl.RawPolicy != "" {
		req, err := http.NewRequest("POST", c.baseURL+path, strings.NewReader(acl.RawPolicy))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/hujson")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
		}
		return nil
	}

	// Validate structured ACL
	resp, err := c.doRequest("POST", path, acl)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// Auth Key API Methods

// CreateAuthKey creates a new authentication key
func (c *APIClient) CreateAuthKey(options AuthKeyOptions) (*AuthKey, error) {
	path := fmt.Sprintf("/tailnet/%s/keys", c.tailnet)

	body := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"devices": map[string]interface{}{
				"create": map[string]interface{}{
					"reusable":      options.Reusable,
					"ephemeral":     options.Ephemeral,
					"preauthorized": options.Preauthorized,
					"tags":          options.Tags,
				},
			},
		},
		"expirySeconds": options.ExpirySeconds,
	}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var key AuthKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, err
	}

	return &key, nil
}

// ListAuthKeys lists all authentication keys
func (c *APIClient) ListAuthKeys() ([]AuthKey, error) {
	path := fmt.Sprintf("/tailnet/%s/keys", c.tailnet)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Keys []AuthKey `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// DeleteAuthKey deletes an authentication key
func (c *APIClient) DeleteAuthKey(keyID string) error {
	path := fmt.Sprintf("/tailnet/%s/keys/%s", c.tailnet, keyID)
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// DNS API Methods

// GetDNS gets the DNS configuration
func (c *APIClient) GetDNS() (*DNSConfig, error) {
	path := fmt.Sprintf("/tailnet/%s/dns/nameservers", c.tailnet)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dns DNSConfig
	if err := json.NewDecoder(resp.Body).Decode(&dns); err != nil {
		return nil, err
	}

	// Also get preferences for MagicDNS
	prefsPath := fmt.Sprintf("/tailnet/%s/dns/preferences", c.tailnet)
	prefsResp, err := c.doRequest("GET", prefsPath, nil)
	if err == nil {
		defer prefsResp.Body.Close()
		var prefs struct {
			MagicDNS bool `json:"magicDNS"`
		}
		if err := json.NewDecoder(prefsResp.Body).Decode(&prefs); err == nil {
			dns.MagicDNS = prefs.MagicDNS
		}
	}

	return &dns, nil
}

// SetDNSNameservers sets the DNS nameservers
func (c *APIClient) SetDNSNameservers(nameservers []string) error {
	path := fmt.Sprintf("/tailnet/%s/dns/nameservers", c.tailnet)
	body := map[string][]string{"dns": nameservers}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// SetDNSPreferences sets DNS preferences including MagicDNS
func (c *APIClient) SetDNSPreferences(magicDNS bool) error {
	path := fmt.Sprintf("/tailnet/%s/dns/preferences", c.tailnet)
	body := map[string]bool{"magicDNS": magicDNS}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// SetDNSSearchPaths sets the DNS search paths
func (c *APIClient) SetDNSSearchPaths(searchPaths []string) error {
	path := fmt.Sprintf("/tailnet/%s/dns/searchpaths", c.tailnet)
	body := map[string][]string{"searchPaths": searchPaths}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// Routes API Methods

// GetRoutes gets the advertised routes for a device
func (c *APIClient) GetRoutes(deviceID string) ([]string, error) {
	path := fmt.Sprintf("/device/%s/routes", deviceID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AdvertisedRoutes []string `json:"advertisedRoutes"`
		EnabledRoutes    []string `json:"enabledRoutes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.AdvertisedRoutes, nil
}

// SetRoutes sets the routes for a device
func (c *APIClient) SetRoutes(deviceID string, routes []string) error {
	path := fmt.Sprintf("/device/%s/routes", deviceID)
	body := map[string][]string{"routes": routes}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// ApproveRoutes approves routes for a device
func (c *APIClient) ApproveRoutes(deviceID string, routes []string) error {
	path := fmt.Sprintf("/device/%s/routes", deviceID)
	body := map[string][]string{"routes": routes}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// Helper function to check if API is available
func (c *APIClient) IsAvailable() bool {
	return c.apiKey != "" && c.tailnet != "" && c.tailnet != "-"
}

// getTailnetPath returns the URL-encoded tailnet for use in API paths
func (c *APIClient) getTailnetPath() (string, error) {
	if c.tailnet == "" || c.tailnet == "-" {
		return "", fmt.Errorf("tailnet not configured - set TAILSCALE_TAILNET environment variable")
	}
	return url.QueryEscape(c.tailnet), nil
}

// AuthKeyOptions defines options for creating an auth key
type AuthKeyOptions struct {
	Reusable      bool     `json:"reusable"`
	Ephemeral     bool     `json:"ephemeral"`
	Preauthorized bool     `json:"preauthorized"`
	Tags          []string `json:"tags,omitempty"`
	ExpirySeconds int      `json:"expirySeconds"`
}

// APIDevice represents a device from the API
type APIDevice struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Hostname      string    `json:"hostname"`
	User          string    `json:"user"`
	OS            string    `json:"os"`
	Addresses     []string  `json:"addresses"`
	Authorized    bool      `json:"authorized"`
	Tags          []string  `json:"tags"`
	KeyExpiryDisabled bool  `json:"keyExpiryDisabled"`
	LastSeen      time.Time `json:"lastSeen"`
	Created       time.Time `json:"created"`
	Expires       time.Time `json:"expires"`
	NodeKey       string    `json:"nodeKey"`
	MachineKey    string    `json:"machineKey"`
}