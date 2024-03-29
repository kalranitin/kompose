/*
Copyright 2016 Skippbox, Ltd All rights reserved.

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

package app

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/Sirupsen/logrus"
    "github.com/codegangsta/cli"

    "github.com/docker/libcompose/project"

    "encoding/json"
    "io/ioutil"

    "k8s.io/kubernetes/pkg/api"
    "k8s.io/kubernetes/pkg/apis/extensions"
    "k8s.io/kubernetes/pkg/util"
    "k8s.io/kubernetes/pkg/api/unversioned"
    client "k8s.io/kubernetes/pkg/client/unversioned"

    "github.com/ghodss/yaml"
)

/* Kubernetes specific configuration */

func ProjectKuberConfig(p *project.Project, c *cli.Context) {
    url := c.String("host")

    outputFilePath := ".kuberconfig"
    wurl := []byte(url)
    if err := ioutil.WriteFile(outputFilePath, wurl, 0644); err != nil {
        logrus.Fatalf("Failed to write k8s api server address to %s: %v", outputFilePath, err)
    }
}

func ProjectKuberPS(p *project.Project, c *cli.Context) {
    server := getK8sServer("")
    version := "v1"

    client := client.NewOrDie(&client.Config{Host: server, Version: version})
	if c.BoolT("svc") {
        fmt.Printf("%-20s%-20s%-20s%-20s\n","Name", "Cluster IP", "Ports", "Selectors")
        for name := range p.Configs {
            var ports string
            var selectors string
            services, err := client.Services(api.NamespaceDefault).Get(name)

            if err != nil {
                logrus.Debugf("Cannot find service for: ", name)
            } else {

                for i := range services.Spec.Ports {
                    p := strconv.Itoa(services.Spec.Ports[i].Port)
                    ports += ports + string(services.Spec.Ports[i].Protocol) + "(" + p + "),"
                }

                for k,v := range services.ObjectMeta.Labels {
                    selectors += selectors + k + "=" + v + ","
                }

                ports = strings.TrimSuffix(ports, ",")
                selectors = strings.TrimSuffix(selectors, ",")

                fmt.Printf("%-20s%-20s%-20s%-20s\n", services.ObjectMeta.Name,
                    services.Spec.ClusterIP, ports, selectors)
            }

        }
	}

	if c.BoolT("rc") {
        fmt.Printf("%-15s%-15s%-30s%-10s%-20s\n", "Name", "Containers", "Images",
            "Replicas", "Selectors")
        for name := range p.Configs {
            var selectors string
            var containers string
            var images string
            rc, err := client.ReplicationControllers(api.NamespaceDefault).Get(name)

            /* Should grab controller, container, image, selector, replicas */

            if err != nil {
                logrus.Debugf("Cannot find rc for: ", string(name))
            } else {

                for k,v := range rc.Spec.Selector {
                    selectors += selectors + k + "=" + v + ","
                }

                for i := range rc.Spec.Template.Spec.Containers {
                    c := rc.Spec.Template.Spec.Containers[i]
                    containers += containers + c.Name + ","
                    images += images + c.Image + ","
                }
                selectors = strings.TrimSuffix(selectors, ",")
                containers = strings.TrimSuffix(containers, ",")
                images = strings.TrimSuffix(images, ",")

                fmt.Printf("%-15s%-15s%-30s%-10d%-20s\n", rc.ObjectMeta.Name, containers,
                    images, rc.Spec.Replicas, selectors)
            }
        }
	}

}

func ProjectKuberDelete(p *project.Project, c *cli.Context) {
    server := getK8sServer("")
    version := "v1"
    client := client.NewOrDie(&client.Config{Host: server, Version: version})

    for name := range p.Configs {
		if len(c.String("name")) > 0 && name != c.String("name") {
			continue
		}

        if c.BoolT("svc") {
            err := client.Services(api.NamespaceDefault).Delete(name)
            if err != nil {
                logrus.Fatalf("Unable to delete service %s: %s\n", name, err)
            }
        } else if c.BoolT("rc") {
            err := client.ReplicationControllers(api.NamespaceDefault).Delete(name)
            if err != nil {
                logrus.Fatalf("Unable to delete replication controller %s: %s\n", name, err)
            }
        }
    }
}

