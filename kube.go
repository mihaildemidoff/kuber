package main

import (
	"flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
	"os"
	"net/http"
	"crypto/tls"
	"log"
	"path/filepath"
)

const authorizationHeader = "Authorization"
const bearerPrefix = "Bearer "
const defaultHttpMethod = "GET"
const linuxHomeEnvVariable = "HOME"
const windowsHomeEnvVariable = "USERPROFILE"

type Flags struct {
	bootstrapPath  *string
	kubeconfigPath *string
}

func main() {
	flags := new(Flags)
	flags.initFlags()
	// Load application settings
	settings, err := LoadBootstrapSettings(*flags.bootstrapPath)
	checkFatalError(err, "Couldn't load bootstrap settings: ")

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *flags.kubeconfigPath)
	checkFatalError(err, "Couldn't create config: ")

	// establish connection to kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	checkFatalError(err, "Couldn't create kubernetes client: ")

	InitStateFromBootstrapSettings(settings, clientset)
	verify(settings, clientset, config.Host)
	CleanUp(settings, clientset)
}

func verify(settings *BootstrapSettings, clientset *kubernetes.Clientset, kubernetesHost string) {
	// disable certificate verification 
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for _, user := range settings.Users {
		token, err := findUserBearerToken(clientset, user)
		if err != nil {
			log.Printf("Error occured during token fetching for user '%s' in namespace '%s'. Error: '%s'",
				user.Name, user.Namespace, err.Error())
			continue
		} else if token == nil {
			log.Printf("service-account-token not found for user '%s' in namespace '%s'",
				user.Name, user.Namespace)
			continue
		}
		doUserTests(user, kubernetesHost, string(token.Data["token"]))
	}
}

// execute user tests
func doUserTests(user UsersSettings, kubernetesHost string, token string) {
	for _, test := range user.Tests {

		req, _ := http.NewRequest(defaultHttpMethod, kubernetesHost+test.Path, nil)

		req.Header.Add(authorizationHeader, bearerPrefix+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error occured during fetching resource %s. Error: %s", test.Path, err.Error())
			continue
		}
		log.Printf("%v. Expected code %v, got %v. User: %v. Namespace: %v. Path: %v",
			resp.StatusCode == test.ExpectedCode, test.ExpectedCode, resp.StatusCode, user.Name,
			user.Namespace,
			test.Path)
		resp.Body.Close()
	}
}

// Find service-account-token token created by default
func findUserBearerToken(clientset *kubernetes.Clientset, user UsersSettings) (*v1.Secret, error) {
	sa, err := clientset.CoreV1().ServiceAccounts(user.Namespace).Get(user.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, secretRef := range sa.Secrets {
		secret, err := clientset.CoreV1().Secrets(user.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err == nil && secret.Type == v1.SecretTypeServiceAccountToken {
			return secret, nil
		}
	}
	return nil, nil
}

// setup flags
func (v *Flags) initFlags() {
	v.bootstrapPath = flag.String("bootstrap",
		"bootstrap.json", "(optional) path to the bootstrap file")
	if home := homeDir(); home != "" {
		v.kubeconfigPath = flag.String("kubeconfig",
			filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		v.kubeconfigPath = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
}

func homeDir() string {
	if h := os.Getenv(linuxHomeEnvVariable); h != "" {
		return h
	}
	return os.Getenv(windowsHomeEnvVariable)
}

func checkFatalError(err error, message string) {
	if err != nil {
		panic(message + err.Error())
	}
}
