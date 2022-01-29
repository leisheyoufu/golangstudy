package main

import (
	"encoding/json"
	"fmt"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
	"log"
	"os"
)

func main() {
	var err error
	md := []string{
		"name=Name, type=UTF8, encoding=PLAIN_DICTIONARY",
		"name=Age, type=INT32",
		"name=Id, type=INT64",
		"name=Weight, type=FLOAT",
		"name=Sex, type=BOOLEAN",
	}

	//write
	fw, err := local.NewLocalFileWriter("output/csv.parquet")
	if err != nil {
		log.Println("Can't open file", err)
		return
	}
	pw, err := writer.NewCSVWriter(md, fw, 4)
	if err != nil {
		log.Println("Can't create csv writer", err)
		return
	}

	num := 10
	for i := 0; i < num; i++ {
		data := []string{
			fmt.Sprintf("%s_%d", "Student Name", i),
			fmt.Sprintf("%d", 20+i%5),
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%f", 50.0+float32(i)*0.1),
			fmt.Sprintf("%t", i%2 == 0),
		}
		rec := make([]*string, len(data))
		for j := 0; j < len(data); j++ {
			rec[j] = &data[j]
		}
		if err = pw.WriteString(rec); err != nil {
			log.Println("WriteString error", err)
		}

		data2 := []interface{}{
			"Student Name",
			int32(20 + i%5),
			int64(i),
			float32(50.0 + float32(i)*0.1),
			i%2 == 0,
		}
		if err = pw.Write(data2); err != nil {
			log.Println("Write error", err)
		}

	}
	if err = pw.WriteStop(); err != nil {
		log.Println("WriteStop error", err)
	}
	log.Println("Write Finished")
	fw.Close()

	fr, err := local.NewLocalFileReader("output/csv.parquet")
	if err != nil {
		log.Println("Can't open file")
		return
	}

	pr, err := reader.NewParquetReader(fr, nil, 4)
	if err != nil {
		log.Println("Can't create parquet reader", err)
		return
	}
	num = int(pr.GetNumRows())
	res, err := pr.ReadByNumber(num)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't cat: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(res[0].([]byte)))
	jsonBs, err := json.Marshal(res[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't to json: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBs))
	pr.ReadStop()
	fr.Close()

}
