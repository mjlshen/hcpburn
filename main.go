package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"path/filepath"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const namespace = "open-cluster-management-scale"

func main() {
	// Make the API Server spam etcd
	// use the current context in kubeconfig
	// creates the in-cluster config
	var (
		config *rest.Config
		err    error
		wg     sync.WaitGroup
	)

	config, err = rest.InClusterConfig()
	if err != nil {
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	config.QPS = 100
	config.Burst = 200
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	log.Println("Starting up")

	wg.Add(8)
	go CreateConfigMaps(&wg, clientset)
	go CreateConfigMaps(&wg, clientset)
	go CreateConfigMaps(&wg, clientset)
	go CreateConfigMaps(&wg, clientset)
	go CreateConfigMaps(&wg, clientset)
	go CreateConfigMaps(&wg, clientset)
	go CreateConfigMaps(&wg, clientset)
	go DeleteConfigMap(&wg, clientset)
	wg.Wait()
}

func CreateConfigMaps(wg *sync.WaitGroup, clientset *kubernetes.Clientset) error {
	defer wg.Done()

	data := RandStringRunes(1024)

	for {
		randomName := RandStringRunes(10)

		// Create a Configmap
		if _, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomName,
				Namespace: namespace,
				Labels: map[string]string{
					"purpose": "watch-the-world-burn",
				},
			},
			Data: map[string]string{
				"data": data,
			},
		}, metav1.CreateOptions{}); err != nil {
			panic(err)
		}

		log.Printf("created Configmap: %s", randomName)
	}
}

func DeleteConfigMap(wg *sync.WaitGroup, clientset *kubernetes.Clientset) error {
	defer wg.Done()

	for {
		// Create a Configmap
		configmaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "purpose=watch-the-world-burn",
		})
		if err != nil {
			panic(err)
		}

		if len(configmaps.Items) > 0 {
			if err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), configmaps.Items[0].Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
			log.Printf("deleted Configmap: %s", configmaps.Items[0].Name)
		}
	}
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
