
package httpfetch

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/imkira/go-ttlmap"
)


var localCache = ttlmap.New(nil)
var templateCache = make(map[string]*template.Template)

func query(fetchArgs HttpFetch, dnsName string) (string, error) {
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
		appendHeaders(&req.Header, fetchArgs.ReqHeaders)
		log.Debugf("About to send HTTP request: %v", req)

		for i := 1; i <= 10; i++ {
			log.Debugf("Sending requesting: trying %d times", i)
			resp, err = client.Do(req)

			if err != nil {
				log.Warningf("HTTP Error %v", err)
			}else if isSuccessful(resp.StatusCode) {
				break
			}

			time.Sleep(1 * time.Second)
		}

		if err != nil{
			return "", err
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
		log.Debugf("Response from server: %v", resBody)

		ipAddress := resBody
		if len(fetchArgs.ResIPExtractor) > 0 {
			ipAddress, err = readString("ip-extractor", fetchArgs.ResIPExtractor, resBody)
			if err != nil {
				log.Warningf("Could not read IP address from response: %v", err)
				return "", err
			}
		}
		if len(ipAddress) <= 0 {
			log.Info("Record not found in", resBody)
			return "", nil
		}

		ttlSeconds := 60
		if len(fetchArgs.ResTTLExtractor) > 0 {
			ttl, err := readString("ttl-extractor", fetchArgs.ResTTLExtractor, resBody)
			if err != nil {
				log.Warningf("Could not read TTL from response, falling back to 60: %v", err)
			} else {
				ttlSeconds, err = strconv.Atoi(ttl)
				if err != nil {
					log.Warningf("Could not read TTL %s as number, falling back to 60", ttl)
				}
			}
		}

		localCache.Set(dnsName, ttlmap.NewItem(ipAddress, ttlmap.WithTTL(time.Duration(ttlSeconds)*time.Second)), nil)
		return ipAddress, nil
	}
}

func isSuccessful(statusCode int) bool {
	return statusCode >= 200 && statusCode < 400
}


func buildUrl(baseUrl string, queryTemplate string, dnsName string) string{
	if len(queryTemplate) <= 0{
		return baseUrl
	}

	var urlBuilder strings.Builder
	urlBuilder.WriteString(baseUrl)
	if strings.LastIndex(baseUrl, "?") < 0 {
		urlBuilder.WriteString("?")
	}else{
		urlBuilder.WriteString("&")
	}

	queryString := fmt.Sprintf(queryTemplate, dnsName)
	urlBuilder.WriteString(queryString)
	return urlBuilder.String()
}

func buildBody(bodyTemplate string, dnsName string) io.Reader {
	if len(bodyTemplate) <= 0{
		return nil
	}

	return strings.NewReader(fmt.Sprint(bodyTemplate, dnsName))
}

func appendHeaders(header *http.Header, reqHeaderArgs []string) {
	if len(reqHeaderArgs) <= 0{
		return
	}

	for _,h := range reqHeaderArgs {
		indexOfColon := strings.Index(h, ":")
		if indexOfColon < 0 {
			continue
		}

		headerName := strings.Trim(h[0:indexOfColon], " ")
		headerValue := strings.Trim(h[indexOfColon+1:], " ")
		header.Set(headerName, headerValue)
	}
}

func readString(templateName string, tmplStr string, text string) (string, error) {
	if len(tmplStr) <= 0 {
		return text, nil
	}

	compiledTemplate := templateCache[templateName]
	if compiledTemplate == nil {
		var err error
		compiledTemplate, err = template.New("extractor").Funcs(createFuncMap()).Parse(tmplStr)
		if err != nil {
			return "", err
		}
		templateCache[templateName] = compiledTemplate
	}

	vars := make(map[string]interface{})
	vars["ResponseText"] = text

	var resultBytes bytes.Buffer
	err := compiledTemplate.Execute(&resultBytes, vars)
	if err != nil {
		return "", err
	}
	return resultBytes.String(), nil
}

func createFuncMap() template.FuncMap {
	txtFuncMap := sprig.TxtFuncMap()
	delete(txtFuncMap, "env")
	delete(txtFuncMap, "expandenv")
	delete(txtFuncMap, "getHostByName")

	extra := template.FuncMap{
		"toXml":        toXml,
		"fromXml":      fromXml,
	}

	for k, v := range extra {
		txtFuncMap[k] = v
	}

	return txtFuncMap
}

// todo: supports XML using https://github.com/beevik/etree
func fromXml(xmlStr string) interface{} {
	var output interface{}
	err := xml.Unmarshal([]byte(xmlStr), &output)
	if err != nil {
		return nil
	}
	return output
}

func toXml(v interface{}) string {
	data, err := xml.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}