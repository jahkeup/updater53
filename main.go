package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jahkeup/updater53/pkg/whatip"
)

func main() {
	flagIPMethod := flag.String("ipmethod", "opendns", "IP lookup provider: opendns, ifconfigme, icanhazip")
	flagRecords := flag.String("records", "", "Records to be updated, these are the A records that you want updated")
	flag.Parse()

	if *flagRecords == "" {
		log.Fatalf("you must provide some records to update")
	}
	records := strings.Split(*flagRecords, ",")

	var iper whatip.IPer
	switch *flagIPMethod {
	case "opendns":
		iper = whatip.OpenDNS
	case "ifconfigme":
		iper = whatip.IfconfigMeHTTP
	case "icanhazip":
		iper = whatip.ICanHazIPHTTP
	default:
		log.Panicf("unknown ipmethod %q", *flagIPMethod)
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("could not setup AWS session: %s", err)
	}

	err = Update(Config{
		Records: records,
		IPer:    iper,
		Session: sess,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Update runs through the records and updates the records if
// necessary.
func Update(conf Config) (err error) {
	r53 := route53.New(conf.Session)

	zones, err := r53.ListHostedZonesByName(nil)
	if err != nil {
		return err
	}

	zonemap := map[string]string{}

	for _, zone := range zones.HostedZones {
		zonemap[*zone.Name] = *zone.Id
	}

zoneCheck:
	for _, rec := range conf.Records {
		if zoneID(zonemap, rec) != "" {
			continue zoneCheck
		}

		return fmt.Errorf("no zone for record %q found", rec)
	}

	newip, err := conf.IPer.GetIP()
	if err != nil {
		return fmt.Errorf("cannot determine IP: %s", err)
	}
	log.Printf("Your IP: %q", newip)

	for _, rec := range conf.Records {
		log.Printf("updating record %q", rec)
		err = updateRecord(r53, zoneID(zonemap, rec), rec, newip)
		if err != nil {
			return fmt.Errorf("error updating record %q: %s", rec, err)
		}
	}

	return nil
}

// updateRecord reaches out to Route53 and UPSERTs the record if
// needed.
func updateRecord(client *route53.Route53, zoneID string, targetRecord string, ip net.IP) (err error) {
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

// recordName returns the FQDN rooted name if not provided as such.
func recordName(targetRecord string) string {
	if targetRecord[len(targetRecord)-1] == '.' {
		return targetRecord
	}
	return string(targetRecord + ".")
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
