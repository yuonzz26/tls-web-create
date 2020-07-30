package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
)

func hello(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "tls.html")
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		company := r.FormValue("company")
		host := r.FormValue("host")
		fmt.Println("company", company)
		fmt.Println("host", host)

		flag.Parse()

		certName := company
		hostName := host

		if certName == "" || hostName == "" {
			usageAndExit("You must supply both a -cn (certificate name) and -h (host name) parameter")
		}

		createPrivateCA(certName)
		createServerCertKey(hostName)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func createPrivateCA(certificateAuthorityName string) {
	_, err := callCommand("openssl", "genrsa", "-out", "myCA.key", "2048")
	if err != nil {
		log.Fatal("Could not create private Certificate Authority key")
	}

	_, err = callCommand("openssl", "req", "-x509", "-new", "-key", "myCA.key", "-out", "myCA.cer", "-days", "730", "-subj", "/CN=\""+certificateAuthorityName+"\"")
	if err != nil {
		log.Fatal("Could not create private Certificate Authority certificate")
	}
}

func createServerCertKey(host string) {
	_, err := callCommand("openssl", "genrsa", "-out", "mycert1.key", "2048")
	if err != nil {
		log.Fatal("Could not create private server key")
	}

	_, err = callCommand("openssl", "req", "-new", "-out", "mycert1.req", "-key", "mycert1.key", "-subj", "/CN="+host)
	if err != nil {
		log.Fatal("Could not create private server certificate signing request")
	}

	_, err = callCommand("openssl", "x509", "-req", "-in", "mycert1.req", "-out", "mycert1.cer", "-CAkey", "myCA.key", "-CA", "myCA.cer", "-days", "365", "-CAcreateserial", "-CAserial", "serial")
	if err != nil {
		log.Fatal("Could not create private server certificate")
	}

}

func callCommand(command string, arg ...string) (string, error) {
	out, err := exec.Command(command, arg...).Output()

	if err != nil {
		log.Println("callCommand failed!")
		log.Println("")
		log.Println(string(debug.Stack()))
		return "", err
	}
	return string(out), nil
}

func usageAndExit(message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func main() {
	http.HandleFunc("/", hello)

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
