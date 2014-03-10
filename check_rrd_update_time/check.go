package main

import (
	"flag"
	"fmt"
	"github.com/fractalcat/nagiosplugin"
	"io/ioutil"
	"math"
	"os"
	"time"
)

func main() {
	c := nagiosplugin.NewCheck()
	defer c.Finish()
	rrd_path := flag.String("rrd-path", "/var/lib/pnp4nagios/perfdata", "Path to directory containing Nagios perfdata")
	warnSec := flag.Int("warn", 300, "Warning threshold (seconds)")
	critSec := flag.Int("crit", 900, "Critical threshold (seconds)")
	warn := time.Duration(*warnSec) * time.Second
	crit := time.Duration(*critSec) * time.Second
	flag.Parse()
	dirInfo, err := os.Stat(*rrd_path)
	if err != nil {
		c.Unknownf("Unable to stat perfdata directory %v: %v", *rrd_path, err)
	}
	if !dirInfo.IsDir() {
		c.Unknownf("Perfdata directory %v is not a directory", *rrd_path)
	}
	var newest, oldest *time.Time
	dents, err := ioutil.ReadDir(*rrd_path)
	if err != nil {
		c.Unknownf("Unable to get directory listing for %v: %v", *rrd_path, err)
	}
	for _, fi := range dents {
		modtime := fi.ModTime()
		if newest == nil {
			newest = &modtime
		}
		if oldest == nil {
			oldest = &modtime
		}
		if modtime.After(*newest) {
			newest = &modtime
		} else if modtime.Before(*oldest) {
			oldest = &modtime
		}
	}
	dNewest := time.Since(*newest)
	dOldest := time.Since(*oldest)
	info := fmt.Sprintf("Newest RRD update is %v old", dNewest)
	if crit < dNewest {
		c.AddResultf(nagiosplugin.CRITICAL, info)
	} else if warn < dNewest {
		c.AddResultf(nagiosplugin.WARNING, info)
	} else {
		c.AddResultf(nagiosplugin.OK, info)
	}
	perfNewest := float64(dNewest) / float64(time.Second)
	perfOldest := float64(dOldest) / float64(time.Second)
	c.AddPerfDatum("newest_update", "s", perfNewest, 0.0, math.Inf(1), float64(*warnSec), float64(*critSec))
	c.AddPerfDatum("oldest_update", "s", perfOldest, 0.0, math.Inf(1))
}
