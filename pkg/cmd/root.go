package cmd

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bkimmig/awscurl/pkg/lib"
	"github.com/spf13/cobra"
)

type CurlFlags struct {
	headers []string
	method  string
	data    string

	awsAccessKey    string
	awsSecretKey    string
	awsSessionToken string
	awsProfile      string
	awsService      string
	awsRegion       string
	insecure        bool
	portMap         string
}

var flags CurlFlags

var rootCmd = &cobra.Command{
	Use:   "awscurl",
	Short: "cURL with AWS request signing",
	Long: `A simple CLI utility with cURL-like syntax allowing to send HTTP requests to AWS resources.
It automatically adds Siganture Version 4 to the request. More details:
https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html
`,
	Args:    cobra.ExactArgs(1),
	RunE:    runCurl,
	Version: getVersion(),
}

func runCurl(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	// setup the config from either the variables or the profile
	cfg, err := lib.GetAwsConfig(flags.awsProfile, flags.awsAccessKey, flags.awsSecretKey, flags.awsSessionToken, flags.awsRegion)
	if err != nil {
		return err
	}

	body, err := lib.ConstructBody(flags.data)
	if err != nil {
		return err
	}

	requestUrl := args[0]
	certificateUrl := requestUrl
	if flags.portMap != "" {
		ports := strings.Split(flags.portMap, ":")
		certificateUrl = strings.Replace(requestUrl, ":"+ports[0], ":"+ports[1], -1)
	}

	request, err := http.NewRequest(flags.method, certificateUrl, body)
	if err != nil {
		return err
	}

	// set up the request
	lib.AddHeaders(request, flags.headers)
	err = lib.Sign(request, cfg, flags.awsService, flags.awsRegion)
	if err != nil {
		return err
	}

	// put the url back to the original url
	u, err := url.Parse(requestUrl)
	if err != nil {
		return err
	}
	request.URL = u

	// Set TLS Client configuration
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: flags.insecure},
	}

	// Send the request and print the response
	client := http.Client{Transport: tr}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flags.method, "request", "X", "GET", "Custom request method to use")
	rootCmd.PersistentFlags().StringVarP(&flags.data, "data", "d", "", `Data payload to send within a request. Could be also read from a file if prefixed with @, example: -d "@/path/to/file.json"`)
	rootCmd.PersistentFlags().StringArrayVarP(&flags.headers, "header", "H", []string{},
		`Extra HTTP header to include in the request. Example: -h "Content-Type: application/json". Could be used multiple times`)
	rootCmd.PersistentFlags().StringVar(&flags.awsAccessKey, "access-key", "", "AWS Access Key ID to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsSecretKey, "secret-key", "", "AWS Secret Access Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsSessionToken, "session-token", "", "AWS Session Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsProfile, "profile", "", "AWS awsProfile to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsService, "service", "execute-api", "The name of AWS Service, used for signing the request")
	rootCmd.PersistentFlags().StringVar(&flags.awsRegion, "region", "", "AWS region to use for the request")
	rootCmd.PersistentFlags().BoolVarP(&flags.insecure, "insecure", "k", false, "Allow insecure server connections when using SSL")
	rootCmd.PersistentFlags().StringVar(&flags.portMap, "port-map", "", "map the local port to the remote port 'localport:remoteport'. Used in cases where you are locally port-forwarding the service")
	// add the cert url here, will be used for constructing the request
	rootCmd.Flags().SortFlags = false

}

// ----------------------------------------------------
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
