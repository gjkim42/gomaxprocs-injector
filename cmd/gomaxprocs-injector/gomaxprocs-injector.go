package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gjkim42/gomaxprocs-injector/pkg/admission"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func main() {
	checkErr(os.Stderr, NewDefaultGOMAXPROCSInjectorCommand().Execute())
}

func checkErr(w io.Writer, err error) {
	if err != nil {
		fmt.Fprintln(w, err)
		os.Exit(1)
	}
}

func NewDefaultGOMAXPROCSInjectorCommand() *cobra.Command {
	options := &GOMAXPROCSInjectorOptions{}
	certFile := "tls.crt"
	keyFile := "tls.key"
	bindAddress := "0.0.0.0"
	port := 443
	cmd := &cobra.Command{
		Use:   "gomaxprocs-injector",
		Short: "The admission controller that injects optimized GOMAXPROCS environment variable into pods",
		Run: func(cmd *cobra.Command, args []string) {
			klog.InfoS("Starting...")
			checkErr(os.Stderr, options.Complete(certFile, keyFile, bindAddress, port))
			checkErr(os.Stderr, options.Run())
		},
	}

	klog.InitFlags(nil)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.Flags().StringVar(&certFile, "cert-file", certFile, "File containing the default Certificate for HTTPS.")
	cmd.Flags().StringVar(&keyFile, "key-file", keyFile, "File containing the default Key for HTTPS.")
	cmd.Flags().StringVar(&bindAddress, "bind-address", bindAddress, "The address on which to listen for the webhook's server")
	cmd.Flags().IntVar(&port, "port", port, "The port on which to serve the webhook's server")

	return cmd
}

type GOMAXPROCSInjectorOptions struct {
	Address   string
	TLSConfig *tls.Config
}

func (o *GOMAXPROCSInjectorOptions) Complete(certFile, keyFile, bindAddress string, port int) error {
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
		o.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	o.Address = fmt.Sprintf("%s:%d", bindAddress, port)

	return nil
}

func (o *GOMAXPROCSInjectorOptions) Run() error {
	http.Handle("/webhook", admission.NewController())
	http.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	server := &http.Server{
		Addr:      o.Address,
		TLSConfig: o.TLSConfig,
	}

	return server.ListenAndServeTLS("", "")
}
