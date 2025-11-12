package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	var ui string
	flag.StringVar(&ui, "ui", "http://localhost:8080", "UI base URL")
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("usage: clustectl [nodes|vms|volumes|networks|storagepools|config|audit|metrics] ...")
		return
	}
	switch os.Args[1] {
	case "nodes":
		resp, err := http.Get(ui + "/api/nodes")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		var nodes []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
			panic(err)
		}
		fmt.Println("NODES:")
		for _, n := range nodes {
			fmt.Printf("- %s (%s) %s\n", n["name"], n["addr"], n["status"])
		}
	case "vms":
		if len(os.Args) == 2 {
			resp, err := http.Get(ui + "/api/vms")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var vms map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&vms); err != nil {
				panic(err)
			}
			for id, v := range vms {
				fmt.Printf("- %s: %v\n", id, v)
			}
		} else if os.Args[2] == "upsert" {
			// expects JSON on stdin
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/vms/upsert", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "delete" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/vms/delete", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "clone" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/vms/clone", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "migrate" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/vms/migrate", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "snapshot" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/vms/snapshot", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		}
	case "networks":
		if len(os.Args) == 2 {
			resp, err := http.Get(ui + "/api/networks")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var nets map[string]map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&nets); err != nil {
				panic(err)
			}
			for id, n := range nets {
				fmt.Printf("- %s: %v\n", id, n)
			}
		} else if os.Args[2] == "upsert" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/networks", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "delete" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/networks/delete", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		}
	case "storagepools":
		if len(os.Args) == 2 {
			resp, err := http.Get(ui + "/api/storagepools")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var pools map[string]map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&pools); err != nil {
				panic(err)
			}
			for id, p := range pools {
				fmt.Printf("- %s: %v\n", id, p)
			}
		} else if os.Args[2] == "upsert" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/storagepools", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "delete" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/storagepools/delete", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		}
	case "volumes":
		if len(os.Args) == 2 {
			resp, err := http.Get(ui + "/api/volumes")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var vols map[string]map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&vols); err != nil {
				panic(err)
			}
			for id, v := range vols {
				fmt.Printf("- %s: %v\n", id, v)
			}
		} else if os.Args[2] == "upsert" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/volumes", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "delete" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/volumes/delete", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		}
	case "config":
		if len(os.Args) == 2 || (len(os.Args) > 2 && os.Args[2] == "get") {
			resp, err := http.Get(ui + "/api/config")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var cfg map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
				panic(err)
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(cfg)
		} else if os.Args[2] == "set" {
			var cfg map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&cfg); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(cfg)
			resp, err := http.Post(ui+"/api/config", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "history" {
			resp, err := http.Get(ui + "/api/config/history")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var hist []map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&hist); err != nil {
				panic(err)
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(hist)
		} else if os.Args[2] == "version" {
			resp, err := http.Get(ui + "/api/config/version")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var v map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
				panic(err)
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(v)
		} else if os.Args[2] == "rollback" {
			resp, err := http.Post(ui+"/api/config/rollback", "application/json", bytes.NewReader([]byte("{}")))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		}
	case "audit":
		resp, err := http.Get(ui + "/api/audit")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		io.Copy(os.Stdout, resp.Body)
	case "metrics":
		resp, err := http.Get(ui + "/metrics")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		io.Copy(os.Stdout, resp.Body)
	case "templates":
		if len(os.Args) == 2 {
			resp, err := http.Get(ui + "/api/templates")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			var tpls map[string]map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&tpls); err != nil {
				panic(err)
			}
			for id, t := range tpls {
				fmt.Printf("- %s: %v\n", id, t)
			}
		} else if os.Args[2] == "upsert" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/templates", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "delete" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/templates/delete", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		} else if os.Args[2] == "instantiate" {
			var body map[string]any
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				panic(err)
			}
			b, _ := json.Marshal(body)
			resp, err := http.Post(ui+"/api/vms/cloneFromTemplate", "application/json", bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			_ = resp.Body.Close()
			fmt.Println("ok")
		}
	}
}
