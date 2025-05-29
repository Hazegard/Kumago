package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Titles struct {
	Config struct {
		Slug                  string      `json:"slug"`
		Title                 string      `json:"title"`
		Description           string      `json:"description"`
		Icon                  string      `json:"icon"`
		Theme                 string      `json:"theme"`
		Published             bool        `json:"published"`
		ShowTags              bool        `json:"showTags"`
		CustomCSS             string      `json:"customCSS"`
		FooterText            interface{} `json:"footerText"`
		ShowPoweredBy         bool        `json:"showPoweredBy"`
		GoogleAnalyticsId     interface{} `json:"googleAnalyticsId"`
		ShowCertificateExpiry bool        `json:"showCertificateExpiry"`
	} `json:"config"`
	Incident        []KumaGroup `json:"incident"`
	PublicGroupList []KumaGroup `json:"publicGroupList"`
	MaintenanceList []KumaGroup `json:"maintenanceList"`
}

type KumaGroup struct {
	Id          int            `json:"id"`
	Name        string         `json:"name"`
	Weight      int            `json:"weight"`
	MonitorList []MonitorTitle `json:"monitorList"`
}

type MonitorTitle struct {
	GroupId   int    `json:"group"`
	GroupName string `json:"groupName"`
	Id        int    `json:"id"`
	Name      string `json:"name"`
	sendUrl   string
	Type      string `json:"type"`
}

func CheckAvailability(url *url.URL) bool {
	r, err := http.Head(fmt.Sprintf("%s/dashboard", url))
	if err != nil {
		return false
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return false
	}
	return true
}

func GetTitleDict(dashboardName string, url *url.URL) (map[string]MonitorTitle, error) {
	r, err := http.Get(fmt.Sprintf("%s/status/%s", url, dashboardName))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	re := regexp.MustCompile(`window\.preloadData\s*=\s*(\{.*\});`)
	matches := re.FindStringSubmatch(string(body))
	if matches == nil {
		return nil, fmt.Errorf("unable to get dashboard %s", dashboardName)
	}
	if len(matches) < 2 {
		return nil, fmt.Errorf("unable to get dashboard %s", dashboardName)
	}
	content := strings.ReplaceAll(matches[1], "'", "\"")
	titles := Titles{}
	err = json.Unmarshal([]byte(content), &titles)
	if err != nil {
		return nil, err
	}
	monitorTitles := make(map[string]MonitorTitle)
	for _, group := range titles.MaintenanceList {
		for _, t := range group.MonitorList {
			t.GroupId = group.Id
			t.GroupName = strings.TrimSpace(group.Name)
			monitorTitles[strconv.Itoa(t.Id)] = t
		}
	}
	for _, group := range titles.Incident {
		for _, t := range group.MonitorList {
			t.GroupId = group.Id
			t.GroupName = strings.TrimSpace(group.Name)
			monitorTitles[strconv.Itoa(t.Id)] = t
		}
	}
	for _, group := range titles.PublicGroupList {
		for _, t := range group.MonitorList {
			t.GroupId = group.Id
			t.GroupName = strings.TrimSpace(group.Name)
			monitorTitles[strconv.Itoa(t.Id)] = t
		}
	}
	return monitorTitles, nil
}

func GetDashboard(dashboardName string, titles map[string]MonitorTitle, url *url.URL) (HeartBeatList, error) {
	r, err := http.Get(fmt.Sprintf("%s/api/status-page/heartbeat/%s", url, dashboardName))
	if err != nil {
		return HeartBeatList{}, err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return HeartBeatList{}, err
	}

	dashboard := Dashboard{}

	err = json.Unmarshal(body, &dashboard)
	if err != nil {
		return HeartBeatList{}, err
	}

	hblist := make(HeartBeatList)
	for monitorId, status := range dashboard.HeartBeat {
		group := Group{
			Id:   titles[monitorId].GroupId,
			Name: titles[monitorId].GroupName,
		}
		monitor := &Monitor{
			Id:     monitorId,
			Name:   strings.TrimSpace(titles[monitorId].Name),
			Status: status,
		}
		hblist[group] = append(hblist[group], monitor)
	}
	return hblist, nil
}
