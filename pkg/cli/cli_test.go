package cli

import (
	"net"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/require"
)

type testResolver struct {
	IP net.IP
}

func (r *testResolver) GetIP() (net.IP, error) {
	return r.IP, nil
}

type testRoute53 struct {
	listHostedZonesByNameResponse *route53.ListHostedZonesByNameOutput
	listResourceRecordSets        *route53.ListResourceRecordSetsOutput

	updated bool
}

func (r53 *testRoute53) ChangeResourceRecordSets(*route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	r53.updated = true
	return nil, nil
}

func (r53 *testRoute53) ListHostedZonesByName(*route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	return r53.listHostedZonesByNameResponse, nil
}

func (r53 *testRoute53) ListResourceRecordSets(*route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	return r53.listResourceRecordSets, nil
}

func TestHappyPath(t *testing.T) {
	route53Mock := &testRoute53{
		listHostedZonesByNameResponse: &route53.ListHostedZonesByNameOutput{
			HostedZones: []*route53.HostedZone{
				&route53.HostedZone{
					Id:   aws.String("hostedZoneId"),
					Name: aws.String("example.com."),
				},
			},
		},
		listResourceRecordSets: &route53.ListResourceRecordSetsOutput{},
	}
	cli := &CLI{
		Records:       []string{"record.example.com"},
		IPResolver:    &testResolver{net.ParseIP("169.254.0.1")},
		Route53Client: route53Mock,
		DoCommit:      true,
	}

	require.NoError(t, cli.Update())
	require.True(t, route53Mock.updated, "record should have been updated")
}

func TestNoHostedZone(t *testing.T) {
	route53Mock := &testRoute53{
		listHostedZonesByNameResponse: &route53.ListHostedZonesByNameOutput{
			HostedZones: []*route53.HostedZone{
				&route53.HostedZone{
					Id:   aws.String("hostedZoneId"),
					Name: aws.String("example.com."),
				},
			},
		},
		listResourceRecordSets: &route53.ListResourceRecordSetsOutput{},
	}
	cli := &CLI{
		Records:       []string{"record.example.net"},
		IPResolver:    &testResolver{net.ParseIP("169.254.0.1")},
		Route53Client: route53Mock,
		DoCommit:      true,
	}

	require.Error(t, cli.Update())
	require.False(t, route53Mock.updated, "record should not have been updated")
}
