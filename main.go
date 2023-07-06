package main

import (
  "flag"
  "fmt"
  "os"
  "embed"
  "io/ioutil"
  "time"
  "encoding/json"
  "net/http"
  "os/exec"
  "html/template"

)


var (
  Port string
  jsn = "data/drugs.json"
  //go:embed templates
  tempDir embed.FS
  //go:embed static
  staticDir embed.FS
  med Medicine
  dtemp *template.Template
  atemp *template.Template
  stemp *template.Template
  quit = make(chan struct{}, 1)
)

func PrepareData() error {
  if _, err := os.Stat(jsn); err != nil {
    if os.IsNotExist(err) {
      init := Medicine{}
      inj, _ := json.Marshal(init)
      os.WriteFile(jsn, inj, 0640)
    }
  }
  f, err := os.Open(jsn)
  if err != nil {
    return fmt.Errorf("open json: %w", err)
  }
  defer f.Close()
  data, err := ioutil.ReadAll(f)
  if err != nil {
    return fmt.Errorf("read json: %w", err)
  }
  err = json.Unmarshal(data, &med)
  if err != nil {
    return fmt.Errorf("unmarshal json: %w", err)
  }
  return nil
}

func (d *DrugList) SetTime() {
  h := time.Now().Hour()
  if h > 18 {
    d.Time = "Вечер"
  } else {
    d.Time = "День"
  }
}

func (m *Medicine) makeList() DrugList {
  dl := DrugList{Port: Port}
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
  return dl
}

func (m *Medicine) AddDrug(d Drug) {
  nb := append(m.Box, d)
  m.Box = nb
}


func (m Medicine) WriteData() error {
  data, err := json.MarshalIndent(med, " ", "  ")
  if err != nil {
    return err
  }
  os.WriteFile(jsn, data, 0640)
  return nil
}


func main() {
  var p int
  flag.IntVar(&p, "p", 8754, "server port number")
  flag.Parse()
  if p < 2500 {
    fmt.Println("Port should be Number greater than 2500")
    time.Sleep(5 * time.Second)
    return
  }
  Port = fmt.Sprintf(":%d", p)

  mux := http.NewServeMux()
  mux.HandleFunc("/", display)
  mux.HandleFunc("/showdb", showDb)
  mux.HandleFunc("/delete", deleteD)
  mux.HandleFunc("/add", addD)
  mux.HandleFunc("/delay", delay)
  mux.HandleFunc("/exit", exit)
  fsrv := http.FileServer(http.FS(staticDir))
  mux.Handle("/static/", fsrv)
  dtemp, _ = template.ParseFS(tempDir, "templates/display.html")
  atemp, _ = template.ParseFS(tempDir, "templates/add.html")
  stemp, _ = template.ParseFS(tempDir, "templates/showdb.html")
  server := http.Server{Addr: Port, Handler: mux}
  go server.ListenAndServe()
  cmd := exec.Command("xdg-open", "http://localhost:8754")
  cmd.Run()
  vol := exec.Command("amixer", "-D", "pulse", "sset", "Master", "75%")
  vol.Run()
  snd := exec.Command("ffplay", "-v", "0", "-nodisp", "-autoexit", "data/shaker.wav")
  snd.Run()
  
  <-quit
  
}