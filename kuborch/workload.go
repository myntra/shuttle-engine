package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"

	b64 "encoding/base64"

	r "gopkg.in/gorethink/gorethink.v4"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/apis/core"
)

func executeWorkload(w http.ResponseWriter, req *http.Request) {
	step := types.Step{}
	helpers.PanicOnErrorAPI(helpers.ParseRequest(req, &step), w)
	// Fetch yaml from predefined_steps table
	cursor, err := r.Table("predefined_steps").Filter(map[string]interface{}{
		"name": step.StepTemplate,
	}).Run(config.RethinkSession)
	helpers.PanicOnErrorAPI(err, w)
	defer cursor.Close()
	var yamlFromDB types.YAMLFromDB
	err = cursor.One(&yamlFromDB)
	helpers.PanicOnErrorAPI(err, w)

	workloadPath := "./yaml/" + step.UniqueKey + ".yaml"
	fileContentInBytes := replaceVariables(yamlFromDB, step, workloadPath)
	helpers.PanicOnErrorAPI(err, w)

	err = ioutil.WriteFile(workloadPath, fileContentInBytes, 0777)

	helpers.PanicOnErrorAPI(err, w)

	var kubeConfigPath string
	if step.KubeConfig != "" {
		kubeConfigPath, err = setUpKubeConfig(step.KubeConfig, fmt.Sprintf("%s-%s", step.K8SCluster, step.UniqueKey))
		log.Printf("Kubeconfig:%s", err)
		if err != nil {
			helpers.PrintErr(err)
		}
	} else {
		if ClientConfigMap[step.K8SCluster].Clientset == nil {
			helpers.PanicOnErrorAPI(fmt.Errorf("kuborch does not have the requested cluster configured - %s", step.K8SCluster), w)
		}
		kubeConfigPath = ClientConfigMap[step.K8SCluster].ConfigPath
	}

	if step.ChartURL != "" {
		go runHelm(kubeConfigPath, workloadPath, step)
	} else if step.IsCommand {
		var restClient *rest.Config
		var k8sclient *kubernetes.Clientset
		if step.KubeConfig != "" {
			restClient, k8sclient, err = CreateRestK8Client(kubeConfigPath)
			if err != nil {
				helpers.PanicOnErrorAPI(fmt.Errorf("Error in creating the k8s client with kubeconfig - %s", step.K8SCluster), w)
				return
			}
		} else {
			restClient = ClientConfigMap[step.K8SCluster].RestConfig
			k8sclient = ClientConfigMap[step.K8SCluster].Clientset
		}
		go runCommand(restClient, k8sclient, kubeConfigPath, workloadPath, step)
	} else {
		var k8sclient *kubernetes.Clientset
		if step.KubeConfig != "" {
			k8sclient, err = CreateK8sClient(kubeConfigPath)
			helpers.PanicOnErrorAPI(fmt.Errorf("Error in creating the k8s client with kubeconfig - %s", step.K8SCluster), w)
		} else {
			k8sclient = ClientConfigMap[step.K8SCluster].Clientset
		}
		go runKubeCTL(kubeConfigPath, step.K8SCluster, step.UniqueKey, workloadPath, k8sclient)
	}

	eRes := helpers.Response{
		State: "Workload triggered",
		Code:  200,
	}
	inBytes, _ := json.Marshal(eRes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

func replaceVariables(yamlFromDB types.YAMLFromDB, step types.Step, workloadPath string) []byte {
	// // Some replaces happen here
	fmt.Println(step.Replacers)
	for singleReplacer, value := range step.Replacers {
		yamlFromDB.Config = strings.Replace(yamlFromDB.Config, "{{."+singleReplacer+"}}", value, -1)
	}
	// Twice for overlapping substitution
	for singleReplacer, value := range step.Replacers {
		yamlFromDB.Config = strings.Replace(yamlFromDB.Config, "{{."+singleReplacer+"}}", value, -1)
	}
	fmt.Println(yamlFromDB.Config)
	return []byte(yamlFromDB.Config)
}

func runKubeCTL(kubeConfigPath, k8scluster, uniqueKey, workloadPath string, k8sclient *kubernetes.Clientset) {
	resChan := make(chan types.WorkloadResult)
	go func(uniqueKey string) {
		for {
			log.Println("Waiting on result channel for " + uniqueKey)
			select {
			case wr := <-resChan:
				log.Println("Sending floworch result for " + uniqueKey)
				_, err := helpers.Post("http://localhost:5500/callback", wr, nil)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}(uniqueKey)
	defer close(resChan)

	cmd := exec.Command("kubectl", "--kubeconfig", kubeConfigPath, "apply", "-f", workloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		resChan <- types.WorkloadResult{
			UniqueKey: uniqueKey,
			Result:    types.FAILED,
			Details:   err.Error(),
		}
		return
	}
	log.Println("yaml executed")
	// Poll Kube API for result of workload
	log.Println("job-name=" + uniqueKey)
	log.Println("listopts done")
	time.Sleep(time.Duration(2 * time.Second))

	k8File, err := os.Open(workloadPath)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(k8File)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 2048)
	workloadTrackMap := make(map[string]int)

	watchChannel := make(chan types.WorkloadResult)
	defer close(watchChannel)
	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
		}

		log.Println("-----------------------------")
		log.Println("raw: ", string(ext.Raw))

		versions := &runtime.VersionedObjects{}
		obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, versions)
		if err != nil {
			log.Fatal(err)
		}

		// GroupKind type
		workloadKind := gvk.GroupKind().Kind
		// non existent value is default which is 0
		workloadTrackMap[workloadKind]++

		objectKindI := obj.GetObjectKind()
		structuredObj := objectKindI.(*unstructured.Unstructured)
		labelSet := structuredObj.GetLabels()
		namespace := structuredObj.GetNamespace()
		labelSelector := labels.Set(labelSet).String()
		listOpts := metav1.ListOptions{
			LabelSelector: labelSelector,
		}

		log.Println("Workload Kind : ", workloadKind)

		if namespace == "" {
			namespace = "default"
		}

		log.Println("Namespace : ", namespace)
		switch workloadKind {
		case "Job":
			listOpts = metav1.ListOptions{
				FieldSelector: fields.OneTermEqualSelector(core.ObjectNameField, structuredObj.GetName()).String(),
			}
			go JobWatch(k8sclient, watchChannel, namespace, listOpts)
		case "StatefulSet":
			go StatefulSetWatch(k8sclient, watchChannel, namespace, listOpts)
		case "Service":
			go ServiceWatch(k8sclient, watchChannel, namespace, listOpts)
		case "Deployment":
			go DeploymentWatch(k8sclient, watchChannel, namespace, listOpts)
		default:
			log.Println("Unknown workload. Completed")
		}
	}

	totalWorkload := len(workloadTrackMap)
	receivedResults := 0
	log.Println("Starting wait Loop ... ")

	/**
	 * result check receieved from go routines
	 */
	for {
		select {
		case event := <-watchChannel:
			log.Println("++++++++++++++++++++++++++Recieved Watch Event++++++++++++++++++++++++++++++")
			log.Println(event)
			receivedResults++
			if event.Result == types.FAILED {
				fmt.Println(event.Details)
				event.UniqueKey = uniqueKey
				resChan <- event
				return
			}
			if workloadTrackMap[event.Kind] == 1 {
				delete(workloadTrackMap, event.Kind)
			} else {
				workloadTrackMap[event.Kind]--
			}

			if receivedResults == totalWorkload && len(workloadTrackMap) == 0 {
				log.Println("Succesfully completed workload", uniqueKey)
				// hack for unique key specific stuff of floworch
				event.UniqueKey = uniqueKey
				resChan <- event
				return
			}
		}
	}

}

