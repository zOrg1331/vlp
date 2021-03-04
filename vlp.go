package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const DateTimeR = `L (?P<date>\d{2}/\d{2}/\d{4}) - (?P<time>\d{2}:\d{2}:\d{2}): `
const MapR = DateTimeR + `Loading map "(?P<map>\w+)"`
const ConnR = DateTimeR + `"(?P<nick>\w+)<\d+><STEAM_\d+:\d+:\d+><\d*>" connected, address .*`
const DisconnR = DateTimeR + `"(?P<nick>\w+)<\d+><STEAM_\d+:\d+:\d+><\d*>" disconnected`
const KillR = DateTimeR + `"(?P<nick>\w+)<\d+><STEAM_\d+:\d+:\d+><\d*>" killed "(?P<victim>\w+)<\d+><STEAM_\d+:\d+:\d+><\d*>" with "(?P<weapon>\w+)"`
const SuicideR = DateTimeR + `"(?P<nick>\w+)<\d+><STEAM_\d+:\d+:\d+><\d*>" committed suicide with "(?P<weapon>\w+)"`
const TimestampR = DateTimeR + `.*`

type Kill struct {
	Killer string
	Victim string
	Weapon string
	Ts     time.Time
}

type Suicide struct {
	Victim string
	Weapon string
	Ts     time.Time
}

type Player struct {
	Nick        string
	Connects    []time.Time
	Disconnects []time.Time
	Playtime    int

	KillsOfPlayer    map[string][]Kill
	KillsWithWeapon  map[string][]Kill
	DeathsFromWeapon map[string][]Kill
	KillsCount       int
	SuicidesCount    int
}

type Stats struct {
	Map         string
	Players     map[string]Player
	PlayersList []string
	Kills       []Kill
	Suicides    []Suicide
	Weapons     []string
	LastTs      time.Time
}

func main() {
	flagLogfile := flag.String("logfile", "L0226001.log", "path to logfile")

	flag.Parse()

	log.SetLevel(log.DebugLevel)

	logfile, err := os.Open(*flagLogfile)
	if err != nil {
		log.Fatalf("failed to open logfile: %v", err)
	}
	defer logfile.Close()

	stats := NewStats()

	// preparing regexps
	var regexps map[string]*regexp.Regexp
	regexps = make(map[string]*regexp.Regexp)

	regexps["map"] = regexp.MustCompile(MapR)
	regexps["conn"] = regexp.MustCompile(ConnR)
	regexps["disconn"] = regexp.MustCompile(DisconnR)
	regexps["kill"] = regexp.MustCompile(KillR)
	regexps["suicide"] = regexp.MustCompile(SuicideR)
	regexps["timestamp"] = regexp.MustCompile(TimestampR)

	// parse each line of the log file
	logScanner := bufio.NewScanner(logfile)
	for logScanner.Scan() {
		processLogEntry(stats, logScanner.Text(), regexps)
	}

	if err := logScanner.Err(); err != nil {
		log.Fatalf("failed to read logfile: %v", err)
	}

	calcStats(stats)
}

func NewStats() *Stats {
	return &Stats{
		Players: make(map[string]Player),
	}
}

func processLogEntry(stats *Stats, line string, regexps map[string]*regexp.Regexp) {
	checkMap(stats, line, regexps["map"])
	checkPlayerConnected(stats, line, regexps["conn"])
	checkPlayerDisconnected(stats, line, regexps["disconn"])
	checkKillEvent(stats, line, regexps["kill"])
	checkSuicideEvent(stats, line, regexps["suicide"])
	storeLastTimestamp(stats, line, regexps["timestamp"])
}

func parseTimestamp(d string, t string) time.Time {
	ts, _ := time.Parse("01/02/2006 15:04:05", d+" "+t)
	return ts
}

func checkMap(stats *Stats, line string, r *regexp.Regexp) {
	if m := r.FindStringSubmatch(line); m != nil {
		stats.Map = m[r.SubexpIndex("map")]
	}
}

