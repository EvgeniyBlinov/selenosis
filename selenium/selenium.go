package selenium

import (
	"strings"

	"github.com/imdario/mergo"
)

// Capabilities ...
// Moon Reserved Label Names Key - Meaning (moon:options)
// app              - Stores unique name for every pod
// browserName      - Stores browser name
// browserVersion   - Stores browser version
// enableVNC        - Stores whether VNC is enabled
// moon             - System label, always equal to browsers
// quota            - Stores user quota name
// screenResolution - Stores screen resolution requested by user

type Capabilities struct {
	BrowserName           string            `json:"browserName,omitempty"`
	DeviceName            string            `json:"deviceName,omitempty"`
	W3CDeviceName         string            `json:"appium:deviceName,omitempty"`
	BrowserVersion        string            `json:"version,omitempty"`
	W3CBrowserVersion     string            `json:"browserVersion,omitempty"`
	Platform              string            `json:"platform,omitempty"`
	WC3PlatformName       string            `json:"platformName,omitempty"`
	ScreenResolution      string            `json:"screenResolution,omitempty"`
	Skin                  string            `json:"skin,omitempty"`
	VNC                   bool              `json:"enableVNC,omitempty"`
	Video                 bool              `json:"enableVideo,omitempty"`
	Log                   bool              `json:"enableLog,omitempty"`
	VideoName             string            `json:"videoName,omitempty"`
	VideoScreenSize       string            `json:"videoScreenSize,omitempty"`
	VideoFrameRate        uint16            `json:"videoFrameRate,omitempty"`
	VideoCodec            string            `json:"videoCodec,omitempty"`
	LogName               string            `json:"logName,omitempty"`
	TestName              string            `json:"name,omitempty"`
	TimeZone              string            `json:"timeZone,omitempty"`
	ContainerHostname     string            `json:"containerHostname,omitempty"`
	Env                   []string          `json:"env,omitempty"`
	ApplicationContainers []string          `json:"applicationContainers,omitempty"`
	AdditionalNetworks    []string          `json:"additionalNetworks,omitempty"`
	HostsEntries          []string          `json:"hostsEntries,omitempty"`
	DNSServers            []string          `json:"dnsServers,omitempty"`
	Labels                map[string]string `json:"labels,omitempty"`
	SessionTimeout        string            `json:"sessionTimeout,omitempty"`
	SelenoidCapabilities  *Capabilities     `json:"selenoid:options,omitempty"`
	MoonCapabilities      *Capabilities     `json:"moon:options,omitempty"`
}

//ValidateCapabilities ...
func (c *Capabilities) ValidateCapabilities() {
	if c.W3CBrowserVersion != "" {
		c.BrowserVersion = c.W3CBrowserVersion
	}

	if c.WC3PlatformName != "" {
		c.Platform = c.WC3PlatformName
	}

	if c.W3CDeviceName != "" {
		c.DeviceName = c.W3CDeviceName
	}

	if c.SelenoidCapabilities != nil {
		mergo.Merge(c, *c.SelenoidCapabilities, mergo.WithOverride) //We probably need to handle returned error
	}

	if c.MoonCapabilities != nil {
		mergo.Merge(c, *c.MoonCapabilities, mergo.WithOverride) //We probably need to handle returned error
	}

}

//GetBrowserName ...
func (c *Capabilities) GetBrowserName() string {
	browserName := c.BrowserName
	if len(strings.TrimSpace(browserName)) != 0 {
		return browserName
	}
	if c.DeviceName != "" {
		return c.DeviceName
	}
	return c.W3CDeviceName
}
