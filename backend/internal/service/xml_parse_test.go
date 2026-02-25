package service

import (
	"testing"
)

const tldListXML = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors/>
  <Warnings/>
  <RequestedCommand>namecheap.domains.getTldList</RequestedCommand>
  <CommandResponse Type="namecheap.domains.getTldList">
    <Tlds>
      <Tld Name="com" NonRealTimeName="false" MinRegisterYears="1" MaxRegisterYears="10" MinRenewYears="1" MaxRenewYears="10" RenewalMinDays="0" RenewalMaxDays="365" ReactivateMaxDays="27" MinTransferYears="1" MaxTransferYears="1" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true" IsEppRequired="true" IsDisableModContact="false" IsDisableWGAllot="false" IsIncludeInExtendedSearchOnly="false" SequenceNumber="10" Type="gTLD" SubType="" IsSupportsIDN="true" Category="A" SupportsRegistrarLock="true" AddGracePeriodDays="5" WhoisVerification="false" ProviderApiDelete="false" TldState="" SearchGroup="" Registry=""/>
      <Tld Name="net" NonRealTimeName="false" MinRegisterYears="1" MaxRegisterYears="10" MinRenewYears="1" MaxRenewYears="10" RenewalMinDays="0" RenewalMaxDays="365" ReactivateMaxDays="27" MinTransferYears="1" MaxTransferYears="1" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true" IsEppRequired="true" IsDisableModContact="false" IsDisableWGAllot="false" IsIncludeInExtendedSearchOnly="false" SequenceNumber="20" Type="gTLD" SubType="" IsSupportsIDN="true" Category="A" SupportsRegistrarLock="true" AddGracePeriodDays="5" WhoisVerification="false" ProviderApiDelete="false" TldState="" SearchGroup="" Registry=""/>
      <Tld Name="org" NonRealTimeName="false" MinRegisterYears="1" MaxRegisterYears="10" IsApiRegisterable="false" IsApiRenewable="true" IsApiTransferable="true" SequenceNumber="30" Type="gTLD"/>
    </Tlds>
  </CommandResponse>
  <Server>SERVER1</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
</ApiResponse>`

const domainCheckXML = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors/>
  <Warnings/>
  <RequestedCommand>namecheap.domains.check</RequestedCommand>
  <CommandResponse Type="namecheap.domains.check">
    <DomainCheckResult Domain="example.com" Available="false" ErrorNo="0" Description="" IsPremiumName="false" PremiumRegistrationPrice="0" PremiumRenewalPrice="0" PremiumRestorePrice="0" PremiumTransferPrice="0" IcannFee="0" EapFee="0.0000"/>
    <DomainCheckResult Domain="example.net" Available="true" ErrorNo="0" Description="" IsPremiumName="false" PremiumRegistrationPrice="0" PremiumRenewalPrice="0" PremiumRestorePrice="0" PremiumTransferPrice="0" IcannFee="0.18" EapFee="0.0000"/>
    <DomainCheckResult Domain="example.xyz" Available="true" ErrorNo="0" Description="" IsPremiumName="true" PremiumRegistrationPrice="100.00" PremiumRenewalPrice="100.00" PremiumRestorePrice="100.00" PremiumTransferPrice="100.00" IcannFee="0.18" EapFee="0.0000"/>
  </CommandResponse>
</ApiResponse>`

const pricingXML = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors/>
  <Warnings/>
  <RequestedCommand>namecheap.users.getPricing</RequestedCommand>
  <CommandResponse Type="namecheap.users.getPricing">
    <UserGetPricingResult>
      <ProductType Name="DOMAIN">
        <ProductCategory Name="REGISTER">
          <Product Name="com">
            <Price Duration="1" DurationType="YEAR" Price="10.28" AdditionalCost="0.18" RegularPrice="13.98" RegularAdditionalCost="0.18" RegularAdditionalCostType="ICANNFEE" YourPrice="10.28" YourAdditionalCost="0.18" YourAdditionalCostType="ICANNFEE" PromotionPrice="0" Currency="USD"/>
            <Price Duration="2" DurationType="YEAR" Price="20.56" Currency="USD"/>
          </Product>
          <Product Name="net">
            <Price Duration="1" DurationType="YEAR" Price="12.28" AdditionalCost="0.18" RegularPrice="15.98" RegularAdditionalCost="0.18" Currency="USD"/>
          </Product>
          <Product Name="org">
            <Price Duration="1" DurationType="YEAR" Price="9.18" AdditionalCost="0.18" Currency="USD"/>
          </Product>
        </ProductCategory>
        <ProductCategory Name="RENEW">
          <Product Name="com">
            <Price Duration="1" DurationType="YEAR" Price="13.28" Currency="USD"/>
          </Product>
        </ProductCategory>
      </ProductType>
    </UserGetPricingResult>
  </CommandResponse>
</ApiResponse>`

const errorXML = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="ERROR" xmlns="http://api.namecheap.com/xml.response">
  <Errors>
    <Error Number="1011150">Parameter APIKey is invalid or API access has not been enabled</Error>
  </Errors>
  <Warnings/>
  <RequestedCommand>namecheap.domains.getTldList</RequestedCommand>
</ApiResponse>`

const errorEmptyXML = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="ERROR" xmlns="http://api.namecheap.com/xml.response">
  <Errors/>
  <Warnings/>
