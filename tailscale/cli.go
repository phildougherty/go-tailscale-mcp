package tailscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CLI wraps the Tailscale CLI commands
type CLI struct {
	binaryPath string
}

// NewCLI creates a new Tailscale CLI wrapper
func NewCLI() *CLI {
	return &CLI{
		binaryPath: "tailscale",
	}
}

// Execute runs a Tailscale CLI command and returns the output
func (c *CLI) Execute(args ...string) (string, error) {
	cmd := exec.Command(c.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ExecuteJSON runs a Tailscale CLI command and parses JSON output
func (c *CLI) ExecuteJSON(v interface{}, args ...string) error {
	// Add --json flag if not present
	hasJSON := false
	for _, arg := range args {
		if arg == "--json" || arg == "-json" {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		args = append(args, "--json")
	}

	output, err := c.Execute(args...)
	if err != nil {
		return err
	}

	if output == "" {
		return fmt.Errorf("empty response from tailscale")
	}

	return json.Unmarshal([]byte(output), v)
}

// Status returns the current Tailscale status
func (c *CLI) Status() (*Status, error) {
	var status Status
	err := c.ExecuteJSON(&status, "status")
	return &status, err
}

// Login connects to Tailscale
func (c *CLI) Login(authKey string, options map[string]string) error {
	args := []string{"up"}

	if authKey != "" {
		args = append(args, "--authkey", authKey)
	}

	for key, value := range options {
		args = append(args, fmt.Sprintf("--%s", key), value)
	}

	_, err := c.Execute(args...)
	return err
}

// Logout disconnects from Tailscale
func (c *CLI) Logout() error {
	_, err := c.Execute("logout")
	return err
}

// Down disconnects from the network but stays logged in
func (c *CLI) Down() error {
	_, err := c.Execute("down")
	return err
}

// SwitchProfile switches to a different Tailscale profile
func (c *CLI) SwitchProfile(profile string) error {
	_, err := c.Execute("switch", profile)
	return err
}

// ListProfiles lists all available profiles
func (c *CLI) ListProfiles() ([]Profile, error) {
	output, err := c.Execute("switch", "--list")
	if err != nil {
		return nil, err
	}

	// Parse the table output
	// Format: ID    Tailnet                   Account
	//         826b  phil.dougherty@gmail.com  phil.dougherty@gmail.com*
	profiles := []Profile{}
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		// Skip header line and empty lines
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and reconstruct fields
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Extract ID and tailnet
		id := fields[0]
		tailnet := fields[1]

		// Extract account and check if it's active (marked with *)
		account := fields[2]
		active := false
		if strings.HasSuffix(account, "*") {
			active = true
			account = strings.TrimSuffix(account, "*")
		}

		profiles = append(profiles, Profile{
			ID:      id,
			Tailnet: tailnet,
			Account: account,
			Active:  active,
		})
	}

	return profiles, nil
}

// Ping pings a peer device
func (c *CLI) Ping(target string, count int) (string, error) {
	args := []string{"ping", target}
	if count > 0 {
		args = append(args, "-c", fmt.Sprintf("%d", count))
	}
	return c.Execute(args...)
}

// Version returns Tailscale version information
func (c *CLI) Version() (string, error) {
	return c.Execute("version")
}

// IP returns the Tailscale IP addresses
func (c *CLI) IP(device string) (string, error) {
	args := []string{"ip"}
	if device != "" {
		args = append(args, device)
	}
	return c.Execute(args...)
}

// SetExitNode sets the exit node
func (c *CLI) SetExitNode(node string) error {
	_, err := c.Execute("set", "--exit-node", node)
	return err
}

// ClearExitNode clears the exit node
func (c *CLI) ClearExitNode() error {
	_, err := c.Execute("set", "--exit-node=")
	return err
}

// AdvertiseRoutes advertises routes
func (c *CLI) AdvertiseRoutes(routes []string) error {
	if len(routes) == 0 {
		return fmt.Errorf("no routes specified")
	}
	_, err := c.Execute("set", "--advertise-routes", strings.Join(routes, ","))
	return err
}

// AcceptRoutes enables accepting routes from peers
func (c *CLI) AcceptRoutes(accept bool) error {
	value := "false"
	if accept {
		value = "true"
	}
	_, err := c.Execute("set", "--accept-routes", value)
	return err
}

// LoginNewProfile logs in with a new profile
func (c *CLI) LoginNewProfile() (string, error) {
	// This will start the login process and return the auth URL
	output, err := c.Execute("login")
	return output, err
}