package pull

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Pull(t *testing.T) {
	data := `apiVersion: kots.io/v1beta1
kind: License
metadata:
  name: expiredtestlicense
spec:
  licenseID: VJDJXAPDAStK62ijnUnIC1zJOW0A2t7z
  licenseType: trial
  customerName: ExpiredTestLicense
  appSlug: testkotsapp
  channelName: Unstable
  licenseSequence: 1
  endpoint: 'http://replicated-app:3000'
  entitlements:
    expires_at:
      title: Expiration
      description: License Expiration
      value: '2019-02-06T08:00:00Z'
      valueType: String
  signature: >-
    eyJsaWNlbnNlRGF0YSI6ImV5SmhjR2xXWlhKemFXOXVJam9pYTI5MGN5NXBieTkyTVdKbGRHRXhJaXdpYTJsdVpDSTZJa3hwWTJWdWMyVWlMQ0p0WlhSaFpHRjBZU0k2ZXlKdVlXMWxJam9pWlhod2FYSmxaSFJsYzNSc2FXTmxibk5sSW4wc0luTndaV01pT25zaWJHbGpaVzV6WlVsRUlqb2lWa3BFU2xoQlVFUkJVM1JMTmpKcGFtNVZia2xETVhwS1QxY3dRVEowTjNvaUxDSnNhV05sYm5ObFZIbHdaU0k2SW5SeWFXRnNJaXdpWTNWemRHOXRaWEpPWVcxbElqb2lSWGh3YVhKbFpGUmxjM1JNYVdObGJuTmxJaXdpWVhCd1UyeDFaeUk2SW5SbGMzUnJiM1J6WVhCd0lpd2lZMmhoYm01bGJFNWhiV1VpT2lKVmJuTjBZV0pzWlNJc0lteHBZMlZ1YzJWVFpYRjFaVzVqWlNJNk1Td2laVzVrY0c5cGJuUWlPaUpvZEhSd09pOHZjbVZ3YkdsallYUmxaQzFoY0hBNk16QXdNQ0lzSW1WdWRHbDBiR1Z0Wlc1MGN5STZleUpsZUhCcGNtVnpYMkYwSWpwN0luUnBkR3hsSWpvaVJYaHdhWEpoZEdsdmJpSXNJbVJsYzJOeWFYQjBhVzl1SWpvaVRHbGpaVzV6WlNCRmVIQnBjbUYwYVc5dUlpd2lkbUZzZFdVaU9pSXlNREU1TFRBeUxUQTJWREE0T2pBd09qQXdXaUlzSW5aaGJIVmxWSGx3WlNJNklsTjBjbWx1WnlKOWZYMTkiLCJpbm5lclNpZ25hdHVyZSI6ImV5SnNhV05sYm5ObFUybG5ibUYwZFhKbElqb2lXRU16TDFwUk1XWmlhMlp2WlRkNWNEQjRZemhLYjFseFREQm1PRkZTY2tkeUsweHlRVFpyTDJwdWVGZGlPVE56ZWxoQlNrWldlVWh1ZWpSamVWRlJNRGRtVDFjemFIaE1SbXhPZW1rd1pHcDJUa2RxV1hocVpFNTZaMkV5U1VzdmJEQjZla2d5Um1GeFNFRllLM056VkRSa2FFSklURlY2TmxGVU9IaGpka1JsZDNNM1ZYYzRlV0pqYVdOalFVSnJXUzk2V201M1pXRk5aSGxCYTJaRFVWUnJkSFY2T0hOak5rWXZZbWxXYVhGeGNuSmlOamhuVUdnMldFNU9SRk56YTBSeVptWkhXREJTTTFsTlFYTldkbGw0V2tOWFNtWlZhSGM0TjBaQlZWaFpVbWh4V21SV2IzaFFkRkZtWm1ocVdEaFZNVW8xYUVaalNtTlBjVGQ2VWxZME9VWXdNV2t2VUhodVV6UjNkVE5MUjFkYWNVWjZhbFJNWlVsSU9UaFVWeXRPYkhObmFYZ3hXWE5qV0dZelNtNUtZVzg1V0hCRU5XUnNha3N4ZEZsUGJITllSR0pyTTNjd2VrOTFOamhEVUZORE1ESkJQVDBpTENKd2RXSnNhV05MWlhraU9pSXRMUzB0TFVKRlIwbE9JRkJWUWt4SlF5QkxSVmt0TFMwdExWeHVUVWxKUWtscVFVNUNaMnR4YUd0cFJ6bDNNRUpCVVVWR1FVRlBRMEZST0VGTlNVbENRMmRMUTBGUlJVRjZkRUpDWjBkR1IxRkpTbWRvYUM5cGFFRnphRnh1UjFZeFJtbHRVMHRQZDJ0TFpHdHVZVWxKUVVOamFGUXJXVXd4UzFjeVZUbFhUamsyVTBzNVdIWjNWblZvVWxsbFlrSjFjRk0xT1RaQ1pFNXplVmRFYWx4dVJpOUVWVEpWV21sbVIycElNM0I2ZEdKdFQzSlFLMnBWWlRsUE0ydFdNVmd5Tnl0YVowaDBha3RPT0dwVFZrSmxSemwyTkZvd1ZGTXplR1EwZDFWSlpWeHVlVzlhYWs1TVdrUjVZVGRMVW5wcFNsWndLMWM0TkUweVNIZEZaamxwSzJseFZuWm1ZVEI0YUhwbFJFTTRWRGw2UmxWNFRFeERZa1Y0YVVOdEsybzVWRnh1VDNaeWFqWmphelpRZG1Zd1FYcHhRazlyWmxKdlFYbEVPWEZPUVM4NFRUQnVUR04xVTFkUWIwcDRja1pHVnpZelYwWnJZazVoT1VSVkwxQnNSVTFTZDF4dVEydFJTWFozS3poSWIydzFUUzlZZGtaM1VVNVZiM2REVnk5elJXeE9ORFkwZDBwNFVuTklUVk4xVkVkU2RVVTBjbGgyUWxkQk9FUlhjSEI1UWtwMmQxeHVNbEZKUkVGUlFVSmNiaTB0TFMwdFJVNUVJRkJWUWt4SlF5QkxSVmt0TFMwdExWeHVJaXdpYTJWNVUybG5ibUYwZFhKbElqb2laWGxLZW1GWFpIVlpXRkl4WTIxVmFVOXBTbXRVVm14dFZVaFNXbVJWVWxwWk0wWnlWbFJHVkZwRVFsSlNha0phVld4YWVFMHlPREZWYTBwdFVUQkdVbFpxUW05VmJUbHlXbXQwZDFkRVRtdFZWMVUwWlZOMGVHRldaSFJsYms0eVVqSjRiRTFyVm01WFJ6VlBXbGhhVVU1Rk1ESmxiVGxXWTFST1NGcEZWbTlaTW1RelQxVXhSazB5YUVKVmFteHZUa2RuZW1WVWFETmxTR1F4VG1rNVZrNTZUalpXVlhoRVdsZG9VR1F6YkhaWlZYZDZXa1pHVDFZeVdsRlRSRUpVWkd4c1JHTlhjSEpVYTJSUldqRmtOVk16VG5aVk1WRjNUakZXUjFkdVdqRlhiWFJzVjFWU2FtTlZSVEJhTTI4d1dtdDBkMDR4V2tOaWVscDZZMnBLZUdGclNUTlNNR00xV1RKT1ZsVlhOVFJsUlRGVlUyMXdhRTlYVGs1VVYyUk1VMnhCZGxGcVl6UlZWM1J0VFVaYVExZFZVbmhOYTJ3MVYyNW9SbVJxUm5KV2FtTjJUREkxTmxOVmVFbE1NR1J4V1RCMFJrNUhNV3BYVlZrd1VteHNlVlZGZUZabGJrSmFVakZhYzJOVVVsaFpNRnBoVW0xb2NWcEhkRk5rTVVaRlRrZHJjbEV4YUhaa01GcDBWV3BDY1ZreWNGZFdNa2wzVmpCU1dHUlhkSFJVYkU1TlRXNVdjR0ZGVGpOalJXeHNVekJHV1U1dVdrTlhiVGxEVTI1YVdHRXlhRTlXUm1SMFYwVnNNMVJIVmtkVFYwcExaRlZWTUUweFJUbFFVMGx6U1cxa2MySXlTbWhpUlhSc1pWVnNhMGxxYjJsTlYxRjZXbXBrYlU1dFNURk5SR040VGtkYWJFNHlTVFJQVkZVeFRsUlNhMXBFV1RGT2VtTjZXV3BCYVdaUlBUMGlmUT09In0=`

	licenseFile, err := ioutil.TempFile("", "license")
	require.NoError(t, err)
	_, err = licenseFile.Write([]byte(data))
	require.NoError(t, err)
	err = licenseFile.Close()
	require.NoError(t, err)

	defer os.RemoveAll(licenseFile.Name())

	rootDir, err := ioutil.TempDir("", "kots")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	pullOptions := PullOptions{
		RootDir:   rootDir,
		Namespace: "default",
		Downstreams: []string{
			"this-cluster",
		},
		LicenseFile: licenseFile.Name(),
		AirgapRoot:  rootDir,
	}

	_, err = Pull("replicated://myapp", pullOptions)
	require.Error(t, err)

	// require.IsType(t, util.ActionableError{}, errors.Cause(err))
	// require.True(t, strings.Contains(err.Error(), "expired"), "error must contain expired")
}
