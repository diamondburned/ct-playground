package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	ctclient "github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
	"github.com/google/certificate-transparency-go/loglist3"
	"github.com/google/certificate-transparency-go/x509"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logList, err := fetchLogList(ctx)
	if err != nil {
		log.Fatalln("cannot fetch log list:", err)
	}

	logList = logList.SelectByStatus([]loglist3.LogStatus{
		loglist3.QualifiedLogStatus,
		loglist3.ReadOnlyLogStatus,
		loglist3.RetiredLogStatus,
		loglist3.UsableLogStatus,
	})

	x509.Certificate

	for _, operator := range logList.Operators {
		log.Println("checking operator", operator.Name)
		for _, ctlog := range operator.Logs {
			log.Println("  checking log", ctlog.Description)
			if err := printLog(ctx, ctlog); err != nil {
				log.Print("cannot print ", ctlog.Description, ": ", err)
			}
			return
		}
	}
}

func printLog(ctx context.Context, log *loglist3.Log) error {
	client, err := ctclient.New(log.URL, http.DefaultClient, jsonclient.Options{})
	if err != nil {
		return errors.Wrap(err, "cannot make new CT client")
	}

	entries, err := client.GetEntries(ctx, 0, 1024)
	if err != nil {
		return errors.Wrap(err, "cannot get entries 0-1024")
	}

	spew.Fdump(os.Stdout, entries)
	return nil
}

func fetchLogList(ctx context.Context) (loglist3.LogList, error) {
	var list loglist3.LogList

	req, err := http.NewRequestWithContext(ctx, "GET", loglist3.LogListURL, nil)
	if err != nil {
		return list, errors.Wrap(err, "cannot create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return list, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return list, errors.Wrap(err, "cannot decode JSON list")
	}

	return list, nil
}
