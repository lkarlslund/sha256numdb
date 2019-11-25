package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"

	"github.com/cheggaaa/pb"
	sha256 "github.com/minio/sha256-simd"
)

// Does fast lookups of SHA256 -> phone number lookups

const dbname = "phone.db"
const prefix = "045"
const from = 0
const to = 99999999 // excessive, but complete
const indexcap = 3

func stringifynumber(num int) string {
	return fmt.Sprintf("045%08v", num)
}

type number int32

func main() {
	db, err := os.Open(dbname)
	if err != nil {
		fmt.Println("Database problem - generating a new one ...")

		f, ferr := os.Create(dbname)
		if ferr != nil {
			fmt.Println("Problem creating database file, aborting")
			os.Exit(1)
		}
		bdb := bufio.NewWriter(f)

		fmt.Print("Calculating hashes ")

		shas := make(map[uint16][]uint32) // maps the first four SHA256 bytes to a phone number
		bar := pb.StartNew(to - from)
		for num := from; num <= to; num++ {
			strnum := stringifynumber(num)
			hash := sha256.Sum256([]byte(strnum))

			hval := uint16(hash[0])<<8 | uint16(hash[1])
			arr := shas[hval]
			arr = append(arr, uint32(num))
			shas[hval] = arr

			bar.Increment()
		}
		bar.Finish()

		fmt.Println("Writing index")
		for num := 0; num <= 65535; num++ {
			binary.Write(bdb, binary.LittleEndian, len(shas[uint16(num)]))
		}
		fmt.Println("Writing numbers")
		for num := 0; num <= 65535; num++ {
			for _, phonenumber := range shas[uint16(num)] {
				binary.Write(bdb, binary.LittleEndian, uint32(phonenumber))
			}
		}
		fmt.Println("Done creating database")

		bdb.Flush()
		f.Close()

		db, err = os.Open(dbname)
	}
	if err != nil {
		fmt.Println("Giving up, sorry")
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hash, ok := r.URL.Query()["lookup"]
		if ok {

			db.Seek(0, 0)
		}
		w.Write([]byte("NOT FOUND"))
	})
	http.ListenAndServe(":8080", nil)
}
