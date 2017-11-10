package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var interval = flag.Int("interval", 60, "Interval in seconds between checks (adjust scaler's load vs precision)")
var destination = flag.String("destination", "http://gpu-scaler-service:1110", "Base hostname of the gpu-scaler service")
var fixed = flag.Float64("fixed-value", -1, "Provide a fixed value instead of calling `nvidia-smi`. For testing")

func main() {
	flag.Parse()

	// Send to the `destination`/v1/gpu_usage?hostname=[hostname]&volatile_gpu_usage=[float]

	for {
		time.Sleep(time.Duration(*interval) * time.Second)

		utilization, err := fetchUtilization()

		req, err := http.NewRequest("POST", *destination, nil)
		if err != nil {
			log.Println("Invalid destination URL:", err)
			continue
		}

		hostname, _ := os.Hostname()

		req.URL.Path = "/v1/gpu_usage"
		vals := url.Values{}
		vals.Set("hostname", hostname)
		vals.Set("volatile_gpu_usage", fmt.Sprintf("%f", utilization))
		req.URL.RawQuery = vals.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error launching HTTP request to scaler:", err)
			continue
		}

		if resp.StatusCode != 200 {
			log.Println("HTTP status error from remote host:", resp.StatusCode)
			continue
		}

		log.Printf("Submitted %f successfully\n", utilization)
	}
}

func fetchUtilization() (float64, error) {
	if *fixed != -1 {
		return *fixed, nil
	}

	out, err := exec.Command("bash", "-c", "nvidia-smi -q -x | grep gpu_util | cut -d'>' -f2 | cut -d' ' -f1").Output()
	if err != nil {
		return 0, fmt.Errorf("Couldn't run nvidia-smi command: %s", err)
	}

	utilization, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid output from nvidia-smi: %s", err)
	}

	return utilization, nil
}
