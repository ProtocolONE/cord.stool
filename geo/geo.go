package geo

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/pborman/uuid"
)

type GeoClient struct {
	client *redis.Client
}

func NewGeoClient(host string, port string) *GeoClient {

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &GeoClient{client: client}
}

func (client *GeoClient) ImportBlocks(fname string) error {

	fmt.Printf("Importing the location blocks from %s ...\n", fname)
	return importBlocks(client.client, fname)
}

func (client *GeoClient) ImportLocations(fname string) error {

	fmt.Printf("Importing the location details from %s ...\n", fname)
	return importLocations(client.client, fname)
}

func (client *GeoClient) LookupLocation(ip string) (string, error) {

	fmt.Printf("Looking up the geo data for %s ...\n", ip)
	return lookupLocation(client.client, ip)
}

func (client *GeoClient) SelectIPByRadius(targetIP string, IPs []string, radius float64) ([]string, error) {

	fmt.Printf("Looking up all IPs in specified radius: %f <km> for %s ...\n", radius, targetIP)
	return selectIPByRadius(client.client, targetIP, IPs, radius)
}

func (client *GeoClient) IsIPInRadius(targetIP string, IP string, radius float64) bool {

	res, err := selectIPByRadius(client.client, targetIP, []string{IP}, radius)
	if err != nil {
		return false
	}

	return len(res) > 0
}

func isDigit(str string) bool {

	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

func ipRange(str string) (net.IP, net.IP, error) {

	_, mask, err := net.ParseCIDR(str)
	if err != nil {
		return nil, nil, err
	}

	first := mask.IP.Mask(mask.Mask).To16()
	second := make(net.IP, len(first))
	copy(second, first)
	ones, _ := mask.Mask.Size()

	if first.To4() != nil {
		ones += 96
	}

	lastBytes := (8*16 - ones) / 8
	lastBits := 8 - ones%8
	or := 0

	for x := 0; x < lastBits; x++ {
		or = or*2 + 1
	}

	for x := 16 - lastBytes; x < 16; x++ {
		second[x] = 0xff
	}

	if lastBits < 8 {
		second[16-lastBytes-1] |= byte(or)
	}

	return first, second, nil
}

func ipToScore(ip string, cidr bool) uint64 {

	var score uint64
	score = 0

	if cidr {

		startIP, _, err := ipRange(ip)
		if err != nil {
			return 0
		}

		ip = startIP.String()
	}

	if strings.Index(ip, ".") != -1 {

		// IPv4
		for _, v := range strings.Split(ip, ".") {

			n, _ := strconv.Atoi(v)
			score = score*256 + uint64(n)
		}

	} else if strings.Index(ip, ":") != -1 {

		//IPv6 is not supported
	}

	return score
}

func importBlocks(client *redis.Client, filename string) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	r := csv.NewReader(f)
	i := 0
	for ; ; i++ {

		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var cityIP uint64
		startIP := record[0]

		if strings.Index(startIP, ".") != -1 {
			cityIP = ipToScore(startIP, true) // CIDR or IP
		} else if isDigit(startIP) {
			cityIP, _ = strconv.ParseUint(startIP, 10, 64) // Integer score
		} else {
			continue
		}

		// Add IP to City info
		cityID := record[1] + "_" + strconv.Itoa(i)
		_, err = client.ZAdd("ip2cityid", redis.Z{
			Score:  float64(cityIP),
			Member: cityID,
		}).Result()

		if err != nil {
			return err
		}

		// Add IP to locatiom
		data, err := json.Marshal([]string{
			strconv.Itoa(i),
			record[8],
			record[7],
		})

		_, err = client.ZAdd("ip2location", redis.Z{
			Score:  float64(cityIP),
			Member: data,
		}).Result()

		if err != nil {
			return err
		}
	}

	return nil
}

func importLocations(client *redis.Client, filename string) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	r := csv.NewReader(f)
	for i := 0; ; i++ {

		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if len(record) < 14 || !isDigit(record[0]) {
			continue
		}

		cityID := record[0]

		data, err := json.Marshal([]string{
			record[1],
			record[2],
			record[3],
			record[4],
			record[5],
			record[6],
			record[7],
			record[8],
			record[9],
			record[10],
			record[11],
			record[12],
			record[13],
		})

		_, err = client.HSet("cityid2city", cityID, data).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

func addGeo(client *redis.Client, key string, ip string, longitude, latitude float64) error {

	_, err := client.GeoAdd(key, &redis.GeoLocation{
		Name:      ip,
		Longitude: longitude,
		Latitude:  latitude,
	}).Result()

	return err
}

func getCord(client *redis.Client, key string, ip string) (float64, float64, error) {

	var longitude, latitude float64
	longitude = 0
	latitude = 0

	IpId := ipToScore(ip, false)

	vals, err := client.ZRevRangeByScore(key, redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatUint(IpId, 10),
		Offset: 0,
		Count:  1,
	}).Result()

	if err != nil {
		return longitude, latitude, err
	}

	var cord []string

	if len(vals) != 0 {

		err = json.Unmarshal([]byte(vals[0]), &cord)
		if err != nil {
			return longitude, latitude, err
		}

		longitude, err = strconv.ParseFloat(cord[1], 64)
		if err != nil || longitude == 0 {
			return longitude, latitude, err
		}

		latitude, err = strconv.ParseFloat(cord[2], 64)
		if err != nil || latitude == 0 {
			return longitude, latitude, err
		}
	}

	return longitude, latitude, nil
}

func lookupLocation(client *redis.Client, ip string) (string, error) {

	cityIP := ipToScore(ip, false)

	vals, err := client.ZRevRangeByScore("ip2cityid", redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatUint(cityIP, 10),
		Offset: 0,
		Count:  1,
	}).Result()

	if err != nil {
		return "", err
	}

	city := ""
	if len(vals) != 0 {
		cityID := strings.Split(vals[0], "_")[0]
		city = client.HGet("cityid2city", cityID).Val()
	}

	longitude, latitude, err := getCord(client, "ip2location", ip)
	if err != nil {
		return "", err
	}

	loc := fmt.Sprintf(", [%f, %f]\n", longitude, latitude)

	return city + loc, nil
}

func selectIPByRadius(client *redis.Client, targetIP string, IPs []string, radius float64) ([]string, error) {

	var result []string

	targetLongitude, targetLatitude, err := getCord(client, "ip2location", targetIP)
	if err != nil {
		return result, err
	}

	redisKey := uuid.New()

	for _, ip := range IPs {

		longitude, latitude, err := getCord(client, "ip2location", ip)
		if err != nil {
			client.Del(redisKey)
			return result, err
		}

		err = addGeo(client, redisKey, ip, longitude, latitude)
		if err != nil {
			client.Del(redisKey)
			return result, err
		}
	}

	defer client.Del(redisKey)

	geolocs, err := client.GeoRadius(redisKey, targetLongitude, targetLatitude, &redis.GeoRadiusQuery{
		Radius:   radius,
		WithDist: true,
	}).Result()

	if err != nil {
		return result, err
	}

	for i, gl := range geolocs {

		t := fmt.Sprintf("#%d, ip: %s, dist: %f\n", i, gl.Name, gl.Dist)
		result = append(result, t)
	}

	return result, nil
}
