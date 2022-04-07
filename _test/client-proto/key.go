package ld_proto

import (
	"fmt"
	"github.com/mmcloughlin/geohash"
	"time"
)

// KeyFromMessage given a Feature, returns a key representing the measurements time, geohash and some extra data
// This is "{YYY}{DDD}{HH}{10M}{GEOHASH8}{Type}{Amp}" the key is at least 19 characters long.
func KeyFromMessage(feature *Feature) string {
	geoHash := geohash.EncodeWithPrecision(feature.Geometry.Coordinates[1], feature.Geometry.Coordinates[0], 8)
	timeGeoHashPrefix := HandmadeTimeKeyString(feature.Properties.Observed) + geoHash
	amp := int(feature.Properties.Amp * 10)
	return fmt.Sprintf("%s%d%d", timeGeoHashPrefix, feature.Properties.Type, amp)
}

//HandmadeTimeKeyString returns a 9 byte string that is breakable on year, day-of-year, hour and ten minutes.
// fx "2018-07-04T19:01:12.324000Z" >> 018185190
func HandmadeTimeKeyString(timeIn string) string {
	featureTime, _ := time.Parse(time.RFC3339, timeIn)
	threeByteYear := featureTime.Year() - int(featureTime.Year()/1000)*1000 //skip the millennium
	dayOfYear := featureTime.YearDay()
	tenMinHour := int(featureTime.Minute() / 10)
	return fmt.Sprintf("%03d%03d%02d%01d", threeByteYear, dayOfYear, featureTime.Hour(), tenMinHour)
}
