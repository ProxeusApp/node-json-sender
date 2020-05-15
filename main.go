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

	"github.com/ProxeusApp/proxeus-core/externalnode"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	defaultServiceName = "JSON Data Sender"
	defaultServicePort = "8015"
	defaultJWTSecret   = "my secret 4"
	defaultProxeusUrl  = "http://127.0.0.1:1323"
	defaultAuthkey     = "auth"
)

func main() {
	proxeusUrl := os.Getenv("PROXEUS_INSTANCE_URL")
	if len(proxeusUrl) == 0 {
		proxeusUrl = defaultProxeusUrl
	}
	servicePort := os.Getenv("SERVICE_PORT")
	if len(servicePort) == 0 {
		servicePort = defaultServicePort
	}
	serviceUrl := os.Getenv("SERVICE_URL")
	if len(serviceUrl) == 0 {
		serviceUrl = "localhost:" + servicePort
	}
	jwtsecret := os.Getenv("SERVICE_SECRET")
	if len(jwtsecret) == 0 {
		jwtsecret = defaultJWTSecret
	}
	serviceName := os.Getenv("SERVICE_NAME")
	if len(serviceName) == 0 {
		serviceName = defaultServiceName
	}
	fmt.Println()
	fmt.Println("#######################################################")
	fmt.Println("# STARTING NODE - " + serviceName)
	fmt.Println("# listening on " + serviceUrl)
	fmt.Println("# connecting to " + proxeusUrl)
	fmt.Println("#######################################################")
	fmt.Println()

	mkrAddress := os.Getenv("PROXEUS_MKR_ADDRESS")
	if mkrAddress == "" {
		//panic("Environment variable xyz is required.")
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.GET("/health", externalnode.Health)
	{
		g := e.Group("/node/:id")
		conf := middleware.DefaultJWTConfig
		conf.SigningKey = []byte(jwtsecret)
		conf.TokenLookup = "query:" + defaultAuthkey
		g.Use(middleware.JWTWithConfig(conf))

		g.POST("/next", next)
		g.GET("/config", externalnode.Nop)
		g.POST("/config", externalnode.Nop)
		g.POST("/remove", externalnode.Nop)
		g.POST("/close", externalnode.Nop)
	}
	externalnode.Register(proxeusUrl, serviceName, serviceUrl, jwtsecret, "Sends form data to a REST endpoint via POST request")
	err := e.Start("0.0.0.0:" + servicePort)
	if err != nil {
		log.Println("[jsondatasender][run] Start err: ", err.Error())
	}
}

func next(c echo.Context) error {
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
	req, err := http.NewRequest("POST", os.Getenv("JSON_SENDER_URL"), bytes.NewReader(b))
	if err != nil {
		log.Printf("[jsondatasender][next] Error '%s'\n", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	addConfigHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b2, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 100*1024))
		log.Printf("[jsondatasender][next] SERVER NOT ACCEPTED '%s', RESPONSE '%s'\n", b, b2)
		return err
	}
	return c.NoContent(http.StatusOK)
}

func addConfigHeaders(req *http.Request) {
	req.Header.Set("clientid", os.Getenv("JSON_SENDER_CLIENT_ID"))
	req.Header.Set("tenantid", os.Getenv("JSON_SENDER_TENANT_ID"))
	req.Header.Set("secret", os.Getenv("JSON_SENDER_SECRET"))
	req.Header.Set("oauthserverurl", os.Getenv("JSON_SENDER_OAUTH_URL"))
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
