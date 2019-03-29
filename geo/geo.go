package geo

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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

	/*ips := []string{ "91.78.44.23", "91.78.40.24", "91.78.80.18", "91.77.248.21", "91.79.205.128", "91.79.205.12", "91.80.0.17", "91.80.130.24", 
			"91.80.152.23", "91.78.224.24", "91.78.148.22", "91.78.144.24", "91.76.40.21", "91.75.168.21", "91.75.165.24" }
	res, _ := client.SelectIPByRadius(ip, ips, 5000)
	fmt.Println(res)*/

	fmt.Printf("Looking up the geo data for %s ...\n", ip)
	return lookupLocation(client.client, ip)
}

func (client *GeoClient) SelectIPByRadius(targetIP string, IPs []string, radius float64) ([]string, error) {

	fmt.Printf("Looking up all IPs in specified radius: %f <km> for %s ...\n", radius, targetIP)

	return selectIPByRadius(client.client, targetIP, IPs, radius)
}

func iPv4ToUint32(iPv4 string) uint32 {

	ipOctets := [4]uint64{}

	for i, v := range strings.SplitN(iPv4, ".", 4) {
		ipOctets[i], _ = strconv.ParseUint(v, 10, 32)
	}

	result := (ipOctets[0] << 24) | (ipOctets[1] << 16) | (ipOctets[2] << 8) | ipOctets[3]

	return uint32(result)
}

func uInt32ToIPv4(iPuInt32 uint32) (iP string) {
	iP = fmt.Sprintf("%d.%d.%d.%d",
		iPuInt32>>24,
		(iPuInt32&0x00FFFFFF)>>16,
		(iPuInt32&0x0000FFFF)>>8,
		iPuInt32&0x000000FF)
	return iP
}

func CIDRRangeToIPv4Range(CIDRs []string) (ipStart string, ipEnd string, err error) {
  
	var ip uint32        // ip address
  var ipS uint32     // Start IP address range
  var ipE uint32         // End IP address range

  for _, CIDR := range CIDRs {
     cidrParts := strings.Split(CIDR, "/")

     ip = iPv4ToUint32(cidrParts[0])
     bits, _ := strconv.ParseUint(cidrParts[1], 10, 32)

     if ipS == 0 || ipS > ip {
        ipS = ip
     }

     ip = ip | (0xFFFFFFFF >> bits)

     if ipE < ip {
        ipE = ip
     }
  }

  ipStart = uInt32ToIPv4(ipS)
  ipEnd = uInt32ToIPv4(ipE)

  return ipStart, ipEnd, err
}

func isDigit(str string) bool {

	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

func ipToScore(ip string, cidr bool) int {

	score := 0

	if cidr {

		var err error
		ip, _, err = CIDRRangeToIPv4Range([]string{ip})
		if err != nil {
			return 0
		}
	} 

	for _, v := range strings.Split(ip, ".") {

		n, _ := strconv.Atoi(v)
		score = score*256 + n
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

		var cityIP int64
		startIP := record[0]

		if strings.Index(startIP, ".") != -1 {
			cityIP = int64(ipToScore(startIP, true)) // CIDR or IP
		} else if isDigit(startIP) {
			cityIP, _ = strconv.ParseInt(startIP, 10, 32) // Integer score
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
		Name: ip,
		Longitude: longitude,
		Latitude: latitude,
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
		Max:    strconv.Itoa(IpId),
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
		Max:    strconv.Itoa(cityIP),
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

	redisKey  := uuid.New()

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
		Radius: radius,
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
