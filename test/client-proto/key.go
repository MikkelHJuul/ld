package ld_proto

import (
	"fmt"
	"github.com/mmcloughlin/geohash"
	"time"
)

func KeyFromMessage(feature *Feature) string {
	geoHash := geohash.EncodeWithPrecision(feature.Geometry.Coordinates[1], feature.Geometry.Coordinates[0], 8)
	timeGeoHashPrefix := HandmadeTimeKeyString(feature.Properties.Observed) + geoHash
	amp := int(feature.Properties.Amp * 10)
	return fmt.Sprintf("%s%d%d", timeGeoHashPrefix, feature.Properties.Type, amp)
}

// "2018-07-04T19:01:12.324000Z" >> 018185190
// ten-minute partitioned time
func HandmadeTimeKeyString(timeIn string) string {
	featureTime, _ := time.Parse(time.RFC3339, timeIn)
	threeByteYear := featureTime.Year() - int(featureTime.Year()/1000)*1000 //skip the millennium
	dayOfYear := featureTime.YearDay()
	tenMinHour := int(featureTime.Minute() / 10)
	return fmt.Sprintf("%03d%03d%02d%01d", threeByteYear, dayOfYear, featureTime.Hour(), tenMinHour)
}