func setUpKubeConfig(kubeconfig string, filename string) (string, error) {
	decodedKubeConfig, err := b64.StdEncoding.DecodeString(kubeconfig)
	if err != nil {
		return "", err
	}
	filePath, err := createConfigFile(string([]byte(decodedKubeConfig)), filename)
	if err != nil {
		return "", err
	}
	return filePath, err
}

func removeKubeConfig(kubeconfigpath string) error {
	err := os.Remove(kubeconfigpath)
	if err != nil {
		return err
	}
	return nil
}

func createConfigFile(kubeconfig string, chartname string) (string, error) {
	path := filepath.Join(homedir.HomeDir(), ".kube", chartname)
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		file, err := os.Create(path)
		log.Println("Create File")
		if err != nil {
			return "", err
		}
		log.Println("File Created")
		_, err = file.WriteString(kubeconfig)
		if err != nil {
			return "", err
		}
		log.Println("Written ")
		defer file.Close()
	}
	return path, nil

}

// go runHelm(kubeConfigPath, step.ChartURL, workloadPath, step.ReleaseName, step.UniqueKey)
// func runHelm(kubeConfigPath, chartURL, workloadPath, releaseName, uniqueKey string) error {
func runHelm(kubeConfigPath, workloadPath string, step types.Step) error {
	resChan := make(chan types.WorkloadResult)
	go func(uniqueKey string) {
		for {
			log.Println("Waiting on result channel for " + uniqueKey)
			select {
			case wr := <-resChan:
				log.Println("Sending floworch result for " + uniqueKey)
				_, err := helpers.Post("http://localhost:5500/callback", wr, nil)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}(step.UniqueKey)
	defer close(resChan)
	defer removeKubeConfig(kubeConfigPath)
	//var installOrUpgrade string

	cmd := exec.Command("helm", "--kubeconfig", kubeConfigPath, "list", "--filter", step.ReleaseName, "--pending", "-n", step.Namespace, "-q")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	cmd.Wait()

	if err != nil {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.FAILED,
			Details:   fmt.Sprintf("Failed to get pending release, Error:%s", err),
		}
		return nil
	}
	//failing the Run if pending releases exist
	if fmt.Sprintf("%s", cmd.Stdout) != "" {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.FAILED,
			Details:   fmt.Sprintf("Pending Releases Exist\nOutput:%s", cmd.Stdout),
		}
		return nil
	}

	cmd = exec.Command("helm", "--kubeconfig", kubeConfigPath, "-n", step.Namespace, "upgrade", "--install", step.ReleaseName, "-f", workloadPath, step.ChartURL, "--atomic", "--wait", "--timeout", step.Timeout)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	cmd.Wait()
	if err != nil {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.FAILED,
			Details:   fmt.Sprintf("%s", cmd.Stderr),
		}
	} else {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.SUCCEEDED,
			Details:   fmt.Sprintf("%s", cmd.Stdout),
		}
	}

	return nil
}

