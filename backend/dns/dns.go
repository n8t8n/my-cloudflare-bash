package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

type CreateDNSRequest struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	Proxied   bool   `json:"proxied"`
}

type CloudflareResponse struct {
	Success bool        `json:"success"`
	Errors  []string    `json:"errors"`
	Result  interface{} `json:"result"`
}

type CloudflareListResponse struct {
	Success bool        `json:"success"`
	Errors  []string    `json:"errors"`
	Result  []DNSRecord `json:"result"`
}

func getCloudflareAPI() (string, string, string) {
	apiToken := os.Getenv("CF_API_TOKEN")
	zoneID := os.Getenv("CF_ZONE_ID")
	domain := os.Getenv("CF_DOMAIN")
	return apiToken, zoneID, domain
}

func ListDNSRecords() ([]DNSRecord, error) {
	apiToken, zoneID, _ := getCloudflareAPI()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100", zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfResp CloudflareListResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, err
	}

	if !cfResp.Success {
		return nil, fmt.Errorf("cloudflare API error: %v", cfResp.Errors)
	}

	return cfResp.Result, nil
}

func CreateDNSRecord(req CreateDNSRequest) (*DNSRecord, error) {
	apiToken, zoneID, domain := getCloudflareAPI()

	fullName := fmt.Sprintf("%s.%s", req.Subdomain, domain)

	// Check if record already exists
	records, err := ListDNSRecords()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.Name == fullName {
			return nil, fmt.Errorf("record already exists for %s", fullName)
		}
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)

	payload := map[string]interface{}{
		"type":    req.Type,
		"name":    fullName,
		"content": req.Target,
		"ttl":     1,
		"proxied": req.Proxied,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, err
	}

	if !cfResp.Success {
		return nil, fmt.Errorf("cloudflare API error: %v", cfResp.Errors)
	}

	// Parse the result back to DNSRecord
	resultBytes, _ := json.Marshal(cfResp.Result)
	var record DNSRecord
	json.Unmarshal(resultBytes, &record)

	return &record, nil
}

func DeleteDNSRecord(recordID string) error {
	apiToken, zoneID, _ := getCloudflareAPI()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return err
	}

	if !cfResp.Success {
		return fmt.Errorf("cloudflare API error: %v", cfResp.Errors)
	}

	return nil
}

type UpdateDNSRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func UpdateDNSRecord(recordID string, req UpdateDNSRequest) (*DNSRecord, error) {
	apiToken, zoneID, _ := getCloudflareAPI()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID)

	payload := map[string]interface{}{
		"type":    req.Type,
		"content": req.Content,
		"ttl":     req.TTL,
		"proxied": req.Proxied,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, err
	}

	if !cfResp.Success {
		return nil, fmt.Errorf("cloudflare API error: %v", cfResp.Errors)
	}

	// Parse the result back to DNSRecord
	resultBytes, _ := json.Marshal(cfResp.Result)
	var record DNSRecord
	json.Unmarshal(resultBytes, &record)

	return &record, nil
}
