package main

import (
	"context"
	flag "github.com/spf13/pflag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/delete"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	v1 "github.com/openshift/api/project/v1"
	projectv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
)

func main() {
	err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}


func start() error {
	
	
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube","config"), "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "")
	}

	//Load config for Openshift's go-client from kubeconfig file	
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}


	clientset, err := projectv1.NewForConfig(config)
	if err != nil {
		return err
	}

	pods, err := clientset.Projects().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d projects in the cluster\n", len(pods.Items))

	//Temp Project defintion 
	project := &v1.Project{

		TypeMeta: metav1.TypeMeta{
			Kind: "NameSpace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-project3",
		},


	}

	fmt.Printf("Creating Project with Openshift's go-client\n")
	_ , err = clientset.Projects().Create(context.TODO(), project, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	
	time.Sleep(10.0 * time.Second)

	pods, err = clientset.Projects().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	
	fmt.Printf("There are %d projects in the cluster\n", len(pods.Items))

	fmt.Printf("Deleting Project with Openshift's go client\n")

	err = clientset.Projects().Delete(context.TODO(), "demo-project3", metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	time.Sleep(10.0 * time.Second)

	pods, err = clientset.Projects().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d projects in the cluster\n", len(pods.Items))
	

	fmt.Println("Creating Project with Programmatic Kubectl wrapper")

	//Load Config for Kubectl Wrapper Function
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := kcmdutil.NewMatchVersionFlags(kubeConfigFlags)
	
	//Create a new Credential factory 
	f := kcmdutil.NewFactory(matchVersionKubeConfigFlags)

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stdout}

	//Make a new kubctl command 
	cmd := apply.NewCmdApply("kubectl",f,ioStreams)

	//Set relevant flags and run command to make project 
	cmd.Flags().Set("filename", "project.yaml")
	//cmd.Flags().Set("output", "json")
	//cmd.Flags().Set("dry-run", "true")
	cmd.Run(cmd, []string{})

	time.Sleep(5.0 * time.Second)

	fmt.Println("Deleting Project with Programmatic Kubectl wrapper")

	//Make a command to delete the new project 
	cmd2 := delete.NewCmdDelete(f, ioStreams)
	
	cmd2.Flags().Set("filename", "project.yaml")
	cmd2.Run(cmd, []string{})

	//file, err := os.Open(os.Stderr)
	//file2, err := os.Open(os.Stdout)
	
	return nil 
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
