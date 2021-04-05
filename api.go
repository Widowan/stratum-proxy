/*
API - API functions.
*/
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	// "fmt"
)

/*
Pool - the pool data for creating of the user.
*/
type Pool struct {
	Pool     string `json:"pool"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type UserToChange struct {
	User string `json:"user"`
}

type ShortWorker struct {
	Ua       string
	Id       string
	Addr     string
	User     string
	Hash     string
	PoolAddr string
	Hashrate float64
}

/*
API - API functions.
*/
type API struct{}

/*
ServeHTTP - web handler.
*/
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	err = nil
	u := r.URL
	w.Header().Set("Content-Type", "application/json")
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	if !strings.Contains(u.Path, "workers") {
		LogInfo("New connection on %s using %s, from %s", "", u.Path, r.Method, r.Header.Get(("Origin")))
	}
	if u.Path == "/api/v1/users" && (r.Method == "POST" || r.Method == "PUT") {
		var p Pool
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&p)
		fmt.Printf("%+v\n", p)
		if err == nil {
			LogInfo("proxy : API request to add user with pool %s and credentials %s:%s", "", p.Pool, p.User, p.Password)
			var user *User
			user, err = db.GetUserByPool(p.Pool, p.User)
			if err == nil {
				if user == nil {
					LogInfo("proxy : user not found and will be added", "")
					user = new(User)
					err = user.Init(p.Pool, p.User, p.Password)
				}
				if err == nil {
					LogInfo("proxy : user successfully created with name %s", "", user.name)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"name": "` + user.name + `", "error": ""}`))
				}
			}
		}
	} else if u.Path == "/api/v1/changepool" && (r.Method == "POST" || r.Method == "PUT") {
		LogInfo("proxy : request to change pool", "")
		var us UserToChange
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&us)
		if err == nil {
			for _, v := range workers.Workers {
				err = v.Auth(us.User, "")
				if err != nil {
					break
				}
			}
			w.WriteHeader(http.StatusOK)
			if err == nil {
				w.Write([]byte(`{"error":""}`))
				LogInfo("proxy : pool changed successfully", "")
			} else {
				w.Write([]byte(`{"error": "` + err.Error() + `"}`))
				LogInfo("proxy : pool change  failed", "")
			}
		}
	} else if u.Path == "/api/v1/getallusers" {
		users, err := db.GetAllUsersShort()
		if err != nil {
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		}
		j, err := json.Marshal(&users)
		if err != nil {
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		}
		w.Write([]byte(j))
	} else if u.Path == "/api/v1/getallworkers" {
		var result []ShortWorker
		var temp ShortWorker
		for workerid := range workers.Workers {
			worker := workers.get(workerid)
			temp.Ua = worker.Ua
			temp.Id = worker.Id
			temp.Addr = worker.Addr
			temp.User = worker.User
			temp.Hash = worker.Hash
			temp.PoolAddr = worker.Pool.Addr
			temp.Hashrate = worker.Hashrate
			result = append(result, temp)
		}
		j, err := json.Marshal(result)
		if err != nil {
			w.Write([]byte(`{"workers":[],"error": "` + err.Error() + `"}`))
		} else {
			// Always crashes with NPE upon j or result check.
			w.Write([]byte(j))
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "command not found"}`))
	}
	if err != nil {
		LogError("proxy : API error: %s", "", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	}
}
