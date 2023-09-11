package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const certFile = "/etc/cert/localhost.crt"
const keyFile = "/etc/cert/localhost.key"
const customerMapFile = "/etc/customerMap.json"

const infoHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
		<style>
			label.login {
				padding: 5px;
				font-weight: bold;
				font-size: large;
			}
	
			input.login {
				padding: 5px;
				font-weight: normal;
				font-size: large;
			}
			.submitButton {
				background-color:#44c767;
				border-radius:23px;
				border:1px solid #18ab29;
				display:inline-block;
				cursor:pointer;
				color:#ffffff;
				font-family: Arial, serif;
				font-size:19px;
				padding:9px 45px;
				text-decoration:none;
				text-shadow:1px 2px 0 #2f6627;
			}
			.submitButton:hover {
				background-color:#5cbf2a;
			}
			.submitButton:active {
				position:relative;
				top:1px;
			}
	
		</style>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="images/logo.png" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<h2>Please use your customer specific URL to connect to the ElektrikGreen cloud service</h2>
		</div>
	</body>
</html>`

const updatedHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
		<style>
			label.login {
				padding: 5px;
				font-weight: bold;
				font-size: large;
			}
	
			input.login {
				padding: 5px;
				font-weight: normal;
				font-size: large;
			}
			.submitButton {
				background-color:#44c767;
				border-radius:23px;
				border:1px solid #18ab29;
				display:inline-block;
				cursor:pointer;
				color:#ffffff;
				font-family: Arial, serif;
				font-size:19px;
				padding:9px 45px;
				text-decoration:none;
				text-shadow:1px 2px 0 #2f6627;
			}
			.submitButton:hover {
				background-color:#5cbf2a;
			}
			.submitButton:active {
				position:relative;
				top:1px;
			}
	
		</style>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="images/logo.png" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<ol>%s
			</ol>
		</div>
	</body>
</html>`

var customerMap map[string]string

func AddNewOrUpdateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		if _, err := fmt.Fprint(w, infoHTML); err != nil {
			log.Fatal(err)
		}
	} else {
		if strings.HasPrefix(r.RequestURI, "/?") {
			qry := r.RequestURI[2:]
			for _, entry := range strings.Split(qry, "&") {
				customer := strings.Split(entry, "=")
				if len(customer) == 2 {
					customerMap[strings.ToLower(customer[0])] = customer[1]
				}
			}
			if fileContent, err := json.MarshalIndent(customerMap, "", "    "); err != nil {
				log.Fatal(err)
			} else {
				if err := ioutil.WriteFile(customerMapFile, fileContent, 0644); err != nil {
					log.Fatal(err)
				}
			}
			list := ""
			for id, url := range customerMap {
				list = list + "<li>" + id + " ==> " + url + "</li>"
			}
			if _, err := fmt.Fprintf(w, updatedHTML, list); err != nil {
				log.Fatal(err)
			}
		} else {
			if _, err := fmt.Fprint(w, infoHTML); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {
	customerMap = make(map[string]string, 0)

	if customerMapContent, err := os.ReadFile(customerMapFile); err != nil {
		customerMap["cedar"] = "https://cloud.cedartechnology.com:41000"
		customerMap["eg1"] = "https://cloud.elektrikgreen.com:41010"
		if fileContent, err := json.MarshalIndent(customerMap, "", "    "); err != nil {
			log.Fatal(err)
		} else {
			if err := ioutil.WriteFile(customerMapFile, fileContent, 0644); err != nil {
				log.Fatal(err)
			}
		}
		log.Fatal(err)
	} else {
		if err := json.Unmarshal(customerMapContent, &customerMap); err != nil {
			log.Fatal(err)
		}
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		host := customerMap[strings.ToLower(strings.Split(req.Host, ".")[0])]
		if host == "" {
			AddNewOrUpdateCustomer(w, req)
		} else {
			http.Redirect(w, req, host, http.StatusPermanentRedirect)
		}
	})
	log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, nil))
}
