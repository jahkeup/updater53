package main

import (
	"flag"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jahkeup/updater53/pkg/cli"
	"github.com/jahkeup/updater53/pkg/whatip"
)

func main() {
	flagIPMethod := flag.String("ipmethod", "opendns", "IP lookup provider: opendns, ifconfigme, icanhazip, aws")
	flagRecords := flag.String("records", "", "Records to be updated, these are the A records that you want updated")
	flagDry := flag.Bool("dryrun", false, "Dry-run")
	flagCustom := flag.String("custom", "", "Custom IP lookup provider")
	flag.Parse()

	if *flagRecords == "" {
		log.Fatalf("you must provide some records to update")
	}
	records := strings.Split(*flagRecords, ",")

	var resolver whatip.IPResolver
	if *flagCustom != "" {
		var err error
		resolver, err = whatip.NewHTTPResolver(*flagCustom)
		if err != nil {
			log.Panicf("error confguring custom ipmethod %q: %s", *flagCustom, err)
		}
	} else {
		switch *flagIPMethod {
		case "opendns":
			resolver = whatip.OpenDNS
		case "ifconfigme":
			resolver = whatip.IfconfigMeHTTP
		case "icanhazip":
			resolver = whatip.ICanHazIPHTTP
		case "aws":
			resolver = whatip.AWSHTTP
		default:
			log.Panicf("unknown ipmethod %q", *flagIPMethod)
		}
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("could not setup AWS session: %s", err)
	}

	run := &cli.CLI{
		Records:       records,
		IPResolver:    resolver,
		Route53Client: route53.New(sess),
		DoCommit:      !*flagDry,
	}

	if err = run.Update(); err != nil {
		log.Fatal(err)
	}
}
