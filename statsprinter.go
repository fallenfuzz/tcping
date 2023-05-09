package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gookit/color"
)

var (
	colorYellow      = color.Yellow.Printf
	colorGreen       = color.Green.Printf
	colorRed         = color.Red.Printf
	colorCyan        = color.Cyan.Printf
	colorLightYellow = color.LightYellow.Printf
	colorLightBlue   = color.FgLightBlue.Printf
	colorLightGreen  = color.LightGreen.Printf
	colorLightCyan   = color.LightCyan.Printf
)

const (
	noReply = "No reply"
)

type statsPrinter interface {
	printStart()
	printLastSucUnsucProbes()
	printDurationStats()
	printStatistics()
	printReply(replyMsg replyMsg)
	printTotalDownTime(time.Time)
	printLongestUptime()
	printLongestDowntime()
	printRetryResolveStats()
	printRetryingToResolve()
}

type statsPlanePrinter struct {
	*stats
}

type statsJsonPrinter struct {
	*stats
}

/* Print host name and port to use on tcping */
func (p *statsPlanePrinter) printStart() {
	colorLightCyan("TCPinging %s on port %d\n", p.hostname, p.port)
}

/* Print the last successful and unsuccessful probes */
func (p *statsPlanePrinter) printLastSucUnsucProbes() {
	formattedLastSuccessfulProbe := p.lastSuccessfulProbe.Format(timeFormat)
	formattedLastUnsuccessfulProbe := p.lastUnsuccessfulProbe.Format(timeFormat)

	colorYellow("last successful probe:   ")
	if formattedLastSuccessfulProbe == nullTimeFormat {
		colorRed("Never succeeded\n")
	} else {
		colorGreen("%v\n", formattedLastSuccessfulProbe)
	}

	colorYellow("last unsuccessful probe: ")
	if formattedLastUnsuccessfulProbe == nullTimeFormat {
		colorGreen("Never failed\n")
	} else {
		colorRed("%v\n", formattedLastUnsuccessfulProbe)
	}
}

