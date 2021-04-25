package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    xerrors "github.com/pkg/errors"
    "log"
)

func openDB() *sql.DB {
    db, err := sql.Open("mysql", "root:lsx19890329@tcp(127.0.0.1:3306)/qmw")
    if err != nil {
        log.Fatalf("Open db 127.0.0.1:3306 error: %v\n", err)
    }
    err = db.Ping()
    if err != nil {
        log.Fatalf("Ping db 127.0.0.1:3306 error: %v\n", err)
    }
    return db
}

func queryCh(db *sql.DB) (string, error) {
    var ch string
    err := db.QueryRow("select ch from t_character where id = ?", 0).Scan(&ch)
    if err != nil {
        // 虽然 sql.ErrNoRows 不是真正意义上的错误，但是dao层并不知道空结果到底该如何处理
        // 把空结果当做Error处理是为了强行让程序员处理结果为空的情况
        return ch, xerrors.Wrap(err, "queryCh error")
    }
    return ch, nil
}

func main() {
    db := openDB()
    defer db.Close()

    ch, err := queryCh(db)
    if err != nil {
        // 即便是 sql.ErrNoRows，也该交由Service层逻辑处理，而不该在dao层直接处理掉
        // 因为我们并不知道结果为空的时候，Service层逻辑会做些什么
        if xerrors.Cause(err) == sql.ErrNoRows {
            log.Printf("queryCh ch1: %v\n", ch)
            return
        }

        log.Printf("queryCh error: %v\n", err)
        return
    }

    log.Printf("queryCh ch2: %v\n", ch)
}
