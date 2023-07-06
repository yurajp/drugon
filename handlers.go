package main

import (
	"net/http"
	"os"
	"encoding/json"
)

func display(w http.ResponseWriter, r *http.Request) {
  err := PrepareData()
  if err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  drl := med.makeList()
  err = dtemp.Execute(w, drl)
  if err != nil {
    fmt.Println(" execute: ", err)
  }
}

func deleteD(w http.ResponseWriter, r *http.Request) {
  nm := r.URL.Query().Get("name")
  newb := []Drug{}
  for _, d := range med.Box {
    if d.Name != nm {
      newb = append(newb, d)
    }
  }
  med = Medicine{newb}
  err := med.WriteData()
  if err != nil {
    fmt.Println(" write data: ", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  dl := med.makeList()
  dtemp.Execute(w, dl)
  err = archive(DrugAct{nm, false})
  if err != nil {
    fmt.Println(" db: ", err)
  }
}

func addD(w http.ResponseWriter, r *http.Request) {
  if r.Method == http.MethodGet {
    atemp.Execute(w, nil)
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      fmt.Println(" form: ", err)
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    name := r.FormValue("name")
    day := r.FormValue("day")
    evening := r.FormValue("evening")
    time := 0
    if day != "" {
      time += 1
    }
    if evening != "" {
      time += 2
    }
    med.AddDrug(Drug{name, time})
    fmt.Printf("  %s Added!\n", name)
    err = med.WriteData()
    if err != nil {
      fmt.Println(" write data: ", err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    dl := med.makeList()
    dtemp.Execute(w, dl)
    err = archive(DrugAct{name, true})
    if err != nil {
      fmt.Println(" db: ", err)
    }
  }
}

func delay(w http.ResponseWriter, r *http.Request) {
  cmd := exec.Command("sv", "up", "atd")
  err := cmd.Run()
  if err != nil {
    fmt.Println(" atd: ", err)
  }
  at := exec.Command("at", "-f", "/data/data/com.termux/files/home/cronsh/drugon.sh", "now", "+", "30", "minutes")
  err = at.Run()
  if err != nil {
    fmt.Println(" at: ", err)
  }
}

func showDb(w http.ResponseWriter, r *http.Request) {
  drs, err := dbRows()
  if err != nil {
    fmt.Println(" dbrows: ", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if len(drs) == 0 {
    fmt.Println(" EMPTY DB")
    return
  }
  stemp.Execute(w, drs)
}

func exit(w http.ResponseWriter, r *http.Request) {
  quit <-struct{}{}
}
