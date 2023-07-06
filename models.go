package main


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

type DrugAct struct {
  Name string
  Added bool
}

type DrugRow struct {
  Name string
  Start string
  Finish string
}