func ProjectKuberScale(p *project.Project, c *cli.Context) {
    server := getK8sServer("")
    version := "v1"
    client := client.NewOrDie(&client.Config{Host: server, Version: version})

    if c.Int("scale") <= 0 {
        logrus.Fatalf("Scale must be defined and a positive number")
    }

    for name := range p.Configs {
        if len(c.String("rc")) == 0 || c.String("rc") == name {
            s, err := client.ExtensionsClient.Scales(api.NamespaceDefault).Get("ReplicationController", name)
            if err != nil {
                logrus.Fatalf("Error retrieving scaling data: %s\n", err)
            }

            s.Spec.Replicas = c.Int("scale")

            s, err = client.ExtensionsClient.Scales(api.NamespaceDefault).Update("ReplicationController", s)
            if err != nil {
                logrus.Fatalf("Error updating scaling data: %s\n", err)
            }

            fmt.Printf("Scaling %s to: %d\n", name, s.Spec.Replicas)
        }
    }
}

func ProjectKuber(p *project.Project, c *cli.Context) {
    createInstance := true
    generateYaml := false
    composeFile := c.String("file")

    p = project.NewProject(&project.Context{
        ProjectName: "kube",
        ComposeFile: composeFile,
    })

    if err := p.Parse(); err != nil {
        logrus.Fatalf("Failed to parse the compose project from %s: %v", composeFile, err)
    }

    server := getK8sServer("")

    if c.BoolT("deployment") || c.BoolT("chart") {
        createInstance = false
    }

    if c.BoolT("yaml") {
      generateYaml = true
    }

    var mServices map[string]api.Service = make(map[string]api.Service)
    var serviceLinks []string

    version := "v1"
    // create new client
    client := client.NewOrDie(&client.Config{Host: server, Version: version})

    for name, service := range p.Configs {
        rc := &api.ReplicationController{
            TypeMeta: unversioned.TypeMeta{
                Kind:       "ReplicationController",
                APIVersion: "v1",
            },
            ObjectMeta: api.ObjectMeta{
                Name:   name,
                Labels: map[string]string{"service": name},
            },
            Spec: api.ReplicationControllerSpec{
                Replicas: 1,
                Selector: map[string]string{"service": name},
                Template: &api.PodTemplateSpec{
                    ObjectMeta: api.ObjectMeta{
                        Labels: map[string]string{"service": name},
                    },
                    Spec: api.PodSpec{
                        Containers: []api.Container{
                            {
                                Name:  name,
                                Image: service.Image,
                            },
                        },
                    },
                },
            },
        }
        sc := &api.Service{
            TypeMeta: unversioned.TypeMeta{
                Kind:       "Service",
                APIVersion: "v1",
            },
            ObjectMeta: api.ObjectMeta{
                Name:   name,
                Labels: map[string]string{"service": name},
            },
            Spec: api.ServiceSpec{
                Selector: map[string]string{"service": name},
            },
        }

        dc := &extensions.Deployment {
            TypeMeta: unversioned.TypeMeta {
                Kind:       "Deployment",
                APIVersion: "extensions/v1beta1",
            },
            ObjectMeta: api.ObjectMeta{
                Name: name,
                Labels: map[string]string{"service": name},
            },
            Spec: extensions.DeploymentSpec {
                Replicas: 1,
                Selector: map[string]string{"service": name},
                UniqueLabelKey: p.Name,
                Template: &api.PodTemplateSpec {
                    ObjectMeta: api.ObjectMeta {
                        Labels: map[string]string {"service": name},
                    },
                    Spec: api.PodSpec {
                        Containers: []api.Container {
                            {
                                Name:  name,
                                Image: service.Image,
                            },
                        },
                    },
                },
            },
        }

        // Configure the environment variables.
        var envs []api.EnvVar
        for _, env := range service.Environment.Slice() {
            var character string = "="
            if strings.Contains(env, character) {
                value := env[strings.Index(env, character) + 1: len(env)]
                name :=  env[0:strings.Index(env, character)]
                name = strings.TrimSpace(name)
                value = strings.TrimSpace(value)
                envs = append(envs, api.EnvVar{
                                               Name: name,
                                               Value: value,
                                              })
            } else {
                character = ":"
                if strings.Contains(env, character) {
                  var charQuote string = "'"
                  value := env[strings.Index(env, character) + 1: len(env)]
                  name := env[0:strings.Index(env, character)]
                  name = strings.TrimSpace(name)
                  value = strings.TrimSpace(value)
                  if strings.Contains(value, charQuote) {
                    value = strings.Trim(value, "'")
                  }
                  envs = append(envs, api.EnvVar{
                                                 Name: name,
                                                 Value: value,
                                                })
                } else {
                  logrus.Fatalf("Invalid container env %s for service %s", env, name)
                }
            }
        }

        rc.Spec.Template.Spec.Containers[0].Env = envs

        // Configure the container ports.
        var ports []api.ContainerPort
        for _, port := range service.Ports {
            var character string = ":"
            if strings.Contains(port, character) {
                //portNumber := port[0:strings.Index(port, character)]
                targetPortNumber := port[strings.Index(port, character) + 1: len(port)]
                targetPortNumber = strings.TrimSpace(targetPortNumber)
                targetPortNumberInt, err := strconv.Atoi(targetPortNumber)
                if err != nil {
                    logrus.Fatalf("Invalid container port %s for service %s", port, name)
                }
                ports = append(ports, api.ContainerPort{ContainerPort: targetPortNumberInt})
            } else {
                portNumber, err := strconv.Atoi(port)
                if err != nil {
                    logrus.Fatalf("Invalid container port %s for service %s", port, name)
                }
                ports = append(ports, api.ContainerPort{ContainerPort: portNumber})
            }
        }

        rc.Spec.Template.Spec.Containers[0].Ports = ports

        // Configure the service ports.
        var servicePorts []api.ServicePort
        for _, port := range service.Ports {
            var character string = ":"
            if strings.Contains(port, character) {
                portNumber := port[0:strings.Index(port, character)]
                portNumber = strings.TrimSpace(portNumber)
                targetPortNumber := port[strings.Index(port, character) + 1: len(port)]
                targetPortNumber = strings.TrimSpace(targetPortNumber)
                portNumberInt, err := strconv.Atoi(portNumber)
                if err != nil {
                    logrus.Fatalf("Invalid container port %s for service %s", port, name)
                }
                targetPortNumberInt, err1 := strconv.Atoi(targetPortNumber)
                if err1 != nil {
                    logrus.Fatalf("Invalid container port %s for service %s", port, name)
                }
                var targetPort util.IntOrString
                targetPort.StrVal = targetPortNumber
                targetPort.IntVal = targetPortNumberInt
                servicePorts = append(servicePorts, api.ServicePort{Port: portNumberInt, Name: portNumber, Protocol: "TCP", TargetPort: targetPort})
            } else {
                portNumber, err := strconv.Atoi(port)
                if err != nil {
                    logrus.Fatalf("Invalid container port %s for service %s", port, name)
                }
                var targetPort util.IntOrString
                targetPort.StrVal = strconv.Itoa(portNumber)
                targetPort.IntVal = portNumber
                servicePorts = append(servicePorts, api.ServicePort{Port: portNumber, Name: strconv.Itoa(portNumber), Protocol: "TCP", TargetPort: targetPort})
            }
        }
        sc.Spec.Ports = servicePorts

        // Configure the container restart policy.
        switch service.Restart {
        case "", "always":
            rc.Spec.Template.Spec.RestartPolicy = api.RestartPolicyAlways
        case "no":
            rc.Spec.Template.Spec.RestartPolicy = api.RestartPolicyNever
        case "on-failure":
            rc.Spec.Template.Spec.RestartPolicy = api.RestartPolicyOnFailure
        default:
            logrus.Fatalf("Unknown restart policy %s for service %s", service.Restart, name)
        }

        // Configure the container privileged mode
        if service.Privileged == true {
          securitycontexts := &api.SecurityContext{
            Privileged: &service.Privileged,
          }
          rc.Spec.Template.Spec.Containers[0].SecurityContext = securitycontexts
        }

        // convert datarc to json / yaml
        datarc, err := json.MarshalIndent(rc, "", "  ")
        if generateYaml == true {
          datarc, err = yaml.Marshal(rc)
        }

        if err != nil {
            logrus.Fatalf("Failed to marshal the replication controller: %v", err)
        }
        logrus.Debugf("%s\n", datarc)

        // convert datasc to json / yaml
        datasc, er := json.MarshalIndent(sc, "", "  ")
        if generateYaml == true {
          datasc, er = yaml.Marshal(sc)
        }

        if er != nil {
            logrus.Fatalf("Failed to marshal the service controller: %v", er)
        }

        logrus.Debugf("%s\n", datasc)

        // convert datadc to json / yaml
        datadc, err := json.MarshalIndent(dc, "", "  ")
        if generateYaml == true {
          datadc, err = yaml.Marshal(dc)
        }

        if err != nil {
            logrus.Fatalf("Failed to marshal the deployment container: %v", err)
        }

        logrus.Debugf("%s\n", datadc)

        mServices[name] = *sc

        if len(service.Links.Slice()) > 0 {
            for i := 0; i < len(service.Links.Slice()); i++ {
                var data string = service.Links.Slice()[i]
                if len(serviceLinks) == 0 {
                    serviceLinks = append(serviceLinks, data)
                } else {
                    for _, v := range serviceLinks {
                        if v != data {
                            serviceLinks = append(serviceLinks, data)
                        }
                    }
                }
            }

        }

        // call create RC api
        if createInstance == true {
            rcCreated, err := client.ReplicationControllers(api.NamespaceDefault).Create(rc)
            if err != nil {
                fmt.Println(err)
            }
            logrus.Debugf("%s\n", rcCreated)
        }

        fileRC := fmt.Sprintf("%s-rc.json", name)
        if generateYaml == true {
          fileRC = fmt.Sprintf("%s-rc.yaml", name)
        }
        if err := ioutil.WriteFile(fileRC, []byte(datarc), 0644); err != nil {
            logrus.Fatalf("Failed to write replication controller: %v", err)
        }

        /* Create the deployment container */
        if c.BoolT("deployment") {
            fileDC := fmt.Sprintf("%s-deployment.json", name)
            if generateYaml == true {
              fileDC = fmt.Sprintf("%s-deployment.yaml", name)
            }
            if err := ioutil.WriteFile(fileDC, []byte(datadc), 0644); err != nil {
                logrus.Fatalf("Failed to write deployment container: %v", err)
            }
        }

        for k, v := range mServices {
            for i :=0; i < len(serviceLinks); i++ {
                //if serviceLinks[i] == k {
                    // call create SVC api
                    if createInstance == true {
                        scCreated, err := client.Services(api.NamespaceDefault).Create(&v)
                        if err != nil {
                            fmt.Println(err)
                        }
                        logrus.Debugf("%s\n", scCreated)
                    }

                    datasvc, er := json.MarshalIndent(v, "", "  ")
                    if er != nil {
                        logrus.Fatalf("Failed to marshal the service controller: %v", er)
                    }

                    fileSVC := fmt.Sprintf("%s-svc.json", k)
                    if generateYaml == true {
                      fileSVC = fmt.Sprintf("%s-svc.yaml", k)
                    }

                    if err := ioutil.WriteFile(fileSVC, []byte(datasvc), 0644); err != nil {
                        logrus.Fatalf("Failed to write service controller: %v", err)
                    }
                //}
            }
        }
    }

    /* Need to iterate through one more time to ensure we capture all service/rc */
    for name := range p.Configs {
        if c.BoolT("chart") {
            err := generateHelm(composeFile, name)
            if err != nil {
                logrus.Fatalf("Failed to create Chart data: %s\n", err)
            }
        }
    }
}