func runCommand(restClient *rest.Config, clientset *kubernetes.Clientset, kubeConfigPath, workloadPath string, step types.Step) {
	log.Println(step.UniqueKey)
	resChan := make(chan types.WorkloadResult)
	go func(uniqueKey string) {
		for {
			log.Println("Waiting on result channel for " + uniqueKey)
			select {
			case wr := <-resChan:
				log.Println("Sending floworch result for " + uniqueKey)
				_, err := helpers.Post("http://localhost:5500/callback", wr, nil)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}(step.UniqueKey)
	defer close(resChan)
	defer removeKubeConfig(kubeConfigPath)

	fileByte, err := ioutil.ReadFile(workloadPath)
	if err != nil {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.FAILED,
			Details:   fmt.Sprintf("Failed to read command, Error:%s\n", err),
		}
		return
	}
	command := string(fileByte)

	// Get namespace, pod and container info from Meta
	metaMap := make(map[string]string)

	for _, metaObject := range step.Meta {
		metaMap[metaObject.Name] = fmt.Sprintf("%v", metaObject.Value)
	}

	fmt.Println("Executing", command)

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(metaMap["pod"]).
		Namespace(metaMap["namespace"]).
		SubResource("exec")
	option := &v1.PodExecOptions{
		Command:   []string{"bash", "-c", command},
		Container: metaMap["container"],
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	cmd, err := remotecommand.NewSPDYExecutor(restClient, "POST", req.URL())
	if err != nil {
		log.Println("Failed in command execution")
		log.Println(err)
		if err != nil {
			resChan <- types.WorkloadResult{
				UniqueKey: step.UniqueKey,
				Result:    types.FAILED,
				Details:   fmt.Sprintf("Failed to read command, Error:%s\n", err),
			}
		}

		return
	}

	var stdout, stderr bytes.Buffer

	err = cmd.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	if err != nil {
		log.Println(err)
		log.Println(stderr.Bytes())
	} else {
		log.Println(stdout.Bytes())
	}

	if err != nil {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.FAILED,
			Details:   fmt.Sprintf("%s", err),
		}
	} else {
		resChan <- types.WorkloadResult{
			UniqueKey: step.UniqueKey,
			Result:    types.SUCCEEDED,
			Details:   fmt.Sprintf("%s", "success"),
		}
	}
}
