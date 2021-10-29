/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	pb "go-grpc/pkg/gopher"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

const (
	port		 = ":9000"
	GopherizeMeURL =  "https://gopherize.me/api/artwork/"
)

type Server struct {
	pb.UnimplementedGopherServer
}

type Gopher struct {
	URL string `json:"categories[0].images[0].href"`
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Schema gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()

		pb.RegisterGopherServer(grpcServer, &Server{})

		log.Printf("GRPC server is listening on %v", lis.Addr())

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func (s *Server) GetGopher(ctx context.Context, req *pb.GopherRequest) (*pb.GopherReply, error) {
	res := &pb.GopherReply{}

	// check request
	if req == nil {
		fmt.Println("request is nil")
		return res, xerrors.Errorf("request is nil")
	}

	if req.Name == "" {
		fmt.Println("name in request is empty")
		return res,  xerrors.Errorf("name in request is empty")
	}

	log.Printf("Received: %v\n", req.GetName())

	// call GopherizeMe API
	response, err := http.Get(GopherizeMeURL)
	if err != nil {
		log.Fatalf("failed to call GopherizeMe API: %v", err)
	}
	defer response.Body.Close()

	var p []byte

	response.Body.Read(p)

	fmt.Println(p)

	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("failer to read response body: %v", err)
		}

		var data Gopher
		err =json.Unmarshal(body, &data)
		if err != nil {
			log.Fatalf("failer to unmarshal JSON: %v", err)
		}

		var gophers strings.Builder

		gophers.WriteString(data.URL + "\n")

		res.Message = gophers.String()
	} else {
		log.Fatal("Can't get the gopher")
	}

	return res, nil
}
