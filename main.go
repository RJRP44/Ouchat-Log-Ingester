package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func closeDB(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func closeTCP(listener net.Listener) {
	err := listener.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	//Start Postgres connection
	dbConnection, err := sql.Open("postgres", "user="+os.Getenv("POSTGRES_USER")+" password="+os.Getenv("POSTGRES_PASSWORD")+" host=postgres dbname=ouchat_logs connect_timeout=5 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	//Close the connection when the program end
	defer closeDB(dbConnection)

	err = dbConnection.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to database")

	db := NewDatabase(dbConnection)

	//Start the http server for videos
	http.HandleFunc("/upload", uploadHandler(db))
	http.HandleFunc("/videos/", downloadHandler())
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//Start the tcp server
	listener, err := net.Listen("tcp", ":3010")
	if err != nil {
		log.Fatal(err)
	}

	defer closeTCP(listener)

	log.Printf("TCP server listening on :3010")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConnection(conn, db)
	}
}
