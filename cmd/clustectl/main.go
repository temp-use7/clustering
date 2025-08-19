package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var ui string
	flag.StringVar(&ui, "ui", "http://localhost:8080", "UI base URL")
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("usage: clustectl [nodes|vms|networks|storagepools|config] ...")
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
		}
	}
}
