package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"
)

type H1Response struct {
	Data []Data `json:"data"`
}

type Data struct {
	Id         string    `json:"id"`
	Type       string    `json:"type"`
	Attributes Attribute `json:"attributes"`
}

type Attribute struct {
	Handle                               string  `json:"handle"`
	Name                                 string  `json:"name"`
	Currency                             string  `json:"currency"`
	Policy                               string  `json:"policy"`
	Profile_picture                      string  `json:"profile_picture"`
	Submission_state                     string  `json:"submission_state"`
	Triage_active                        string  `json:"triage_active"`
	State                                string  `json:"state"`
	Started_accepting_at                 string  `json:"started_accepting_at"`
	Number_of_reports_for_user           int     `json:"number_of_reports_for_user"`
	Number_of_valid_reports_for_user     int     `json:"number_of_valid_reports_for_user"`
	Bounty_earned_for_user               float32 `json:"bounty_earned_for_user"`
	Last_invitation_accepted_at_for_user string  `json:"last_invitation_accepted_at_for_user"`
	Bookmarked                           bool    `json:"bookmarked"`
	Allows_bounty_splitting              bool    `json:"allows_bounty_splitting"`
}

type ProgramDetails struct {
	Attributes    Attribute     `json:"attributes"`
	Relationships Relationships `json:"relationships"`
}

type Relationships struct {
	StructedScopes StructedScopes `json:"structured_scopes"`
}

type StructedScopes struct {
	Data []StructedScope `json:"data"`
}

type StructedScope struct {
	Id         string                 `json:"id"`
	Type       string                 `json:"string"`
	Attributes StructedScopeAttribute `json:"attributes"`
}

type StructedScopeAttribute struct {
	AssetType                  string `json:"asset_type"`
	AssetIdentifier            string `json:"asset_identifier"`
	EligibleForBounty          bool   `json:"eligible_for_bounty"`
	Instruction                string `json:"instruction"`
	MaxSeverity                string `json:"max_severity"`
	CreatedAt                  string `json:"created_at"`
	UpdatedAt                  string `json:"updated_at"`
	ConfidentialityRequirement string `json:"confidentiality_requirement"`
	IntegrityRequirement       string `json:"integrity_requirement"`
	Availabilityrequirement    string `json:"availibity_requirement"`
}

func main() {
	username := flag.String("u", "", "Username for basic auth")
	apiToken := flag.String("t", "", "API token for basic auth")
	flag.Parse()

	// urls := make(chan string)

	// wg := sync.WaitGroup{}
	// s := bufio.NewScanner(os.Stdin)

	// for s.Scan() {
	// 	urls <- s.Text()
	// }

	makeAPIRequest(*username, *apiToken)

	// close(urls)
	// wg.Wait()
}

func makeAPIRequest(username string, token string) {
	client := buildHttpClient()
	var data H1Response

	req, err := http.NewRequest("GET", "https://api.hackerone.com/v1/hackers/programs?page[size]=100", nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.SetBasicAuth(username, token)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&data)

	for _, details := range data.Data {
		var programDetails ProgramDetails

		detailReq, err := http.NewRequest("GET", "https://api.hackerone.com/v1/hackers/programs/"+details.Attributes.Handle, nil)

		if err != nil {
			fmt.Println(err)
			return
		}

		detailReq.SetBasicAuth(username, token)

		detailResp, err := client.Do(detailReq)

		if err != nil {
			fmt.Println(err)
		}

		defer detailResp.Body.Close()

		json.NewDecoder(detailResp.Body).Decode(&programDetails)

		for _, scope := range programDetails.Relationships.StructedScopes.Data {
			if scope.Attributes.AssetType == "URL" {
				fmt.Println(scope.Attributes.AssetIdentifier)
			}
		}
	}

	fmt.Println(data)
}

func buildHttpClient() *http.Client {
	var tr = &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: time.Second,
		}).DialContext,
	}

	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client := &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       10 * time.Second,
	}

	return client
}
