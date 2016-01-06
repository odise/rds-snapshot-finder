package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/bradfitz/slice"
)

// Config commandline parameter
type Config struct {
	Db     string
	Latest bool
}

var config *Config

func init() {
	config = &Config{}
	flag.StringVar(&config.Db, "db", "", "RDS DB identifier to filter snapshots.")
	flag.BoolVar(&config.Latest, "latest", false, "Only return the latest snapshot with status 'available'.")
}

func main() {
	flag.Parse()

	svc := rds.New(session.New())
	var (
		snaps []*rds.DBSnapshot
	)
	marker := aws.String("")
	for {
		params := &rds.DescribeDBSnapshotsInput{
			DBInstanceIdentifier: aws.String(config.Db),
			MaxRecords:           aws.Int64(100),
		}
		if *marker != "" {
			params.Marker = marker
		}
		resp, err := svc.DescribeDBSnapshots(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return
		}

		snaps = append(snaps, resp.DBSnapshots...)
		if resp.Marker != nil {
			marker = resp.Marker
		} else {
			break
		}

	}

	slice.Sort(snaps, func(i, j int) bool {
		if snaps[i].SnapshotCreateTime == nil || *snaps[i].Status != "available" {
			return false
		}
		if snaps[j].SnapshotCreateTime == nil || *snaps[j].Status != "available" {
			return true
		}
		return snaps[i].SnapshotCreateTime.Before(*snaps[j].SnapshotCreateTime)
	})

	// only print the latest snapshot
	if config.Latest {
		i := 1
		for ; i < len(snaps); i++ {
			if *snaps[len(snaps)-i].Status == "available" {
				break
			}
		}
		// if there are snapshots and the last one found is available print it out and return
		if len(snaps) > 0 && *snaps[len(snaps)-i].Status == "available" {
			fmt.Println(snaps[len(snaps)-i])
			return
		}
	} else {
		fmt.Println(snaps)
	}
}
