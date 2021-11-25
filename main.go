package main

import (
	"context"
	"encoding/csv"
	"flag"
	"github.com/gen2brain/dlgs"
	"github.com/google/go-github/v40/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func main() {
	ctx := context.Background()

	token := *flag.String("token", "", "GitHub personal access token with admin:read scope and authorized for SAML SSO.")
	org := *flag.String("org", "", "GitHub enterprise slug.")
	outFile := *flag.String("out", "github_users.csv", "Output CSV filepath")
	flag.Parse()

	if org == "" {
		org, _, _ = dlgs.Entry("GitHub organisation", "Please provide your GitHub Enterprise slug.", "")
	}

	if token == "" {
		token, _, _ = dlgs.Entry("GitHub Token", "Please provide your GitHub access token.\nIt should have admin:read scope and be authorized for SAML SSO.", "")
	}

	if outFile == "" {
		logrus.Fatal("output file not provided")
	}

	if strings.TrimSpace(token) == "" {
		logrus.Fatal("GitHub token not provided")
	}

	if strings.TrimSpace(org) == "" {
		logrus.Fatal("GitHub organisation not provided")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	tc.Timeout = time.Second * 10

	client := github.NewClient(tc)

	recordCount := 10000
	resp, err := client.SCIM.ListSCIMProvisionedIdentities(ctx, org, &github.ListSCIMProvisionedIdentitiesOptions{
		StartIndex: nil,
		Count:      &recordCount,
		Filter:     nil,
	})

	if err != nil {
		logrus.WithError(err).Fatal("request to github failed")
	}

	bResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Fatal("could not read github response")
	}

	logrus.Printf("%s", string(bResp))

	// TODO: retrieve users from SCIM

	ghUsers := [][]string{}

	fileCSV, err := os.Create(outFile)
	if err != nil {
		logrus.WithField("out", outFile).WithError(err).Fatal("could not create output file")
	}
	defer fileCSV.Close()

	csvWriter := csv.NewWriter(fileCSV)
	defer csvWriter.Flush()

	if err := csvWriter.WriteAll(ghUsers); err != nil {
		logrus.WithField("out", outFile).WithError(err).Fatal("could not write to output file")
	}

	logrus.WithField("out", outFile).WithField("users", len(ghUsers)).Info("successfully exported to file")
}