package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
)

// localXMLTag strips namespace prefix from etree element tags.
func localXMLTag(tag string) string {
	if idx := strings.LastIndex(tag, "}"); idx >= 0 {
		return tag[idx+1:]
	}
	return tag
}

// NamecheapClient is the client for Namecheap API
type NamecheapClient struct {
	cfg     *model.NamecheapConfig
	limiter *RateLimiter
	client  *http.Client
}

// NewNamecheapClient creates a new Namecheap API client
func NewNamecheapClient(cfg *model.NamecheapConfig, ratePerMin float64, timeoutSec float64) *NamecheapClient {
	if timeoutSec <= 0 {
		timeoutSec = 60.0
	}
	return &NamecheapClient{
		cfg:     cfg,
		limiter: NewRateLimiter(ratePerMin),
		client: &http.Client{
			Timeout: time.Duration(timeoutSec * float64(time.Second)),
		},
	}
}

// Close closes the client
func (c *NamecheapClient) Close() {
	// Nothing to close for now
}

// call makes an API call with retry logic
func (c *NamecheapClient) call(ctx context.Context, command string, params map[string]string) (string, error) {
	var result string
	err := Retry(ctx, func() error {
		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}

		// Build query parameters
		q := url.Values{}
		q.Set("ApiUser", c.cfg.ApiUser)
		q.Set("ApiKey", c.cfg.ApiKey)
		q.Set("UserName", c.cfg.Username)
		q.Set("ClientIp", c.cfg.ClientIp)
		q.Set("Command", command)
		for k, v := range params {
			q.Set(k, v)
		}

		// Make request
		reqURL := c.cfg.ApiBase + "?" + q.Encode()
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return err
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Check HTTP status
		if resp.StatusCode == 429 {
			return &RetriableError{Message: "HTTP 429 Too Many Requests"}
		}
		if resp.StatusCode >= 500 {
			return &RetriableError{Message: fmt.Sprintf("HTTP %d server error", resp.StatusCode)}
		}
		if resp.StatusCode >= 400 {
			return &model.ApiCallError{Message: "HTTP error", HTTPStatus: resp.StatusCode}
		}

		// Parse XML and check API status
		root, err := ParseXML(string(body))
		if err != nil {
			return err
		}

		status := GetStatus(root)
		errs := GetErrors(root)

		if status != "OK" || len(errs) > 0 {
			var msg string
			if len(errs) > 0 {
				var texts []string
				for _, e := range errs {
					if e.Text != "" {
						texts = append(texts, e.Text)
					}
				}
				msg = strings.Join(texts, " | ")
			}
			if msg == "" {
				msg = fmt.Sprintf("API Status=%s", status)
			}

			if IsRateLimitError(msg) {
				return &RetriableError{Message: msg}
			}
			return &model.ApiCallError{Message: msg, HTTPStatus: resp.StatusCode, Errors: errs}
		}

		result = string(body)
		return nil
	}, IsRetriableError, nil, nil)

	return result, err
}

// GetTldList retrieves the list of all TLDs
func (c *NamecheapClient) GetTldList(ctx context.Context) ([]map[string]string, error) {
	xml, err := c.call(ctx, "namecheap.domains.getTldList", nil)
	if err != nil {
		return nil, err
	}

	root, err := ParseXML(xml)
	if err != nil {
		return nil, err
	}

	var tlds []map[string]string

	for _, tld := range root.FindElements(".//Tld") {
		attrs := make(map[string]string)
		for _, attr := range tld.Attr {
			attrs[attr.Key] = attr.Value
		}
		tlds = append(tlds, attrs)
	}

	return tlds, nil
}

// DomainsCheck checks the availability of multiple domains (max 50 per call)
func (c *NamecheapClient) DomainsCheck(ctx context.Context, domains []string) ([]model.DomainCheckResult, error) {
	if len(domains) == 0 {
		return nil, nil
	}
	if len(domains) > 50 {
		return nil, fmt.Errorf("domains.check supports max 50 domains per call")
	}

	xml, err := c.call(ctx, "namecheap.domains.check", map[string]string{
		"DomainList": strings.Join(domains, ","),
	})
	if err != nil {
		return nil, err
	}

	root, err := ParseXML(xml)
	if err != nil {
		return nil, err
	}

	var results []model.DomainCheckResult

	for _, r := range root.FindElements(".//DomainCheckResult") {
		result := model.DomainCheckResult{}
		if domain := r.SelectAttrValue("Domain", ""); domain != "" {
			result.Domain = domain
		}
		result.Available = r.SelectAttrValue("Available", "")
		result.IsPremiumName = r.SelectAttrValue("IsPremiumName", "")
		result.PremiumRegistrationPrice = r.SelectAttrValue("PremiumRegistrationPrice", "")
		result.IcannFee = r.SelectAttrValue("IcannFee", "")
		result.EapFee = r.SelectAttrValue("EapFee", "")
		results = append(results, result)
	}

	return results, nil
}

// GetPricingRegister1y retrieves 1-year registration pricing for all TLDs
func (c *NamecheapClient) GetPricingRegister1y(ctx context.Context) (map[string]map[string]string, error) {
	tryParams := []map[string]string{
		{"ProductType": "DOMAIN", "ProductCategory": "DOMAINS", "ActionName": "REGISTER"},
		{"ProductType": "DOMAIN"},
	}

	for _, params := range tryParams {
		xml, err := c.call(ctx, "namecheap.users.getPricing", params)
		if err != nil {
			return nil, err
		}

		root, err := ParseXML(xml)
		if err != nil {
			return nil, err
		}

		pricing := make(map[string]map[string]string)

		for _, cat := range root.FindElements(".//ProductCategory") {
			catName := strings.ToUpper(cat.SelectAttrValue("Name", ""))
			if catName != "REGISTER" {
				continue
			}

			for _, prod := range cat.FindElements("Product") {
				tld := strings.ToLower(prod.SelectAttrValue("Name", ""))
				if tld == "" {
					continue
				}

				for _, price := range prod.FindElements("Price") {
					duration := price.SelectAttrValue("Duration", "")
					durationType := strings.ToUpper(price.SelectAttrValue("DurationType", ""))

					if duration == "1" && durationType == "YEAR" {
						attrs := make(map[string]string)
						for _, attr := range price.Attr {
							attrs[attr.Key] = attr.Value
						}
						pricing[tld] = attrs
						break
					}
				}
			}
		}

		if len(pricing) > 0 {
			return pricing, nil
		}
	}

	return make(map[string]map[string]string), nil
}
