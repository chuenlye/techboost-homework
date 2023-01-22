package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const URL = "https://www.opendata.metro.tokyo.lg.jp/suisyoudataset/130001_event.csv"

type Record struct {
	Name    string
	Desc    string
	Org     string
	Address string
	Start   string
	End     string
}

func HtmlTemplate() string {
	return `
<!DOCTYPE html>
<html lang='en'>
<head>
    <meta charset='UTF-8'>
    <meta name='viewport' content='width=device-width, initial-scale=1.0'>
    <title>techboost golang course demo code </title>
</head>
<body>
    <h1>東京都イベント一覧</h1>
    <a href='/echo'>トップへ戻る</a>
    <ul>
        {{ range . }}
        <li>
            <h4>イベント名: {{ .Name }}</h4>
			<h4>説明: {{ .Desc }}</h4>
            <h4>主催者: {{ .Org }}</h4>
            <h4>開始日: {{ .Start }}</h4>
			<h4>終了日: {{ .End }}</h4>
			<br>
			<hr>
        </li>
        {{ end }}
    </ul>
</body>
</html>
	`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	defer resp.Body.Close()

	//body, _ := io.ReadAll(resp.Body)
	r := csv.NewReader(transform.NewReader(resp.Body, japanese.ShiftJIS.NewDecoder()))
	records, err := r.ReadAll()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	recordsStr, recordObjects := func(data [][]string) (string, []Record) {
		var recordsSlice []string
		var recordObjects []Record
		for idx, record := range data {
			recordsSlice = append(recordsSlice, strings.Join(record, ","))
			if idx > 0 {
				recordObjects = append(recordObjects, Record{
					Name:    record[4],
					Desc:    record[12],
					Org:     record[19],
					Address: record[21],
					Start:   record[7],
					End:     record[9],
				})
			}
		}
		return strings.Join(recordsSlice, ","), recordObjects
	}(records)

	// Debug: output raw csv data to console
	fmt.Println(recordsStr)
	fmt.Printf("records number: %d\n", len(recordObjects))
	fmt.Printf("last record: %v\n", recordObjects[len(recordObjects)-1])

	tpl, err := template.New("").Parse(HtmlTemplate())
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	b := bytes.NewBufferString("")
	if err := tpl.Execute(b, recordObjects); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
		Body:       b.String(),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
