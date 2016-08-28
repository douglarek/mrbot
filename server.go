package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
)

var (
	config, gitlab, token string
)

type event struct {
	Ok   string `json:"object_kind"`
	User struct {
		Username string `json:"username"`
	} `json:"user"`
	Project struct {
		Namespace string `json:"namespace"`
	} `json:"project"`
	Oa struct {
		Pid    int    `json:"target_project_id"`
		Mid    int    `json:"id"`
		Action string `json:"action"`
		Source struct {
			Path string `json:"path_with_namespace"`
		} `json:"source"`
	} `json:"object_attributes"`
}

func init() {
	flag.StringVar(&config, "config", "", "code review config url")
	flag.StringVar(&gitlab, "gitlab", "https://gitlab.com", "gitlab root url")
	flag.StringVar(&token, "token", "", "private token")
}

func mr(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	var e event
	if json.NewDecoder(r.Body).Decode(&e) == nil {
		if e.Ok == "merge_request" {
			if a, p := e.Oa.Action, e.Oa.Source.Path; a == "open" || a == "reopen" {
				api := fmt.Sprintf("%s/api/v3/projects/%v/merge_requests/%v/notes", gitlab, e.Oa.Pid, e.Oa.Mid)
				cc := "cc @" + e.Project.Namespace
				if len(config) != 0 {
					if m, err := parseConfig(config); err == nil {
						cc = rebuildString(m[p], e.User.Username)
					}
				}
				go comment(api, cc)
			}

		}
	}
	fmt.Fprintf(w, "%s", "OK")
}

func rebuildString(s, exclude string) string {
	s = strings.Replace(s, " ", "", -1)
	var buf bytes.Buffer
	for _, i := range strings.Split(s, ",") {
		if i == exclude {
			continue
		}
		buf.WriteString("@")
		buf.WriteString(i)
		buf.WriteString(", ")
	}
	buf.Truncate(buf.Len() - 2)
	return buf.String()
}

func parseConfig(config string) (map[string]string, error) {
	resp, err := http.Get(config)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	m := map[string]string{}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func comment(api, cc string) {
	req, err := http.NewRequest("POST", api, strings.NewReader("body=cc "+cc))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, _ := (&http.Client{}).Do(req)
	defer resp.Body.Close()
}

func main() {
	flag.Parse()

	http.HandleFunc("/mr", mr)
	http.ListenAndServe(":8080", nil)
}
