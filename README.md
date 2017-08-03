# Update Route53 - Dynamic DNS on Route53

[![Build Status](https://travis-ci.org/jahkeup/updater53.svg?branch=master)](https://travis-ci.org/jahkeup/updater53)

This program enables you to update a record on Route53 to point to
your public IP address if you don't have a static address. This can be
scheduled using cron or any other scheduling tool of choice.

## Example usage:

```bash
$ ./updater53 -records house.example.com
2017/03/17 00:11:18 Your IP: "108.228.144.143"
2017/03/17 00:11:18 updating record "house.example.com"

# and if you run this again after updating the record:

$ ./updater53 -records house.example.com
2017/03/17 00:11:18 Your IP: "108.228.144.143"
2017/03/17 00:11:18 updating record "house.example.com"
2017/03/17 00:12:42 no need to update record "house.example.com.",
already pointing to "108.228.144.143"
```

## Install

This tool is go gettable and can be installed by running `go get -u -x
github.com/jahkeup/updater53`. The AWS Authentication is handled by
the Golang SDK and can be configured in the same fashion as
you'd
[configure the aws cli utility or other SDK usages](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) (in
the `$HOME/.aws/credentials` file or via environment variables).
