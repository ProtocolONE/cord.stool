package geo

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
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

	return importBlocks(client.client, fname)
}

func (client *GeoClient) ImportLocations(fname string) error {

	return importLocations(client.client, fname)
}

func (client *GeoClient) LookupLocation(ip string) (string, error) {

	return lookupLocation(client.client, ip)
}

func isDigit(str string) bool {

	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

func ipToScore(ip string) int {

	score := 0

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
	for i := 0; ; i++ {

		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var cityIP int64
		startIP := record[0]

		if strings.Index(startIP, ".") != -1 {
			cityIP = int64(ipToScore(startIP))
		} else if isDigit(startIP) {
			cityIP, _ = strconv.ParseInt(startIP, 10, 32)
		} else {
			continue
		}

		cityID := record[1] + "_" + strconv.Itoa(i)
		_, err = client.ZAdd("ip2cityid", redis.Z{
			Score:  float64(cityIP),
			Member: cityID,
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

func lookupLocation(client *redis.Client, ip string) (string, error) {

	ipScore := ipToScore(ip)

	vals, err := client.ZRevRangeByScore("ip2cityid", redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.Itoa(ipScore),
		Offset: 0,
		Count:  10,
	}).Result()

	if err != nil {
		return "", err
	}

	if len(vals) == 0 {
		return "", nil
	}

	cityID := strings.Split(vals[0], "_")[0]
	res := client.HGet("cityid2city", cityID)

	return res.Val(), nil
}
