package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//const certFile = "/etc/cert/localhost.crt"
//const keyFile = "/etc/cert/localhost.key"
var (
	certFile string
	keyFile  = "/etc/letsencrypt/live/cloud.elektrikgreen.com/privkey.pem"
)

const customerMapFile = "/etc/customerMap.json"

const infoHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="https://logo.cloud.elektrikgreen.com" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<h2>Please use your customer specific URL to connect to the ElektrikGreen cloud service</h2>
		</div>
	</body>
</html>`

const portErrorHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="https://logo.cloud.elektrikgreen.com" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<h2>Invalid port number specified (40000 - 49999)</h2>
			<p>?subdomain=port</p>
		</div>
	</body>
</html>`

const portInUseHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="https://logo.cloud.elektrikgreen.com" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<h2>Port number is already in use by %s</h2>
			<p>?subdomain=port</p>
		</div>
	</body>
</html>`

const updatedHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="https://logo.cloud.elektrikgreen.com" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<ol>%s
			</ol>
		</div>
	</body>
</html>`

const helpHTML = `<html>
	<head>
		<meta charset="UTF-8">
		<title>ElektrikGreen Cloud</title>
	</head>
	<body>
        <div style="background-color: black; color: lightcyan; padding: 5px;">
            <h1>ElektrikGreen Cloud Services<img style="float: right;" src="https://logo.cloud.elektrikgreen.com" alt="ElektrikGreen Logo"/></h1>
        </div>
		<div>
			<ul>
				<li>https://help.cloud.elektrikgreen.com : Print this page</li>
				<li>https://list.cloud.elektrikgreen.com : List the current redirection settings</li>
				<li>https://update.cloud.elektrikgreen.com?name:port : Adds a new mapping</li>
			</ul>
		</div>
	</body>
</html>`

var customerMap map[string]string

func printCustomerMap(w http.ResponseWriter) {
	list := ""
	for id, url := range customerMap {
		list = list + "<li>" + id + " ==> " + url + "</li>"
	}
	if _, err := fmt.Fprintf(w, updatedHTML, list); err != nil {
		log.Println(err)
	}
}

func AddNewOrUpdateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		if _, err := fmt.Fprint(w, infoHTML); err != nil {
			log.Println(err)
			return
		}
	} else {
		// .../?name1=41070&name2=41080	This will add two subdomains.
		if strings.HasPrefix(r.RequestURI, "/?") {
			qry := r.RequestURI[2:]
			for _, entry := range strings.Split(qry, "&") {
				customer := strings.Split(entry, "=")
				// There should be exactly two entries separated by the = sign
				if len(customer) == 2 {
					// Check that the port is numeric
					if port, err := strconv.ParseInt(customer[1], 10, 32); err != nil {
						if _, err := fmt.Fprint(w, portErrorHTML); err != nil {
							log.Println(err)
						}
						return
					} else {
						// If numeric check its range is between 40000 and 49999
						if port < 40000 || port > 49999 {
							if _, err := fmt.Fprint(w, portErrorHTML); err != nil {
								log.Println(err)
							}
							return
						}
						// Make sure it isn't in use already
						for key, val := range customerMap {
							if val == customer[1] && key != customer[0] {
								if _, err := fmt.Fprintf(w, portInUseHTML, key); err != nil {
									log.Println(err)
								}
								return
							}
						}
						// Set or update the subdomain in the map
						customerMap[strings.ToLower(customer[0])] = customer[1]
					}
				}
			}
			// Convert the map to JSON
			if fileContent, err := json.MarshalIndent(customerMap, "", "    "); err != nil {
				log.Println(err)
				return
			} else {
				// Save the map to the file system
				if err := ioutil.WriteFile(customerMapFile, fileContent, 0644); err != nil {
					log.Println(err)
					return
				}
			}
			printCustomerMap(w)
		} else {
			if _, err := fmt.Fprint(w, infoHTML); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func printHelp(w http.ResponseWriter) {
	if _, err := fmt.Fprint(w, helpHTML); err != nil {
		log.Println(err)
	}
}

func getLogo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/images/logo.png")
}

func init() {
	flag.StringVar(&keyFile, "keyFile", "/etc/letsencrypt/live/cloud.elektrikgreen.com/privkey.pem", "Path to the key file")
	flag.StringVar(&certFile, "certFile", "/etc/letsencrypt/live/cloud.elektrikgreen.com/fullchain.pem", "Name of the certificate full chain file")
	flag.Parse()
}

func main() {
	customerMap = make(map[string]string, 0)

	if customerMapContent, err := os.ReadFile(customerMapFile); err != nil {
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
		hostName := strings.ToLower(strings.Split(req.Host, ".")[0])
		switch hostName {
		case "update":
			AddNewOrUpdateCustomer(w, req)
			break
		case "cloud":
			log.Println("cloud")
			if _, err := fmt.Fprint(w, infoHTML); err != nil {
				log.Fatal(err)
			}
			break
		case "list":
			printCustomerMap(w)
			break
		case "help":
			printHelp(w)
			break
		case "logo":
			getLogo(w, req)
		default:
			host := customerMap[hostName]
			if host == "" {
				if _, err := fmt.Fprint(w, infoHTML); err != nil {
					log.Println(err)
				}
			} else {
				hostURL := "https://" + strings.Join(strings.Split(req.Host, ".")[1:], ".") + ":" + host
				http.Redirect(w, req, hostURL, http.StatusPermanentRedirect)
			}
		}
	})
	log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, nil))
}