</ApiResponse>`

func TestGetStatus(t *testing.T) {
	root, err := ParseXML(tldListXML)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}
	status := GetStatus(root)
	if status != "OK" {
		t.Errorf("expected OK, got %s", status)
	}

	root2, _ := ParseXML(errorXML)
	status2 := GetStatus(root2)
	if status2 != "ERROR" {
		t.Errorf("expected ERROR, got %s", status2)
	}
}

func TestGetErrors(t *testing.T) {
	root, _ := ParseXML(errorXML)
	errs := GetErrors(root)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Number != "1011150" {
		t.Errorf("expected error number 1011150, got %s", errs[0].Number)
	}
	if errs[0].Text == "" {
		t.Error("expected non-empty error text")
	}
	t.Logf("Error: Number=%s Text=%s", errs[0].Number, errs[0].Text)
}

func TestGetErrorsEmpty(t *testing.T) {
	root, _ := ParseXML(errorEmptyXML)
	errs := GetErrors(root)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestGetErrorsOK(t *testing.T) {
	root, _ := ParseXML(tldListXML)
	errs := GetErrors(root)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for OK response, got %d", len(errs))
	}
}

func TestParseTldList(t *testing.T) {
	root, err := ParseXML(tldListXML)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}

	var tlds []map[string]string
	for _, tld := range root.FindElements(".//Tld") {
		attrs := make(map[string]string)
		for _, attr := range tld.Attr {
			attrs[attr.Key] = attr.Value
		}
		tlds = append(tlds, attrs)
	}

	if len(tlds) != 3 {
		t.Fatalf("expected 3 TLDs, got %d", len(tlds))
	}

	if tlds[0]["Name"] != "com" {
		t.Errorf("expected first TLD to be com, got %s", tlds[0]["Name"])
	}
	if tlds[1]["Name"] != "net" {
		t.Errorf("expected second TLD to be net, got %s", tlds[1]["Name"])
	}
	if tlds[2]["Name"] != "org" {
		t.Errorf("expected third TLD to be org, got %s", tlds[2]["Name"])
	}
	if tlds[0]["IsApiRegisterable"] != "true" {
		t.Errorf("expected com IsApiRegisterable=true, got %s", tlds[0]["IsApiRegisterable"])
	}
	if tlds[2]["IsApiRegisterable"] != "false" {
		t.Errorf("expected org IsApiRegisterable=false, got %s", tlds[2]["IsApiRegisterable"])
	}
	t.Logf("Parsed %d TLDs: %s, %s, %s", len(tlds), tlds[0]["Name"], tlds[1]["Name"], tlds[2]["Name"])
}

func TestParseDomainCheck(t *testing.T) {
	root, err := ParseXML(domainCheckXML)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}

	var results []map[string]string
	for _, r := range root.FindElements(".//DomainCheckResult") {
		attrs := make(map[string]string)
		for _, attr := range r.Attr {
			attrs[attr.Key] = attr.Value
		}
		results = append(results, attrs)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0]["Domain"] != "example.com" {
		t.Errorf("expected example.com, got %s", results[0]["Domain"])
	}
	if results[0]["Available"] != "false" {
		t.Errorf("expected example.com Available=false, got %s", results[0]["Available"])
	}
	if results[1]["Domain"] != "example.net" {
		t.Errorf("expected example.net, got %s", results[1]["Domain"])
	}
	if results[1]["Available"] != "true" {
		t.Errorf("expected example.net Available=true, got %s", results[1]["Available"])
	}
	if results[2]["IsPremiumName"] != "true" {
		t.Errorf("expected example.xyz IsPremiumName=true, got %s", results[2]["IsPremiumName"])
	}
	t.Logf("Parsed %d domain check results", len(results))
}

func TestParsePricing(t *testing.T) {
	root, err := ParseXML(pricingXML)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}

	pricing := make(map[string]map[string]string)

	for _, cat := range root.FindElements(".//ProductCategory") {
		catName := cat.SelectAttrValue("Name", "")
		t.Logf("Found ProductCategory: %s", catName)

		if catName != "REGISTER" {
			continue
		}

		for _, prod := range cat.FindElements("Product") {
			tld := prod.SelectAttrValue("Name", "")
			t.Logf("  Found Product: %s", tld)

			for _, price := range prod.FindElements("Price") {
				duration := price.SelectAttrValue("Duration", "")
				durationType := price.SelectAttrValue("DurationType", "")
				priceVal := price.SelectAttrValue("Price", "")
				t.Logf("    Found Price: Duration=%s DurationType=%s Price=%s", duration, durationType, priceVal)

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

	if len(pricing) != 3 {
		t.Fatalf("expected 3 pricing entries (com, net, org), got %d", len(pricing))
	}

	if pricing["com"]["Price"] != "10.28" {
		t.Errorf("expected com price 10.28, got %s", pricing["com"]["Price"])
	}
	if pricing["com"]["Currency"] != "USD" {
		t.Errorf("expected com currency USD, got %s", pricing["com"]["Currency"])
	}
	if pricing["net"]["Price"] != "12.28" {
		t.Errorf("expected net price 12.28, got %s", pricing["net"]["Price"])
	}
	if pricing["org"]["Price"] != "9.18" {
		t.Errorf("expected org price 9.18, got %s", pricing["org"]["Price"])
	}

	t.Logf("Parsed pricing for %d TLDs", len(pricing))
	for tld, info := range pricing {
		t.Logf("  %s: Price=%s Currency=%s", tld, info["Price"], info["Currency"])
	}
}