func checkPlayerConnected(stats *Stats, line string, r *regexp.Regexp) {
	if m := r.FindStringSubmatch(line); m != nil {
		nick := m[r.SubexpIndex("nick")]
		date := m[r.SubexpIndex("date")]
		time := m[r.SubexpIndex("time")]

		addPlayerConnected(stats, nick, parseTimestamp(date, time))
	}
}

func checkPlayerDisconnected(stats *Stats, line string, r *regexp.Regexp) {
	if m := r.FindStringSubmatch(line); m != nil {
		nick := m[r.SubexpIndex("nick")]
		date := m[r.SubexpIndex("date")]
		time := m[r.SubexpIndex("time")]

		addPlayerDisconnected(stats, nick, parseTimestamp(date, time))
	}
}

func checkKillEvent(stats *Stats, line string, r *regexp.Regexp) {
	if m := r.FindStringSubmatch(line); m != nil {
		date := m[r.SubexpIndex("date")]
		time := m[r.SubexpIndex("time")]
		nick := m[r.SubexpIndex("nick")]
		victim := m[r.SubexpIndex("victim")]
		weapon := m[r.SubexpIndex("weapon")]

		addKillEvent(stats, nick, parseTimestamp(date, time), victim, weapon)
	}
}

func checkSuicideEvent(stats *Stats, line string, r *regexp.Regexp) {
	if m := r.FindStringSubmatch(line); m != nil {
		date := m[r.SubexpIndex("date")]
		time := m[r.SubexpIndex("time")]
		nick := m[r.SubexpIndex("nick")]
		weapon := m[r.SubexpIndex("weapon")]

		addSuicideEvent(stats, nick, parseTimestamp(date, time), weapon)
	}
}

func storeLastTimestamp(stats *Stats, line string, r *regexp.Regexp) {
	if m := r.FindStringSubmatch(line); m != nil {
		date := m[r.SubexpIndex("date")]
		time := m[r.SubexpIndex("time")]

		stats.LastTs = parseTimestamp(date, time)
	}
}

func createNewPlayer(nick string) Player {
	p := Player{
		Nick: nick,
	}
	p.KillsOfPlayer = make(map[string][]Kill)
	p.KillsWithWeapon = make(map[string][]Kill)
	p.DeathsFromWeapon = make(map[string][]Kill)
	return p
}

func addPlayerConnected(stats *Stats, nick string, ts time.Time) {
	// check if we already know this player
	// if yes, just add new connect timestamp
	// if no, create a player instance
	player, exists := stats.Players[nick]
	if !exists {
		player = createNewPlayer(nick)
	}

	player.Connects = append(player.Connects, ts)

	stats.Players[nick] = player
}

func addPlayerDisconnected(stats *Stats, nick string, ts time.Time) {
	// we presume that player exists
	player, exists := stats.Players[nick]
	if !exists {
		log.Errorf("got disconnected event for an unknown player: %v", nick)
		return
	}
	player.Disconnects = append(player.Disconnects, ts)

	stats.Players[nick] = player
}

func addKillEvent(stats *Stats, nick string, ts time.Time, victim string, weapon string) {
	k := Kill{
		Killer: nick,
		Victim: victim,
		Weapon: weapon,
		Ts:     ts,
	}
	stats.Kills = append(stats.Kills, k)

	// we presume that player exists
	player, exists := stats.Players[nick]
	if !exists {
		log.Errorf("got kill event for an unknown player: %v", nick)
		return
	}

	// calc some stats to save processing time
	player.KillsCount += 1
	player.KillsOfPlayer[victim] = append(player.KillsOfPlayer[victim], k)
	player.KillsWithWeapon[weapon] = append(player.KillsWithWeapon[weapon], k)

	stats.Players[victim].DeathsFromWeapon[weapon] = append(stats.Players[victim].DeathsFromWeapon[weapon], k)

	stats.Players[nick] = player
}

func addSuicideEvent(stats *Stats, victim string, ts time.Time, weapon string) {
	stats.Suicides = append(stats.Suicides, Suicide{
		Victim: victim,
		Weapon: weapon,
		Ts:     ts,
	})

	// we presume that player exists
	player, exists := stats.Players[victim]
	if !exists {
		log.Errorf("got suicide event for an unknown player: %v", victim)
		return
	}

	// calc some stats to save processing time
	player.SuicidesCount += 1

	stats.Players[victim] = player
}

