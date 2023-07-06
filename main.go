package main

import (
  "flag"
  "fmt"
  "os"
  "io/ioutil"
  "time"
  "encoding/json"
  "net/http"
  "os/exec"
  "html/template"
  "embed"
)

type Medicine struct {
  Box []Drug `json:"box"`
}

type DrugList struct {
  Time string
  List []string
  Port string
}

type Drug struct {
  Name string `json:"name"`
  DayTime int `json:"day_time"`
}

var (
  Port string
  med Medicine
  //go:embed static
  staticDir embed.FS
  //go:embed templates
  tempDir embed.FS
  jsn = "data/med.json"
  dtemp *template.Template
  atemp *template.Template
  stemp *template.Template
  quit = make(chan struct{}, 1)
)

func (d *DrugList) SetTime() {
  h := time.Now().Hour()
  if h > 18 {
    d.Time = "Вечер"
  } else {
    d.Time = "День"
  }
}

func (m *Medicine) makeList() DrugList {
  dl := DrugList{}
  dl.SetTime()
  lst := []string{}
  for _, d := range m.Box {
    switch dl.Time {
      case "Вечер":
        if (d.DayTime == 2) || (d.DayTime == 3) {
          lst = append(lst, d.Name)
        }
      case "День":
        if d.DayTime == 1 || d.DayTime == 3 {
          lst = append(lst, d.Name)
        }
    }
  }
  dl.List = lst
  dl.Port = Port
  return dl
}

func (m *Medicine) AddDrug(d Drug) {
  nb := append(m.Box, d)
  m.Box = nb
}

func display(w http.ResponseWriter, r *http.Request) {
  if _, err := os.Stat(jsn); err != nil {
    if os.IsNotExist(err) {
      init := Medicine{}
      inj, _ := json.Marshal(init)
      os.WriteFile(jsn, inj, 0640)
    }
  }
  f, err := os.Open(jsn)
  if err != nil {
    fmt.Println(" file: ", err)
    return
  }
  defer f.Close()
  data, err := ioutil.ReadAll(f)
  if err != nil {
    fmt.Println(" read: ", err)
  }
  err = json.Unmarshal(data, &med)
  if err != nil {
    fmt.Println(" unmarshal: ", err)
    return
  }
  drl := med.makeList()
  err = dtemp.Execute(w, drl)
  if err != nil {
    fmt.Println(" execute: ", err)
  }
}

func deleteD(w http.ResponseWriter, r *http.Request) {
  nm := r.URL.Query().Get("name")
  fmt.Printf(" %s will be deleted\n", nm)
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

func (m Medicine) WriteData() error {
  data, err := json.MarshalIndent(med, " ", "  ")
  if err != nil {
    return err
  }
  os.WriteFile(jsn, data, 0640)
  return nil
}

func delay(w http.ResponseWriter, r *http.Request) {
  cmd := exec.Command("sv-enable", "atd")
  err := cmd.Run()
  if err != nil {
    fmt.Println(" atd: ", err)
  }
  at := exec.Command("at", "-f", "/data/data/com.termux/files/home/cronsh/drugon.sh", "now", "+", "20", "minutes")
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

func main() {
  var portNum int
  flag.IntVar(&portNum, "p", 8754, "server port number")
  flag.Parse()
  if portNum < 2500 {
    fmt.Println(" port should be greater than 2500")
    return
  }
  Port = fmt.Sprintf(":%d", portNum)
  mux := http.NewServeMux()
  mux.HandleFunc("/", display)
  mux.HandleFunc("/showdb", showDb)
  mux.HandleFunc("/delete", deleteD)
  mux.HandleFunc("/add", addD)
  mux.HandleFunc("/delay", delay)
  mux.HandleFunc("/exit", exit)
  fsr := http.FileServer(http.FS(staticDir))
  mux.Handle("/static/", fsr)
  dtemp, _ = template.ParseFS(tempDir, "templates/display.html")
  atemp, _ = template.ParseFS(tempDir, "templates/add.html")
  stemp, _ = template.ParseFS(tempDir, "templates/showdb.html")
  server := http.Server{Addr: Port, Handler: mux}
  go func() {
    err := server.ListenAndServe()
    if err != nil {
      fmt.Println(" server: ", err)
    }
  }()
  cmd := exec.Command("termux-open-url", "http://localhost" + Port)
  cmd.Run()
  vol := exec.Command("termux-volume", "music", "10")
  vol.Run()
  snd := exec.Command("play-audio", "data/shaker.wav")
  err := snd.Run()
  if err != nil {
    fmt.Println(" sound: ", err)
  }
  <-quit
}