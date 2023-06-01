package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/pkg/errors"

	ctclient "github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
	"github.com/google/certificate-transparency-go/loglist3"
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

	for _, operator := range logList.Operators {
		log.Println("checking operator", operator.Name)
		for _, ctlog := range operator.Logs {
			log.Println("  checking log", ctlog.Description)
			if err := writeLog(ctx, ctlog, "data/json"); err != nil {
				log.Print("cannot write ", ctlog.Description, ": ", err)
			}
		}
	}
}

func writeLog(ctx context.Context, log *loglist3.Log, dstDir string) error {
	client, err := ctclient.New(log.URL, http.DefaultClient, jsonclient.Options{})
	if err != nil {
		return errors.Wrap(err, "cannot make new CT client")
	}

	const npages = 1
	name := slugify(log.Description)

	for i := 0; i < npages; i++ {
		start := 1024 * i
		end := 1024 * (i + 1)

		entries, err := client.GetEntries(ctx, int64(start), int64(end))
		if err != nil {
			return errors.Wrap(err, "cannot get entries 0-1024")
		}

		dstPath := filepath.Join(dstDir, fmt.Sprintf("%s_%d-%d.json", name, start, end))
		if err := writef(dstPath, func(w io.Writer) error {
			enc := json.NewEncoder(w)
			enc.SetIndent("", " ")
			return enc.Encode(entries)
		}); err != nil {
			return errors.Wrapf(err, "failed to write to %s", dstPath)
		}
	}

	return nil
}

func writef(dst string, writeFunc func(io.Writer) error) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := writeFunc(f); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func slugify(str string) string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return -1
		}
		return r
	}, str)
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
