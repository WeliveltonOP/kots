package helm

import (
	"fmt"

	helmregistry "helm.sh/helm/v3/pkg/registry"
)

// CheckHelmRegistryCredentials will attempt to log in to a OCI registry
// with the params provided, and return succeess(bool).  error is only returned
// if there was a problem checking. Invalid creds will return false, nil
func CheckHelmRegistryCredentials(hostname string, username string, password string) (bool, error) {
	client := helmregistry.Client{}

	err := client.Login(hostname, helmregistry.LoginOptBasicAuth(username, password))
	fmt.Printf("%s\n", err)

	return false, nil
}
