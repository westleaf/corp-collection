package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
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

	vnet, found, err := findVnet(ctx, *resourceGroupClientResponse.Name, "go-sdk-example-network", virtualNetworkClient)
	if !found {
		vnetPollerResponse, err := virtualNetworkClient.BeginCreateOrUpdate(ctx,
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
		vnet = vnetResp.VirtualNetwork
	}

	subnetsClient, err := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	subnetPollerResponse, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		*vnet.Name,
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

	// Public IP
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	publicIPPollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		"go-sdk-example-ip",
		armnetwork.PublicIPAddress{
			Location: to.Ptr(location),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			},
		},
		nil,
	)

	publicIPResponse, err := publicIPPollerResponse.PollUntilDone(ctx, nil)
	fmt.Printf("public ip %s created succefully\n", *publicIPResponse.Name)

	networkSecurityGroupClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	networkSecurityPollerResponse, err := networkSecurityGroupClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		"go-sdk-example-nsg",
		armnetwork.SecurityGroup{
			Location: to.Ptr(location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					{
						Name: to.Ptr("go-sdk-example-allow-ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("*"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("22"),
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Description:              to.Ptr("Allow ssh on port 22"),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr(int32(1001)),
						},
					},
				},
			},
		},
		nil,
	)

	networkSecurityResponse, err := networkSecurityPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("nsg %s created successfully\n", *networkSecurityResponse.Name)

	interfaceClient, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	nicPollerResponse, err := interfaceClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		"go-sdk-example-interfaceclient",
		armnetwork.Interface{
			Location: to.Ptr(location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				NetworkSecurityGroup: &armnetwork.SecurityGroup{
					ID: networkSecurityResponse.ID,
				},
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.Ptr("go-sdk-example-interfaceipconfig"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							Subnet: &armnetwork.Subnet{
								ID: subnetResponse.ID,
							},
							PublicIPAddress: &armnetwork.PublicIPAddress{
								ID: publicIPResponse.ID,
							},
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return err
	}

	nicResponse, err := nicPollerResponse.PollUntilDone(ctx, nil)

	fmt.Printf("nic %s created successfully\n", *nicResponse.Name)

	// Create VM
	fmt.Printf("Creating VM\n")

	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(location),
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     to.Ptr("ubuntu-24_04-lts"),
					Publisher: to.Ptr("Canonical"),
					SKU:       to.Ptr("server"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr("go-sdk-example-diskname"),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
					},
					DiskSizeGB: to.Ptr[int32](50),
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes("Standard_B1s")),
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  to.Ptr("go-sdk-example-compute"),
				AdminUsername: to.Ptr("go-sdk-example-admin"),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", "go-sdk-example-admin")),
								KeyData: &pubKey,
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: nicResponse.ID,
					},
				},
			},
		},
	}

	vmPollerResponse, err := vmClient.BeginCreateOrUpdate(
		ctx,
		*resourceGroupClientResponse.Name,
		"go-sdk-example-vm",
		parameters,
		nil,
	)
	if err != nil {
		return err
	}

	vmResponse, err := vmPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Printf("vm %s successfully created\n", *vmResponse.Name)

	return nil
}

func findVnet(ctx context.Context, resourceGroupName, vnetName string, vnetClient *armnetwork.VirtualNetworksClient) (armnetwork.VirtualNetwork, bool, error) {
	vnet, err := vnetClient.Get(ctx, resourceGroupName, vnetName, nil)
	if err != nil {
		var errResponse *azcore.ResponseError
		if errors.As(err, &errResponse) && errResponse.ErrorCode == "NotFound" {
			return vnet.VirtualNetwork, false, nil
		}
		return vnet.VirtualNetwork, false, err
	}

	return vnet.VirtualNetwork, true, nil
}
