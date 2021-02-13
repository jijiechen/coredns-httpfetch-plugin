
package httpfetch

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/imkira/go-ttlmap"
)


var localCache = ttlmap.New(nil)

func query(fetchArgs *Httpfetch, dnsName string) (string, error) {
	item, err := localCache.Get(dnsName)
	if err == nil {
		log.Debug(fmt.Sprintf("Found in local cache %s", dnsName))
		return item.Value().(string), nil
	} else {
		client := &http.Client{}
		var resp *http.Response

		req, err := http.NewRequest(fetchArgs.ReqMethod,
			buildUrl(fetchArgs.ReqUrl, fetchArgs.ReqQueryTemplate, dnsName),
			buildBody(fetchArgs.ReqBodyTemplate, dnsName))
		appendHeaders(req.Header, fetchArgs.ReqHeaders)
		log.Debug("About to send HTTP request: %v", req)

		for i := 1; i <= 10; i++ {
			log.Debug("Sending requesting: trying %d times", i)
			resp, err = client.Do(req)

			if err != nil {
				log.Warning("HTTP Error %v", err)
			}

			if isSuccessful(resp.StatusCode) {
				break
			}

			time.Sleep(1 * time.Second)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if !isSuccessful(resp.StatusCode) {
			err := fmt.Errorf("HTTP Error: response code did not indicate success: %d, body: %v", resp.StatusCode, body)
			return "", err
		}

		resBody := string(body)
		log.Debug("Response from server: %v", resBody)

		ipAddress, err := readString(fetchArgs.ResIPExtractor, resBody)
		if err != nil {
			log.Warning("Could not read IP address from response: %v", err)
			return "", err
		}
		if len(ipAddress) <= 0 {
			log.Info("Record not found in", resBody)
			return "", nil
		}

		ttl, err := readString(fetchArgs.ResTTLExtractor, resBody)
		if err != nil {
			log.Warning("Could not read TTL from response: %v", err)
			return "", err
		}
		ttlSeconds, err := strconv.Atoi(ttl)
		if err != nil {
			ttlSeconds = 60
			log.Warning("Could not read TTL %s as number, falling back to 60", ttl)
			return "", err
		}

		localCache.Set(dnsName, ttlmap.NewItem(ipAddress, ttlmap.WithTTL(time.Duration(ttlSeconds)*time.Second)), nil)
		return ipAddress, nil
	}
}

func isSuccessful(statusCode int) bool {
	return statusCode >= 200 && statusCode < 400
}


func buildUrl(baseUrl string, queryStringTemplate string, dnsName string) string{
	if len(queryStringTemplate) <= 0{
		return baseUrl
	}

	var urlBuilder strings.Builder
	urlBuilder.WriteString(baseUrl)
	if strings.LastIndex(baseUrl, "?") < 0 {
		urlBuilder.WriteString("?")
	}else{
		urlBuilder.WriteString("&")
	}

	queryString := fmt.Sprintf(queryStringTemplate, dnsName)
	urlBuilder.WriteString(queryString)
	return urlBuilder.String()
}

func buildBody(bodyTemplate string, dnsName string) io.Reader {
	if len(bodyTemplate) <= 0{
		return nil
	}

	return strings.NewReader(fmt.Sprint(bodyTemplate, dnsName))
}

func substr(s string, start, end int) string {
	if end < 0{
		end = len(s) - start
	}

	counter, startIdx := 0, 0
	for i := range s {
		if counter == start {
			startIdx = i
		}
		if counter == end {
			return s[startIdx:i]
		}
		counter++
	}
	return s[startIdx:]
}

func appendHeaders(header http.Header, reqHeaderArgs []string) {
	if len(reqHeaderArgs) <= 0{
		return
	}

	for _,h := range reqHeaderArgs {
		indexOfColon := strings.Index(h, ":")
		if indexOfColon < 0 {
			continue
		}

		headerName := strings.Trim(substr(h, 0, indexOfColon), " ")
		headerValue := strings.Trim(substr(h, indexOfColon + 1, -1 /*end of string*/), " ")
		header.Set(headerName, headerValue)
	}
}

func readJson(jsonStr string, fieldSelector string) string{

}

func readString(tmplStr string, text string) (string, error) {
	if len(tmplStr) <= 0 {
		return text, nil
	}

	tmpl, err := template.New("extractor").Parse(tmplStr)
	if err != nil {
		return text, err
	}

	var resultBytes bytes.Buffer

	err = tmpl.Execute(&resultBytes, /* responseText, json, xml,  */)
	if err != nil {
		return text, err
	}
	return resultBytes.String(), nil
}