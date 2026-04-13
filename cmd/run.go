package cmd

import (
	"fmt"
	"os"
	"time"

	"kubeload/internal/config"
	"kubeload/internal/k8s"
	"kubeload/internal/load"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "kubeload",
	Short: "Kubernetes load generator",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run kubeload",
	Run: func(cmd *cobra.Command, args []string) {

		// create client
		client, err := k8s.NewClient()
		if err != nil {
			fmt.Println("error creating client:", err)
			os.Exit(1)
		}

		// connectivity check
		_, err = client.CoreV1().
			Pods(cfg.Namespace).
			List(cmd.Context(), metav1.ListOptions{})

		if err != nil {
			fmt.Println("cluster unreachable:", err)
			os.Exit(1)
		}

		fmt.Println("Connected to cluster")

		err = load.Run(
			client,
			cfg.Namespace,
			"kubeload", // deployment name
			cfg.Label,
			cfg.Replicas,
			2*time.Second,
		)

		if err != nil {
			fmt.Println("run failed:", err)
			os.Exit(1)
		}

		// print config
		fmt.Println("Replicas:", cfg.Replicas)
		fmt.Println("Namespace:", cfg.Namespace)
		fmt.Println("Image:", cfg.Image)
		fmt.Println("Concurrency:", cfg.Concurrency)
		fmt.Println("Timeout:", cfg.Timeout)
		fmt.Println("Label:", cfg.Label)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	runCmd.Flags().IntVar(&cfg.Replicas, "replicas", 1, "Number of pods")
	runCmd.Flags().StringVar(&cfg.Namespace, "namespace", "default", "Namespace")
	runCmd.Flags().StringVar(&cfg.Image, "image", "nginx", "Container image")
	runCmd.Flags().IntVar(&cfg.Concurrency, "concurrency", 1, "Worker count")
	runCmd.Flags().DurationVar(&cfg.Timeout, "timeout", 60*time.Second, "Timeout")
	runCmd.Flags().StringVar(&cfg.Label, "label", "app=kubeload", "Label")

	rootCmd.AddCommand(runCmd)
}
