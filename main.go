package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type UrlHit struct {
	Url  string
	Hits int
}

//list of struct containing hits for urls.
type urlHitList []UrlHit

// sort.Interface implementation to make this list sortable by hits
func (l urlHitList) Len() int {
	return len(l)
}

func (l urlHitList) Less(i, j int) bool {
	return l[i].Hits > l[j].Hits //We want decreasing order, so "less is more" in this case.
}

func (l urlHitList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

//Computes a key we can use to represent the date which is also easily sortable
func getDateKey(epochTs string) (int, error) {
	epochTsInt, err := strconv.ParseInt(epochTs, 10, 64)
	if err != nil {
		return 0, err
	}
	t := time.Unix(epochTsInt, 0).UTC()
	key, err := strconv.Atoi(t.Format("20060102")) //Format is YYYYMMDD to make sorting easy while still parsable
	if err != nil {
		return 0, err
	}
	return key, nil
}

//Convert Date Key to formated string
func formatDateFromKey(dateKey int) (string, error) {
	dateKeyStr := strconv.Itoa(dateKey)
	if len(dateKeyStr) != 8 {
		return "", errors.New("Invalid date key: " + dateKeyStr)
	}
	return fmt.Sprintf("%s/%s/%s GMT", dateKeyStr[4:6], dateKeyStr[6:], dateKeyStr[:4]), nil
}

//Parses the data line by line and returns the nested mapped data and a sorted list of dates
func parseData(scanner *bufio.Scanner) (map[int]map[string]int, []int, error) {
	dateUrlHits := make(map[int]map[string]int)
	var dates []int
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "|")
		if len(parts) != 2 {
			continue //If we wanted to add debug level, maybe print a message saying this line could not be parsed. But might be noisy for a big dataset
		}
		ts := parts[0]
		url := parts[1]
		d, err := getDateKey(ts)
		if err != nil {
			continue //once again just drop the data if it's not parsable
		}
		if uh, ok := dateUrlHits[d]; ok {
			if _, ok2 := uh[url]; ok2 {
				uh[url]++
			} else {
				uh[url] = 1
			}
		} else {
			dateUrlHits[d] = make(map[string]int)
			dateUrlHits[d][url] = 1
			dates = append(dates, d)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	sort.Ints(dates)

	return dateUrlHits, dates, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("One argument is needed! (path to input file)")
		os.Exit(1)
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	parsed, dates, err := parseData(scanner)
	if err != nil {
		log.Fatal(err)
	}

	for _, date := range dates {
		formatedDate, err := formatDateFromKey(date)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(formatedDate)

		var urlHits urlHitList
		for url, hits := range parsed[date] {
			urlHits = append(urlHits, UrlHit{url, hits})
		}
		sort.Sort(urlHits)
		for _, uh := range urlHits {
			fmt.Printf("%s %d\n", uh.Url, uh.Hits)
		}
	}
}
