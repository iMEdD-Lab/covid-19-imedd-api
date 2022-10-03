package file

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func ReadCsv(url string, sourceIsFile bool) ([][]string, error) {
	var err error
	filepath := url
	if !sourceIsFile {
		filepath, err = DownloadFile(url)
		if err != nil {
			return nil, err
		}
	}
	return ReadCsvFile(filepath)
}

func ReadCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Unable to parse file as CSV for "+filePath, err)
	}

	return records, nil
}

func DownloadFile(url string) (string, error) {
	filepath := strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("cannot create file: %s", err)
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot download file %s: %s", url, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("io.Copy error: %s", err)
	}

	return filepath, nil
}
