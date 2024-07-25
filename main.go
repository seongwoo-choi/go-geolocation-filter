package main

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"net/http"
	"strings"
)

var db *geoip2.Reader

func main() {
	var err error
	db, err = geoip2.Open("./GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *geoip2.Reader) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	http.HandleFunc("/", geoIPMiddleware(handleRequest))
	log.Println("Starting server on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func geoIPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		originalIP := r.RemoteAddr

		record, err := db.City(net.ParseIP(ip))
		if err != nil {
			log.Printf("Error looking up IP %s: %v", ip, err)
		} else {
			w.Header().Set("X-Geo-Country", record.Country.IsoCode)
			w.Header().Set("X-Geo-City", record.City.Names["en"])
			w.Header().Set("X-Geo-Latitude", fmt.Sprintf("%f", record.Location.Latitude))
			w.Header().Set("X-Geo-Longitude", fmt.Sprintf("%f", record.Location.Longitude))

			log.Printf("URL: %v, USER_AGETN: %v, IP: %s (Original: %s), Country: %s, City: %s, Lat: %f, Lon: %f",
				r.URL.Path, r.Header.Get("User-Agent"), ip, originalIP, record.Country.IsoCode, record.City.Names["en"],
				record.Location.Latitude, record.Location.Longitude)
		}
		next.ServeHTTP(w, r)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	return
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	if strings.Contains(ip, "127.0.0.1") || strings.Contains(ip, "::1") || strings.Contains(ip, "localhost") {
		return "8.8.8.8"
	}

	ip = strings.Split(ip, ":")[0]

	return ip
}
