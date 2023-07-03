package main

import (
  "fmt"
  "time"
  "database/sql"
  
  _ "github.com/mattn/go-sqlite3"
)

type DrugAct struct {
  Name string
  Added bool
}

type DrugRow struct {
  Name string
  Start string
  Finish string
}

func dbRows() ([]DrugRow, error) {
  db, err := sql.Open("sqlite3", "data/drugs.db")
  if err != nil {
    return []DrugRow{}, err
  }
  defer db.Close()
  query := "select * from courses"
  rows, err := db.Query(query)
  if err != nil {
    return []DrugRow{}, err
  }
  defer rows.Close()
  res := []DrugRow{}
  for rows.Next() {
    var dr DrugRow
    err = rows.Scan(&dr.Name, &dr.Start, &dr.Finish)
    if err != nil {
      return []DrugRow{}, fmt.Errorf(" db scan: %w", err)
    }
    res = append(res, dr)
  }
  return res, nil
}

func archive(da DrugAct) error {
  db, err := sql.Open("sqlite3", "data/drugs.db")
  if err != nil {
    return err
  }
  defer db.Close()
  statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS courses(name varchar, start varchar, finish varchar)")
  statement.Exec()
  dnow := time.Now().Format("2006-01-02")
  var query string
  query = "select * from courses where name = ?"
  row := db.QueryRow(query, da.Name)
  var dr DrugRow
  err = row.Scan(&dr.Name, &dr.Start, &dr.Finish)
  if err == sql.ErrNoRows {
    ins := "insert into courses (name, start, finish) values(?, ?, ?)"
    if da.Added {
      _, err := db.Exec(ins, da.Name, dnow, "---")
      if err != nil {
        return err
      }
    } else {
      _, err := db.Exec(ins, da.Name, "---", dnow)
      if err != nil {
        return err
      }
    }
  } else {
    var upd string
    if da.Added {
      upd = "update courses set start = ? where name = ?"
    } else {
      upd = "update courses set finish = ? where name = ?"
    }
    _, err := db.Exec(upd, dnow, da.Name)
    if err != nil {
      return err
    }
  }
  return nil
}