func calcPlaytime(player Player, lastTs time.Time) int64 {
	// in case disconnected event is not recorded
	// calculate playtime using the last recorded timestamp
	var playtime int64

	for i, ts := range player.Connects {
		if i < len(player.Disconnects) {
			playtime += int64(player.Disconnects[i].Sub(ts).Seconds())
		} else {
			playtime += int64(lastTs.Sub(ts).Seconds())
			break
		}
	}

	return playtime
}

func calcWhoKillsWhom(stats *Stats, writer io.Writer) {
	csvW := csv.NewWriter(writer)

	// the first line is victims
	csvW.Write(append([]string{"who"}, stats.PlayersList...))

	// the first column is killers
	for _, p := range stats.PlayersList {
		var csvLine []string
		csvLine = append(csvLine, p)

		// victims list
		for _, v := range stats.PlayersList {
			if p == v {
				// diagonal -- suicides
				csvLine = append(csvLine, strconv.Itoa(stats.Players[p].SuicidesCount))
			} else {
				// player p kills victim v
				csvLine = append(csvLine, strconv.Itoa(len(stats.Players[p].KillsOfPlayer[v])))
			}
		}
		csvW.Write(csvLine)
	}
	csvW.Flush()
}

func calcWhoKillsWithWhat(stats *Stats, writer io.Writer) {
	csvW := csv.NewWriter(writer)

	// the first line -- players
	csvW.Write(append([]string{"what"}, stats.PlayersList...))

	// the left column -- weapons
	for _, w := range stats.Weapons {
		var csvLine []string
		csvLine = append(csvLine, w)

		for _, p := range stats.PlayersList {
			csvLine = append(csvLine, strconv.Itoa(len(stats.Players[p].KillsWithWeapon[w])))
		}
		csvW.Write(csvLine)
	}
	csvW.Flush()
}

func calcWhoIsKilledOfWhat(stats *Stats, writer io.Writer) {
	csvW := csv.NewWriter(writer)

	// the first line -- players
	csvW.Write(append([]string{"what"}, stats.PlayersList...))

	// the left column -- weapons
	for _, w := range stats.Weapons {
		var csvLine []string
		csvLine = append(csvLine, w)

		for _, p := range stats.PlayersList {
			csvLine = append(csvLine, strconv.Itoa(len(stats.Players[p].DeathsFromWeapon[w])))
		}
		csvW.Write(csvLine)
	}
	csvW.Flush()
}

func fillPlayersList(stats *Stats) {
	// map keys return in random order, save players as a static array
	for _, p := range stats.Players {
		stats.PlayersList = append(stats.PlayersList, p.Nick)
	}
}

func collectUsedWeapons(stats *Stats) {
	// form a list of weapons used throughout a fight
	var weaponsMap map[string]int
	weaponsMap = make(map[string]int)
	for _, p := range stats.Players {
		for w := range p.KillsWithWeapon {
			weaponsMap[w] = 0
		}
	}

	for k := range weaponsMap {
		stats.Weapons = append(stats.Weapons, k)
	}
}

func calcStats(stats *Stats) {
	fillPlayersList(stats)
	collectUsedWeapons(stats)

	log.Infof("map played: %v", stats.Map)
	for _, p := range stats.Players {
		log.Infof("%v summary:", p.Nick)

		log.Infof("\t playtime: %vs", calcPlaytime(p, stats.LastTs))
		log.Infof("\t frags: %v", p.KillsCount)
		log.Infof("\t suicides: %v", p.SuicidesCount)
	}
	log.Infof("who kills whom:")
	calcWhoKillsWhom(stats, os.Stdout)
	log.Infof("who kills with what:")
	calcWhoKillsWithWhat(stats, os.Stdout)
	log.Infof("who is killed of what:")
	calcWhoIsKilledOfWhat(stats, os.Stdout)
}
