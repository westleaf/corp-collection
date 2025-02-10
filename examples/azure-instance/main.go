package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	sshkeygen "github.com/westleaf/corp-collection/examples/ssh-keygen"
)

const location = "swedencentral"

func main() {
	var (
		token  azcore.TokenCredential
		pubKey string
		err    error
	)

	ctx := context.Background()
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		fmt.Printf("no subscription ID was provided, set the AZURE_SUBSCRIPTION_ID environment variable")
		os.Exit(1)
	}

	if pubKey, err = generateKeys(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if token, err = getToken(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = createInstance(ctx, subscriptionID, pubKey, token); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func generateKeys() (string, error) {
	var (
		privateKey []byte
		publicKey  []byte
		err        error
	)
	if privateKey, publicKey, err = sshkeygen.GenerateKeys(); err != nil {
		return "", fmt.Errorf("failed to generate keys: %w", err)
	}
	if err = os.WriteFile("secretkey.pem", privateKey, 0600); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = os.WriteFile("secretkey.pub", publicKey, 0644); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	return string(publicKey), nil
}

func getToken() (azcore.TokenCredential, error) {
	token, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return token, fmt.Errorf("failed to create Azure CLI credential: %w", err)
	}
	return token, nil
}

func createInstance(ctx context.Context, subscriptionID, pubKey string, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	resourceGroupParams := armresources.ResourceGroup{
		Location: to.Ptr(location),
	}

	resourceGroupClientResponse, err := resourceGroupClient.CreateOrUpdate(ctx, "go-sdk-example", resourceGroupParams, nil)
	if err != nil {
		return err
	}
	fmt.Printf("ResourceGroupClientResponse status is: %s\n", *resourceGroupClientResponse.Properties.ProvisioningState)

	// Create VNET
	virtualNetworkClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	// vnetGroupParams :=

	vnetPollerResponse, err := virtualNetworkClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		"go-sdk-example-network",
		armnetwork.VirtualNetwork{
			Location: to.Ptr(location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.Ptr("10.1.0.0/16"),
					},
				},
			},
		},
		nil,
	)

	if err != nil {
		return err
	}

	vnetResp, err := vnetPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Printf("vnet %s created successfully\n", *vnetResp.Name)

	// Create subnet
	subnetsClient, err := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	subnetPollerResponse, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		*vnetResp.Name,
		"go-sdx-example-subnet",
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.Ptr("10.1.0.0/24"),
			},
		},
		nil,
	)

	if err != nil {
		return err
	}

	subnetResponse, err := subnetPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("subnet %s created successfully\n", *subnetResponse.Name)

	return nil
}
