//go:generate go run generate/protocol.go

package ga

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var trackingIDMatcher = regexp.MustCompile(`^UA-\d+-\d+$`)

func NewClient(trackingID string) (*Client, error) {
	if !trackingIDMatcher.MatchString(trackingID) {
		return nil, fmt.Errorf("Invalid Tracking ID: %s", trackingID)
	}
	c := &Client{}
	c.UseTLS = true
	c.ProtocolVersion = "1"
	c.ClientID = "go-ga"
	c.TrackingID = trackingID
	return c, nil
}

type hitType interface {
	addFields(url.Values) error
}

func (c *Client) Send(h hitType) error {
	v := url.Values{}

	c.setType(h)

	err := c.addFields(v)
	if err != nil {
		return err
	}

	err = h.addFields(v)
	if err != nil {
		return err
	}

	url := ""
	if c.UseTLS {
		url = "http://www.google-analytics.com/collect"
	} else {
		url = "https://ssl.google-analytics.com/collect"
	}

	str := v.Encode()
	buf := bytes.NewBufferString(str)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", buf)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Rejected by Google with code %d", resp.StatusCode)
	}

	// fmt.Printf("POST %s => %d\n", str, resp.StatusCode)

	return nil
}
