/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scope

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

const (
	// ChinaCloud is the cloud environment operated in China
	ChinaCloud = "AzureChinaCloud"
	// GermanCloud is the cloud environment operated in Germany
	GermanCloud = "AzureGermanCloud"
	// PublicCloud is the default public Azure cloud environment
	PublicCloud = "AzurePublicCloud"
	// USGovernmentCloud is the cloud environment for the US Government
	USGovernmentCloud = "AzureUSGovernmentCloud"
)

// AzureClients contains all the Azure clients used by the scopes.
type AzureClients struct {
	Authorizer                 autorest.Authorizer
	environment                string
	ResourceManagerEndpoint    string
	ResourceManagerVMDNSSuffix string
	subscriptionID             string
	tenantID                   string
	clientID                   string
	clientSecret               string
}

// CloudEnvironment returns the Azure environment the controller runs in.
func (c *AzureClients) CloudEnvironment() string {
	return c.environment
}

// SubscriptionID returns the Azure subscription id from the controller environment
func (c *AzureClients) SubscriptionID() string {
	return c.subscriptionID
}

// TenantID returns the Azure tenant id the controller runs in.
func (c *AzureClients) TenantID() string {
	return c.tenantID
}

// ClientID returns the Azure client id from the controller environment
func (c *AzureClients) ClientID() string {
	return c.clientID
}

// ClientSecret returns the Azure client secret from the controller environment
func (c *AzureClients) ClientSecret() string {
	return c.clientSecret
}

func (c *AzureClients) setCredentials(subscriptionID string) error {
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return err
	}

	if subscriptionID == "" {
		subscriptionID = settings.GetSubscriptionID()
		if subscriptionID == "" {
			return fmt.Errorf("error creating azure services. subscriptionID is not set in cluster or AZURE_SUBSCRIPTION_ID env var")
		}
	}

	c.subscriptionID = subscriptionID
	c.tenantID = strings.TrimSuffix(settings.Values[auth.TenantID], "\n")
	c.clientID = strings.TrimSuffix(settings.Values[auth.ClientID], "\n")
	c.clientSecret = strings.TrimSuffix(settings.Values[auth.ClientSecret], "\n")

	c.environment = settings.Values[auth.EnvironmentName]
	if c.environment == "" {
		c.environment = azure.PublicCloud.Name
	}

	c.ResourceManagerEndpoint = settings.Environment.ResourceManagerEndpoint
	c.ResourceManagerVMDNSSuffix = GetAzureDNSZoneForEnvironment(settings.Environment.Name)
	settings.Values[auth.SubscriptionID] = subscriptionID
	settings.Values[auth.TenantID] = c.tenantID

	c.Authorizer, err = settings.GetAuthorizer()
	return err
}

// GetAzureDNSZoneForEnvironment returnes the DNSZone to be used with the
// cloud environment, the default is the public cloud
func GetAzureDNSZoneForEnvironment(environmentName string) string {
	// default is public cloud
	switch environmentName {
	case ChinaCloud:
		return "cloudapp.chinacloudapi.cn"
	case GermanCloud:
		return "cloudapp.microsoftazure.de"
	case PublicCloud:
		return "cloudapp.azure.com"
	case USGovernmentCloud:
		return "cloudapp.usgovcloudapi.net"
	default:
		return "cloudapp.azure.com"
	}
}
