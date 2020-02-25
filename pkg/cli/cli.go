package cli

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jahkeup/updater53/pkg/whatip"
	"github.com/pkg/errors"
)

// CLI is the route53 update command runner.
type CLI struct {
	Records       []string
	IPResolver    whatip.IPResolver
	Route53Client route53Client
	DoCommit      bool
}

// route53Client masks the used API calls exported in the AWS SDK's Route53
// Client implementation. This permits testing implementations and scopes the
// CLI's updating activities.
type route53Client interface {
	ChangeResourceRecordSets(*route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
	ListHostedZonesByName(*route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error)
	ListResourceRecordSets(*route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error)
}

// Update runs through the records and updates the records if
// necessary.
func (cli *CLI) Update() error {
	currentIP, err := cli.IPResolver.GetIP()
	if err != nil {
		return fmt.Errorf("cannot determine IP: %s", err)
	}
	log.Printf("Your IP: %q", currentIP)

	zonemap, err := getHostedZoneMap(cli.Route53Client)
	if err != nil {
		return err
	}

	// Confirm each record has a matching Zone that exists.
	for _, rec := range cli.Records {
		if zoneID(zonemap, rec) == "" {
			return errors.Errorf("no hosted zone for record %q found", rec)
		}
	}

	for _, rec := range cli.Records {
		log.Printf("updating record %q", rec)
		if !cli.DoCommit {
			log.Printf("skipping...")
			continue
		}
		err = updateRecord(cli.Route53Client, zoneID(zonemap, rec), rec, currentIP)
		if err != nil {
			return errors.Errorf("error updating record %q: %s", rec, err)
		}
	}

	return nil
}

func getHostedZoneMap(client route53Client) (map[string]string, error) {
	zones, err := client.ListHostedZonesByName(nil)
	if err != nil {
		return nil, err
	}

	zonemap := map[string]string{}
	for _, zone := range zones.HostedZones {
		if zone.Name == nil || zone.Id == nil {
			return nil, errors.Errorf("error mapping hosted zone and its ID: %#v", zone)
		}
		zonemap[*zone.Name] = *zone.Id
	}
	return zonemap, nil
}

// zoneID resolves the appropriate zoneId for a given target record
// name.
func zoneID(zonemap map[string]string, targetRecord string) (zoneID string) {
	for name, id := range zonemap {
		if strings.HasSuffix(recordName(targetRecord), name) {
			return id
		}
	}

	return ""
}

// recordName returns the FQDN rooted name if not provided as such.
func recordName(targetRecord string) string {
	if targetRecord[len(targetRecord)-1] == '.' {
		return targetRecord
	}
	return string(targetRecord + ".")
}

// updateRecord reaches out to Route53 and UPSERTs the record if
// needed.
func updateRecord(client route53Client, zoneID string, targetRecord string, ip net.IP) error {
	name := recordName(targetRecord)
	// retrieve current record sets starting with our target name
	rrsets, err := client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneID),
		StartRecordName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("could not retrieve records for zoneID %q: %s", zoneID, err)
	}

	// check the IP address that there if it is.
	for _, rr := range rrsets.ResourceRecordSets {
		if *rr.Name == name && *rr.Type == route53.RRTypeA {
			if len((*rr).ResourceRecords) != 1 {
				return fmt.Errorf("cowardly refusing to modify a complicated ResourceRecord: multiple RR")
			}
			curr := *(*rr).ResourceRecords[0].Value
			if curr == ip.String() {
				log.Printf("no need to update record %q, already pointing to %q", name, ip)
				return nil
			}
		}
	}

	// UPSERT to create or update the record!
	_, err = client.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(name),
						Type: aws.String(route53.RRTypeA),
						TTL:  aws.Int64(60),
						ResourceRecords: []*route53.ResourceRecord{
							{Value: aws.String(ip.String())},
						},
					},
				},
			},
		},
	})

	return err
}
