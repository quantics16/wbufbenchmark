package wbufbenchmark

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkPgConnWBufViaPostgres(b *testing.B) {
	user := os.Getenv("PQL_USER")
	password := os.Getenv("PQL_PASSWORD")
	database := os.Getenv("PQL_DATABASE")

	numberOfInserts := 1000
	jsonBlobSize := 16 * 1024

	sqlFmt := `insert into "sometable" as t (id, colA, colB, colC, colD) values %s 
                      on conflict(id) do update set colA=excluded.colA, colB=excluded.colB, colC=excluded.colC, colD=excluded.colD;`
	var args []interface{}
	var values []string
	valFmt := "(%s, %s, %s, %s, %s)"
	counter := 1
	randString := func() string {
		return strconv.FormatInt(rand.Int63(), 10)
	}
	for i := 0; i < numberOfInserts; i++ {
		values = append(values, fmt.Sprintf(valFmt,
			"$"+strconv.FormatInt(int64(counter+0), 10),
			"$"+strconv.FormatInt(int64(counter+1), 10),
			"$"+strconv.FormatInt(int64(counter+2), 10),
			"$"+strconv.FormatInt(int64(counter+3), 10),
			"$"+strconv.FormatInt(int64(counter+4), 10),
		))
		jsonblobFmt := `{ "foo": "%s" }`
		longstring := strings.Repeat("abcdefghijklmnop", jsonBlobSize/16)

		args = append(args, randString(), randString(), randString(), randString(), fmt.Sprintf(jsonblobFmt, longstring))

		counter += 5
	}
	sql := fmt.Sprintf(sqlFmt, strings.Join(values, ","))

	address := fmt.Sprintf("postgresql://%s", net.JoinHostPort("localhost", "5432"))

	config2, err := pgx.ParseConfig(address)
	if err != nil {
		b.FailNow()
	}

	config2.User = user
	config2.Password = password
	config2.Database = database

	conn, err := pgx.ConnectConfig(context.TODO(), config2)
	if err != nil {
		println("failed to create pgconn", err.Error())
		b.FailNow()
	}
	defer func() {
		b.StopTimer()
		conn.Close(context.TODO())
	}()

	_, err = conn.Exec(context.TODO(), `CREATE TABLE IF NOT EXISTS "sometable" (
			id VARCHAR,
			colA VARCHAR,
			colB VARCHAR,
			colC VARCHAR,
			colD JSONB,
			PRIMARY KEY (id)
		);`)
	if err != nil {
		println("failed to create table", err.Error())
		b.FailNow()
	}

	insertSize := len(sql)
	for i := 0; i < len(args); i++ {
		insertSize += len(args[i].(string))
	}
	fmt.Printf("message size in bytes: %d\n", insertSize)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = conn.Exec(context.TODO(), sql, args...)
		if err != nil {
			println("failed to run insert", err.Error())
			b.FailNow()
		}
	}
}
