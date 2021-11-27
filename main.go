package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Get env var or default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// TODO env filedan istifade et
// FIXME just test

// Get the port to listen on
func getListenAddress() string {
	// port := getEnv("PORT", "8089")
	port := viper.GetString("port")
	return ":" + port
}

// Log the env variables required for a reverse proxy
func logSetup() {
	// a_condtion_url := os.Getenv("A_CONDITION_URL")
	a_condtion_url := viper.GetString("condition.auth")
	// b_condtion_url := os.Getenv("B_CONDITION_URL")
	b_condtion_url := viper.GetString("condition.client")
	// c_condtion_url := os.Getenv("C_CONDITION_URL")
	c_condtion_url := viper.GetString("condition.courier")
	// d_condtion_url := os.Getenv("D_CONDITION_URL")
	d_condtion_url := viper.GetString("condition.order_cook")
	// default_condtion_url := os.Getenv("DEFAULT_CONDITION_URL")
	default_condtion_url := viper.GetString("condition.default")

	log.Printf("Server will run on: %s\n", getListenAddress())
	log.Printf("Redirecting to A url: %s\n", a_condtion_url)
	log.Printf("Redirecting to B url: %s\n", b_condtion_url)
	log.Printf("Redirecting to C url: %s\n", c_condtion_url)
	log.Printf("Redirecting to D url: %s\n", d_condtion_url)
	log.Printf("Redirecting to Default url: %s\n", default_condtion_url)
}

type requestPayloadStruct struct {
	ProxyCondition string `json:"service_name"`
}

// Get a json decoder for a given requests body
func requestBodyDecoder(request *http.Request) *json.Decoder {
	// Read body to buffer
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		panic(err)
	}

	// Because go lang is a pain in the ass if you read the body then any susequent calls
	// are unable to read the body again....
	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
}

// Parse the requests body
func parseRequestBody(request *http.Request) requestPayloadStruct {
	decoder := requestBodyDecoder(request)

	var requestPayload requestPayloadStruct
	err := decoder.Decode(&requestPayload)

	if err != nil {
		panic(err)
	}

	return requestPayload
}

// Log the typeform payload and redirect url
func logRequestPayload(requestionPayload requestPayloadStruct, proxyUrl string) {
	log.Printf("proxy_condition: %s, proxy_url: %s\n", requestionPayload.ProxyCondition, proxyUrl)
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	log.Println("url.Host:", url.Host)
	req.URL.Scheme = url.Scheme
	log.Println("url.Scheme:", url.Scheme)
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

// Get the url for a given proxy condition
func getProxyUrl(proxyConditionRaw string) string {
	proxyCondition := strings.ToUpper(proxyConditionRaw)

	a_condtion_url := viper.GetString("condition.auth")
	b_condtion_url := viper.GetString("condition.client")
	c_condtion_url := viper.GetString("condition.courier")
	d_condtion_url := viper.GetString("condition.order_cook")
	default_condtion_url := viper.GetString("condition.default")

	if proxyCondition == "AUTH" {
		log.Println("Entered auth condition")
		return a_condtion_url
	}

	if proxyCondition == "CLIENT" {
		log.Println("Entered auth condition")
		return b_condtion_url
	}

	if proxyCondition == "COURIER" {
		log.Println("Entered courier condition")
		return c_condtion_url
	}

	if proxyCondition == "ORDER-COOK" {
		log.Println("Entered order-cook condition")
		return d_condtion_url
	}
	log.Println("condition-lara girmeden geldi bura!")
	return default_condtion_url
}

func getUrlByPath(path []string) string {

	a_condtion_url := viper.GetString("condition.auth")
	b_condtion_url := viper.GetString("condition.client")
	c_condtion_url := viper.GetString("condition.courier")
	d_condtion_url := viper.GetString("condition.order_cook")
	default_condtion_url := viper.GetString("condition.default")
	basePath := path[1]
	if strings.ToLower(basePath) == "register" || strings.ToLower(basePath) == "login" {
		return a_condtion_url
	}

	if strings.ToLower(basePath) == "clients" {
		return b_condtion_url
	}

	if strings.ToLower(basePath) == "couriers" {
		return c_condtion_url
	}

	if strings.ToLower(basePath) == "orders" || strings.ToLower(basePath) == "order-create" || strings.ToLower(basePath) == "meal-create" || strings.ToLower(basePath) == "mealoptions" || strings.ToLower(basePath) == "categories" || strings.ToLower(basePath) == "ingredients" || strings.ToLower(basePath) == "meals" || strings.ToLower(basePath) == "meal" || strings.ToLower(basePath) == "ingredient-create" || strings.ToLower(basePath) == "mealoption-create" || strings.ToLower(basePath) == "category-create" || strings.ToLower(basePath) == "deliveryareas" || strings.ToLower(basePath) == "cook-create" || strings.ToLower(basePath) == "recommend-create" || strings.ToLower(basePath) == "resume-create" || strings.ToLower(basePath) == "cooks"{
		return d_condtion_url
	}
	return default_condtion_url
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {

	requestPayload := parseRequestBody(req)

	path := req.URL.Path
	fmt.Println("PATH:", path)
	parts := strings.Split(path, "/")

	fmt.Println("*************")
	fmt.Println("parts:", parts)
	fmt.Println("parts[0]:", parts[0])
	fmt.Println("parts[1]:", parts[1])
	fmt.Println("parts[1]-len:", len(parts[1]))

	fmt.Println("in handle - path:", path)
	// url := getProxyUrl(requestPayload.ProxyCondition)
	url := getUrlByPath(parts)
	fmt.Println("in handle - url:", url)
	logRequestPayload(requestPayload, url)
	fmt.Println("in handle - requestPayload:", requestPayload)
	serveReverseProxy(url, res, req)
}

func main() {

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	// Log setup values
	logSetup()

	// start server
	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
