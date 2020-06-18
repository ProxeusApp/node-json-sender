package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	externalnode "github.com/ProxeusApp/node-go"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	defaultServiceName = "JSON Data Sender"
	defaultServicePort = "8015"
	defaultJWTSecret   = "my secret 4"
	defaultProxeusUrl  = "http://127.0.0.1:1323"
	defaultAuthkey     = "auth"
	serviceDetail      = "Sends form data to a REST endpoint via POST request"
)

type handler struct {
	proxeusURL  string
	serviceName string
	servicePort string
	serviceUrl  string
	jwtSecret   string
	targetURL   string
	headers     [][]string
}

func main() {
	proxeusURL := os.Getenv("PROXEUS_INSTANCE_URL")
	if len(proxeusURL) == 0 {
		proxeusURL = defaultProxeusUrl
	}
	servicePort := os.Getenv("SERVICE_PORT")
	if len(servicePort) == 0 {
		servicePort = defaultServicePort
	}
	serviceUrl := os.Getenv("SERVICE_URL")
	if len(serviceUrl) == 0 {
		serviceUrl = "localhost:" + servicePort
	}
	jwtSecret := os.Getenv("SERVICE_SECRET")
	if len(jwtSecret) == 0 {
		jwtSecret = defaultJWTSecret
	}
	serviceName := os.Getenv("SERVICE_NAME")
	if len(serviceName) == 0 {
		serviceName = defaultServiceName
	}

	targetURL := os.Getenv("JSON_SENDER_URL")
	if len(targetURL) == 0 {
		panic("JSON_SENDER_URL not defined")
	}

	h := &handler{
		proxeusURL:  proxeusURL,
		serviceName: serviceName,
		servicePort: servicePort,
		serviceUrl:  serviceUrl,
		jwtSecret:   jwtSecret,
		targetURL:   targetURL,
		headers:     extractHeaders(os.Environ()),
	}

	fmt.Println()
	fmt.Println("#######################################################")
	fmt.Println("# STARTING NODE - " + serviceName)
	fmt.Println("# listening on " + serviceUrl)
	fmt.Println("# connecting to " + proxeusURL)
	fmt.Println("#######################################################")
	fmt.Println()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.GET("/health", externalnode.Health)
	{
		g := e.Group("/node/:id")
		conf := middleware.DefaultJWTConfig
		conf.SigningKey = []byte(jwtSecret)
		conf.TokenLookup = "query:" + defaultAuthkey
		g.Use(middleware.JWTWithConfig(conf))

		g.POST("/next", h.next)
		g.GET("/config", externalnode.Nop)
		g.POST("/config", externalnode.Nop)
		g.POST("/remove", externalnode.Nop)
		g.POST("/close", externalnode.Nop)
	}
	err := h.register()
	if err != nil {
		panic("Could not register")
	}

	err = e.Start("0.0.0.0:" + servicePort)
	if err != nil {
		log.Println("[jsondatasender][run] Start err: ", err.Error())
	}
}

func (h *handler) register() error {
	return externalnode.Register(h.proxeusURL, h.serviceName, h.serviceUrl, h.jwtSecret, serviceDetail, 5)
}

func (h *handler) next(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)

	if err != nil {
		log.Printf("[jsondatasender][next] Error '%s'\n", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("[jsondatasender][next] Error '%s'\n", err)
		return err
	}

	b, err := json.Marshal(changeDataBeforeSend(data))
	if err != nil {
		log.Printf("[jsondatasender][next] Error '%s'\n", err)
		return err
	}
	req, err := http.NewRequest("POST", h.targetURL, bytes.NewReader(b))
	if err != nil {
		log.Printf("[jsondatasender][next] Error '%s'\n", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	addConfigHeaders(req, h.headers)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b2, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 100*1024))
		log.Printf("[jsondatasender][next] SERVER NOT ACCEPTED '%s', RESPONSE '%s'\n", b, b2)
		return err
	}

	return c.Stream(resp.StatusCode, resp.Header.Get("Content-Type"), resp.Body)
}

var envRegexp = regexp.MustCompile("JSON_SENDER_HEADER_(.+)=(.+)")

func extractHeaders(env []string) [][]string {
	var headers [][]string

	for _, env := range env {
		match := envRegexp.FindAllStringSubmatch(env, -1)
		if len(match) != 1 {
			continue
		}

		headers = append(headers, match[0][1:])
	}

	return headers
}

func addConfigHeaders(req *http.Request, headers [][]string) {
	for _, header := range headers {
		req.Header.Set(header[0], header[1])
	}
}

//data changes requested by customer
func changeDataBeforeSend(dat interface{}) interface{} {
	if m, ok := dat.(map[string]interface{}); ok {
		if d, ok := m["input"]; ok {
			bts, _ := json.Marshal(d)
			var dataCopy map[string]interface{}
			json.Unmarshal(bts, &dataCopy)
			if cs, ok := dataCopy["CapitalSource"]; ok {
				bts, _ := json.Marshal(cs)
				dataCopy["CapitalSource"] = string(bts)
			}
			return dataCopy
		}
	}
	return dat
}

func env(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}
