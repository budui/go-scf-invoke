package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/tencentyun/scf-go-lib/events"
	"wrong.wang/x/go-scf-invoke/convert"
	"wrong.wang/x/go-scf-invoke/scfinvoke"
)

func main() {
	var port int

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "expected 'json' or 'server' subcommands")
	}

	var jsonInputPath string
	jsonCommand := flag.NewFlagSet("json", flag.ExitOnError)
	jsonCommand.StringVar(&jsonInputPath, "path", "input.json", "input json path")
	jsonCommand.IntVar(&port, "port", 8001, "rpc port")

	var ListenPort int
	var integratedResponse bool
	serverCommand := flag.NewFlagSet("server", flag.ExitOnError)
	serverCommand.IntVar(&ListenPort, "l", 8080, "listen port")
	serverCommand.BoolVar(&integratedResponse, "i", false, "integrated Response")
	serverCommand.IntVar(&port, "port", 8001, "rpc port")
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("expected 'json' or 'server' subcommands")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "json":
		jsonCommand.Parse(os.Args[2:])
		jsonMock(port, jsonInputPath)
	case "server":
		serverCommand.Parse(os.Args[2:])
		serverMock(port, fmt.Sprintf(":%d", ListenPort), integratedResponse)
	default:
		fmt.Println("expected 'json' or 'server' subcommands")
		os.Exit(1)
	}
}

func jsonMock(RPCPort int, jsonPath string) {

	var payload interface{}

	jsonByte, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(jsonByte, &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("sending %s...\n", jsonPath)
	response, err := scfinvoke.Run(scfinvoke.Input{
		Port:    RPCPort,
		Payload: payload,
	})
	if err != nil {
		fmt.Println("\nERROR:")
		fmt.Println("===============================")
		fmt.Println(err)
	}
	fmt.Println("\nRESPONSE:")
	fmt.Println("===============================")
	fmt.Println(response)
}

func serverMock(RPCPort int, ListenAddr string, integratedResponse bool) {
	fmt.Printf("listen %v...\n", ListenAddr)
	fmt.Printf("integratedResponse %v\n", integratedResponse)
	simpleConvertHandler := func(w http.ResponseWriter, r *http.Request) {
		response, err := scfinvoke.Run(scfinvoke.Input{
			Port:    RPCPort,
			Payload: convert.NewAPIGatewayRequestFromRequest(r),
		})
		if err != nil {
			fmt.Println("\nRPC ERROR:")
			fmt.Println("===============================")
			fmt.Println(err)
		}
		if !integratedResponse {
			_, err = w.Write(response)
			if err != nil {
				panic(err)
			}
		} else {
			var apigwrp events.APIGatewayResponse

			if err := json.Unmarshal(response, &apigwrp); err != nil {
				panic(err)
			}

			for k, v := range apigwrp.Headers {
				w.Header().Set(k, v)
			}

			w.WriteHeader(apigwrp.StatusCode)
			if apigwrp.IsBase64Encoded {
				bodyByte, err := base64.StdEncoding.DecodeString(apigwrp.Body)
				if err != nil {
					panic(err)
				}
				_, err = w.Write(bodyByte)
				if err != nil {
					panic(err)
				}
			} else {
				io.WriteString(w, apigwrp.Body)
			}
		}
	}
	http.HandleFunc("/", simpleConvertHandler)
	log.Fatal(http.ListenAndServe(ListenAddr, nil))
}
