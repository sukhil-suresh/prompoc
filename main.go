package main

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	endpointCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "poc",
			Name:      "endpoint_counter",
			Help:      "Number of times endpoint was accessed",
		},
	)

	podsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "poc",
			Name:      "pods_gauge",
			Help:      "Number of pods in the cluster",
		},
	)
)

const counterFilepath = "/usr/prom/counter"
const port = ":8080"

func main() {
	cs, err := getK8sClientSet()
	if err != nil {
		log.Fatalf("error getting kubernetes clientset: %s", err.Error())
	}

	fbi, err := newFileBackedInt(counterFilepath)
	if err != nil {
		log.Fatalf("error creating FileBackedInt: %s", err.Error())
	}

	prometheus.MustRegister(endpointCounter, podsGauge)

	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err = updateEndpointCounter(fbi); err != nil {
			log.Fatalf("error updating endpoint counter: %s", err.Error())
		}
		if err = updatePodsGauge(cs); err != nil {
			log.Fatalf("error updating pods gauge: %s", err.Error())
		}
		promhttp.Handler().ServeHTTP(w, r)
	}))

	log.Fatal(http.ListenAndServe(port, nil))
}

func getK8sClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func updatePodsGauge(cs *kubernetes.Clientset) error {
	pods, err := cs.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	countPods := len(pods.Items)
	podsGauge.Set(float64(countPods))
	return nil
}

func updateEndpointCounter(fbi fileBackedInt) error {
	count, err := fbi.read()
	if err != nil {
		return err
	}

	endpointCounter.Inc()
	return fbi.write(count)
}