/* Print the start and end time of the program */
func (p *statsPlanePrinter) printDurationStats() {
	var duration time.Time
	var durationDiff time.Duration

	colorYellow("--------------------------------------\n")
	colorYellow("TCPing started at: %v\n", p.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if p.endTime.Format(timeFormat) == nullTimeFormat {
		durationDiff = time.Since(p.startTime)
	} else {
		colorYellow("TCPing ended at:   %v\n", p.endTime.Format(timeFormat))
		durationDiff = p.endTime.Sub(p.startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	colorYellow("duration (HH:MM:SS): %v\n\n", duration.Format(hourFormat))
}

func (p *statsPlanePrinter) printRttResults(rtt *rttResults) {
	colorYellow("rtt ")
	colorGreen("min")
	colorYellow("/")
	colorCyan("avg")
	colorYellow("/")
	colorRed("max: ")
	colorGreen("%.3f", rtt.min)
	colorYellow("/")
	colorCyan("%.3f", rtt.average)
	colorYellow("/")
	colorRed("%.3f", rtt.max)
	colorYellow(" ms\n")
}

/* Print statistics when program exits */
func (p *statsPlanePrinter) printStatistics() {

	totalPackets := p.totalSuccessfulProbes + p.totalUnsuccessfulProbes
	totalUptime := calcTime(uint(p.totalUptime.Seconds()))
	totalDowntime := calcTime(uint(p.totalDowntime.Seconds()))
	packetLoss := (float32(p.totalUnsuccessfulProbes) / float32(totalPackets)) * 100

	/* general stats */
	if !p.isIP {
		colorYellow("\n--- %s (%s) TCPing statistics ---\n", p.hostname, p.ip)
	} else {
		colorYellow("\n--- %s TCPing statistics ---\n", p.hostname)
	}
	colorYellow("%d probes transmitted on port %d | ", totalPackets, p.port)
	colorYellow("%d received, ", p.totalSuccessfulProbes)

	/* packet loss stats */
	if packetLoss == 0 {
		colorGreen("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		colorLightYellow("%.2f%%", packetLoss)
	} else {
		colorRed("%.2f%%", packetLoss)
	}

	colorYellow(" packet loss\n")

	/* successful packet stats */
	colorYellow("successful probes:   ")
	colorGreen("%d\n", p.totalSuccessfulProbes)

	/* unsuccessful packet stats */
	colorYellow("unsuccessful probes: ")
	colorRed("%d\n", p.totalUnsuccessfulProbes)

	p.printLastSucUnsucProbes()

	/* uptime and downtime stats */
	colorYellow("total uptime: ")
	colorGreen("  %s\n", totalUptime)
	colorYellow("total downtime: ")
	colorRed("%s\n", totalDowntime)

	/* calculate the last longest time */
	if !p.wasDown {
		calcLongestUptime(p.stats, p.lastSuccessfulProbe)
	} else {
		calcLongestDowntime(p.stats, p.lastUnsuccessfulProbe)
	}

	/* longest uptime stats */
	p.printLongestUptime()

	/* longest downtime stats */
	p.printLongestDowntime()

	/* resolve retry stats */
	if !p.isIP {
		p.printRetryResolveStats()
	}

	rttResults := findMinAvgMaxRttTime(p.rtt)

	if rttResults.hasResults {
		p.printRttResults(&rttResults)
	}

	/* duration stats */
	p.printDurationStats()
}

/* Print TCP probe replies according to our policies */
func (p *statsPlanePrinter) printReply(replyMsg replyMsg) {
	if p.isIP {
		if replyMsg.msg == noReply {
			colorRed("%s from %s on port %d TCP_conn=%d\n",
				replyMsg.msg, p.ip, p.port, p.totalUnsuccessfulProbes)
		} else {
			colorLightGreen("%s from %s on port %d TCP_conn=%d time=%.3f ms\n",
				replyMsg.msg, p.ip, p.port, p.totalSuccessfulProbes, replyMsg.rtt)
		}
	} else {
		if replyMsg.msg == noReply {
			colorRed("%s from %s (%s) on port %d TCP_conn=%d\n",
				replyMsg.msg, p.hostname, p.ip, p.port, p.totalUnsuccessfulProbes)
		} else {
			colorLightGreen("%s from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
				replyMsg.msg, p.hostname, p.ip, p.port, p.totalSuccessfulProbes, replyMsg.rtt)
		}
	}
}

/* Print the total downtime */
func (p *statsPlanePrinter) printTotalDownTime(now time.Time) {
	latestDowntimeDuration := time.Since(p.startOfDowntime).Seconds()
	calculatedDowntime := calcTime(uint(math.Ceil(latestDowntimeDuration)))
	colorYellow("No response received for %s\n", calculatedDowntime)
}

/* Print the longest uptime */
func (p *statsPlanePrinter) printLongestUptime() {
	if p.longestUptime.duration == 0 {
		return
	}

	uptime := calcTime(uint(math.Ceil(p.longestUptime.duration)))

	colorYellow("longest consecutive uptime:   ")
	colorGreen("%v ", uptime)
	colorYellow("from ")
	colorLightBlue("%v ", p.longestUptime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", p.longestUptime.end.Format(timeFormat))
}

/* Print the longest downtime */
func (p *statsPlanePrinter) printLongestDowntime() {
	if p.longestDowntime.duration == 0 {
		return
	}

	downtime := calcTime(uint(math.Ceil(p.longestDowntime.duration)))

	colorYellow("longest consecutive downtime: ")
	colorRed("%v ", downtime)
	colorYellow("from ")
	colorLightBlue("%v ", p.longestDowntime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", p.longestDowntime.end.Format(timeFormat))
}

/* Print the number of times that we tried resolving a hostname after a failure */
func (p *statsPlanePrinter) printRetryResolveStats() {
	colorYellow("retried to resolve hostname ")
	colorRed("%d ", p.retriedHostnameResolves)
	colorYellow("times\n")
}

/* Print the message retrying to resolve */
func (p *statsPlanePrinter) printRetryingToResolve() {
	colorLightYellow("retrying to resolve %s\n", p.hostname)
}

/*

JSON output section

*/

type JSONEventType string

const (
	Start JSONEventType = "start"
	Probe JSONEventType = "probe"
	Retry JSONEventType = "retry"
	Stats JSONEventType = "stats"
)

// printJson is a shortcut for Encode() on os.Stdout.
var printJson = json.NewEncoder(os.Stdout).Encode

// JSONData contains all possible fields for JSON output.
// Because one event usually contains only a subset of fields,
// other fields will be omitted in the output.
type JSONData struct {
	// Type is a mandatory field that specifies type of a message.
	// Possible types are:
	//	- start
	// 	- probe
	// 	- retry
	// 	- stats
	Type JSONEventType `json:"type"`
	// Message contains a human-readable message.
	Message string `json:"message"`
	// Timestamp contains data when a message was sent.
	Timestamp time.Time `json:"timestamp"`

	// Optional fields below

	// Success is a special field from probe event, containing information
	// whether request was successful or not.
	// It's a pointer on purpose, otherwise success=false will be omitted,
	// but we still need to omit it for non-probe events.
	Success *bool `json:"success,omitempty"`

	// Latency in ms for a successful probe.
	Latency float32 `json:"latency,omitempty"`

	Addr                    string     `json:"addr,omitempty"`
	Hostname                string     `json:"hostname,omitempty"`
	IsIP                    *bool      `json:"is_ip,omitempty"`
	LastSuccessfulProbe     *time.Time `json:"last_successful_probe,omitempty"`
	LastUnsuccessfulProbe   *time.Time `json:"last_unsuccessful_probe,omitempty"`
	LongestUptime           float64    `json:"longest_uptime,omitempty"`
	LongestUptimeEnd        *time.Time `json:"longest_uptime_end,omitempty"`
	LongestUptimeStart      *time.Time `json:"longest_uptime_start,omitempty"`
	Port                    uint16     `json:"port,omitempty"`
	TotalPacketLoss         float32    `json:"total_packet_loss,omitempty"`
	TotalPackets            uint       `json:"total_packets,omitempty"`
	TotalSuccessfulProbes   uint       `json:"total_successful_probes,omitempty"`
	TotalUnsuccessfulProbes uint       `json:"total_unsuccessful_probes,omitempty"`
	// TotalUptime in seconds.
	TotalUptime float64 `json:"total_uptime,omitempty"`
	// TotalDowntime in seconds.
	TotalDowntime float64 `json:"total_downtime,omitempty"`
}

// TODO: remove
func jsonPrintf(format string, a ...interface{}) {
	data := struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf(format, a...),
	}
	outputJson, _ := json.Marshal(&data)
	fmt.Println(string(outputJson))
}

// printStart prints the initial message before doing probes.
func (j *statsJsonPrinter) printStart() {
	_ = printJson(JSONData{
		Type:      Start,
		Message:   fmt.Sprintf("TCPinging %s on port %d", j.hostname, j.port),
		Hostname:  j.hostname,
		Port:      j.port,
		Timestamp: time.Now(),
	})
}

// printReply prints TCP probe replies according to our policies in JSON format.
func (j *statsJsonPrinter) printReply(replyMsg replyMsg) {
	// for *bool fields
	f := false
	t := true

	data := JSONData{
		Type:      Probe,
		Addr:      j.ip.String(),
		Port:      j.port,
		IsIP:      &t,
		Success:   &f,
		Timestamp: time.Now(),
	}

	ipStr := data.Addr
	if !j.isIP {
		data.Hostname = j.hostname
		data.IsIP = &f
		ipStr = fmt.Sprintf("%s (%s)", data.Hostname, ipStr)
	}

	data.Message = fmt.Sprintf("%s from %s on port %d", replyMsg.msg, ipStr, j.port)

	if replyMsg.msg != noReply {
		data.Latency = replyMsg.rtt
		data.TotalSuccessfulProbes = j.totalSuccessfulProbes
		data.Success = &t
	} else {
		data.TotalUnsuccessfulProbes = j.totalUnsuccessfulProbes
	}

	_ = printJson(data)
}

/* Print the start and end time of the program in JSON format */
func (j *statsJsonPrinter) printDurationStats() {
	var duration time.Time
	var durationDiff time.Duration
	endMsg := "still running"

	startMSg := fmt.Sprintf("started at: %v ", j.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if j.endTime.Format(timeFormat) == nullTimeFormat {
		durationDiff = time.Since(j.startTime)
	} else {
		endMsg = fmt.Sprintf("ended at: %v ", j.endTime.Format(timeFormat))
		durationDiff = j.endTime.Sub(j.startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	durationFormatted := fmt.Sprintf("duration (HH:MM:SS): %v", duration.Format(hourFormat))

	jsonPrintf(startMSg + endMsg + durationFormatted)
}

// printStatistics prints all gathered stats when program exits.
func (j *statsJsonPrinter) printStatistics() {
	rttResults := findMinAvgMaxRttTime(j.rtt)
	if !rttResults.hasResults {
		return
	}

	data := JSONData{
		Type:      Stats,
		Hostname:  j.hostname,
		Timestamp: time.Now(),
	}

	totalPackets := j.totalSuccessfulProbes + j.totalUnsuccessfulProbes
	packetLoss := (float32(j.totalUnsuccessfulProbes) / float32(totalPackets)) * 100

	data.TotalPacketLoss = packetLoss
	data.TotalPackets = totalPackets
	data.TotalSuccessfulProbes = j.totalSuccessfulProbes
	data.TotalUnsuccessfulProbes = j.totalUnsuccessfulProbes
	if !j.lastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = &j.lastSuccessfulProbe
	}
	if !j.lastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = &j.lastUnsuccessfulProbe
	}
	data.TotalUptime = j.totalUptime.Seconds()
	data.TotalDowntime = j.totalDowntime.Seconds()

	/* calculate the last longest time */
	if !j.wasDown {
		calcLongestUptime(j.stats, j.lastSuccessfulProbe)
	} else {
		calcLongestDowntime(j.stats, j.lastUnsuccessfulProbe)
	}

	if j.longestUptime.duration != 0 {
		data.LongestUptime = j.longestUptime.duration
		data.LongestUptimeStart = &j.longestUptime.start
		data.LongestUptimeEnd = &j.longestUptime.end
	}

	/* longest downtime stats */
	j.printLongestDowntime()

	/* resolve retry stats */
	if !j.isIP {
		j.printRetryResolveStats()
	}

	/* latency stats.*/
	jsonPrintf("rtt min/avg/max: %.3f/%.3f/%.3f", rttResults.min, rttResults.average, rttResults.max)

	/* duration stats */
	j.printDurationStats()

	_ = printJson(data)
}

/* Print the total downtime in JSON format */
func (j *statsJsonPrinter) printTotalDownTime(now time.Time) {
	latestDowntimeDuration := time.Since(j.startOfDowntime).Seconds()
	calculatedDowntime := calcTime(uint(math.Ceil(latestDowntimeDuration)))

	jsonPrintf("No response received for %s", calculatedDowntime)
}

/* Print the longest downtime in JSON format */
func (j *statsJsonPrinter) printLongestDowntime() {
	if j.longestDowntime.duration == 0 {
		return
	}

	downtime := calcTime(uint(math.Ceil(j.longestDowntime.duration)))

	longestDowntimeStart := j.longestDowntime.start.Format(timeFormat)
	longestDowntimeEnd := j.longestDowntime.end.Format(timeFormat)

	jsonPrintf("longest consecutive downtime: %v from %v to %v", downtime, longestDowntimeStart, longestDowntimeEnd)
}

/* Print the number of times that we tried resolving a hostname after a failure in JSON format */
func (j *statsJsonPrinter) printRetryResolveStats() {
	jsonPrintf("retried to resolve hostname %d times", j.retriedHostnameResolves)
}

// printRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (j *statsJsonPrinter) printRetryingToResolve() {
	_ = printJson(JSONData{
		Type:      Retry,
		Message:   fmt.Sprintf("retrying to resolve %s", j.hostname),
		Hostname:  j.hostname,
		Timestamp: time.Now(),
	})
}
