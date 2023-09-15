package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/yz3358/clash-ctl/common"
	"log"
	"math"
	"net/url"
	"os"
	"sort"
	"sync"
)

var currentSelector SelectorTable
var maxRendered = 60

var (
	markTrue  = "✓"
	markFalse = "✗"
)

var (
	ErrSelectorNotInitialized = errors.New("the selector table is not initialized")
)

type SelectorTable struct {
	Selector Proxy
	Proxies  []Proxy
}

// Render the table
func (s SelectorTable) Render() (string, error) {
	if s.Selector.Name == "" {
		return "", ErrSelectorNotInitialized
	}

	// fmt.Println("proxy now", s.Selector.Now)

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Id", "Selector", "Proxy Name", "Delay"})

	var rows []table.Row
	for id, proxy := range s.Proxies {
		if id > maxRendered {
			break
		}

		delay := markFalse
		formatter := text.FgRed

		if d := proxy.LastestDelay(); d > 0 {
			formatter = text.FgGreen
			if d > 500 {
				formatter = text.FgYellow
			}
			delay = fmt.Sprintf("%vms", d)
		}

		idStr := fmt.Sprintf("%v", id)
		if s.Selector.Now == proxy.Name {
			idStr = fmt.Sprintf("%v <-", id)
		}

		delay = formatter.Sprint(delay)

		proxyName := proxy.Name
		if proxy.Now != "" {
			proxyName = fmt.Sprintf("%s --> %s", proxyName, proxy.Now)
		}

		rows = append(rows, []any{idStr, s.Selector.Name, proxyName, delay})
	}
	t.AppendRows(rows)

	return t.Render(), nil
}

// Use proxy based on id
func (s SelectorTable) Use(id int) error {
	proxy := s.findProxy(id)
	if proxy == nil {
		return errors.New("id out of range")
	}

	server, err := defaultServer()
	if err != nil {
		return err
	}

	body := map[string]any{
		"name": proxy.NameEncoded(),
	}
	group := url.PathEscape(s.Selector.NameEncoded())

	req := common.MakeRequest(*server)
	fail := common.HTTPError{}
	resp, err := req.R().SetError(&fail).SetBody(body).Put("/proxies/" + group)
	if err != nil {
		return fmt.Errorf("%s", text.FgRed.Sprint(err.Error()))
	}

	if resp.IsError() {
		return fmt.Errorf("%s", text.FgRed.Sprint(fail.Message))
	}

	fmt.Println(text.FgGreen.Sprint("proxy switched", markTrue), proxy.Name)

	return nil
}

func (s SelectorTable) findProxy(id int) *Proxy {
	if id <= 0 {
		id = 0
	} else if id >= len(s.Proxies) {
		return nil
	}

	return &(s.Proxies[id])
}

func (s SelectorTable) BenchMark() {
	if s.Selector.Name == "" {
		fmt.Println(text.FgRed.Sprint(ErrSelectorNotInitialized.Error()))
		return
	}

	var wg sync.WaitGroup
	for _, proxy := range s.Proxies {
		wg.Add(1)
		go func(p Proxy) {
			err := getProxyDelay(p)
			if err != nil {
				fmt.Println(p.Name, err.Error())
			}
			wg.Done()
		}(proxy)
	}

	wg.Wait()

	log.Println(text.FgGreen.Sprint("benchmark test finished\n"))

	GetSelectorTable()
	currentSelector.Render()
}

type ProxyList []Proxy

func (l ProxyList) Len() int {
	return len(l)
}
func (l ProxyList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l ProxyList) Less(i, j int) bool {
	a, b := l[i].LastestDelay(), l[j].LastestDelay()
	if a == 0 {
		a = math.MaxInt
	}

	if b == 0 {
		b = math.MaxInt
	}

	return a < b
}

func GetSelectorTable() (*SelectorTable, error) {
	var proxyList ProxyList

	proxies, err := GetProxies()
	if err != nil {
		return nil, err
	}

	// --- get rule-based selector
	var selector *Proxy
	for _, proxy := range proxies {
		if proxy.Type == ProxyTypeSelector && proxy.Name != ProxyNameGlobal {
			selector = &proxy
			break
		}
	}

	if selector == nil {
		return nil, errors.New("no rule-based selector found")
	}

	// --- select matched proxies
	for _, name := range selector.All {
		// proxy := proxies[name]
		proxyList = append(proxyList, proxies[name])
	}

	// --- sort them
	sort.Sort(proxyList)

	currentSelector = SelectorTable{
		Selector: *selector,
		Proxies:  proxyList,
	}

	return &currentSelector, nil
}

func getProxyDelay(proxy Proxy) error {
	server, err := defaultServer()
	if err != nil {
		return err
	}

	req := common.MakeRequest(*server)
	fail := common.HTTPError{}
	resp, err := req.R().SetError(&fail).SetQueryParams(map[string]string{
		"timeout": "3000",
		// "url":     "https://www.google.com",
		"url": "http://cp.cloudflare.com/generate_204",
	}).Get("/proxies/" + proxy.NameEncoded() + "/delay")
	if err != nil {
		return fmt.Errorf("%s", text.FgRed.Sprint(err.Error()))
	}

	if resp.IsError() {
		return fmt.Errorf("%s", text.FgRed.Sprint(fail.Message))
	}

	type body struct {
		Delay int `json:"delay"`
	}
	var b body
	_ = json.Unmarshal(resp.Body(), &b)

	fmt.Println(proxy.Name, text.FgGreen.Sprintf("%vms", b.Delay))

	return nil
}
